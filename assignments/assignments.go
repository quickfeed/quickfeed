package assignments

import (
	"context"
	"fmt"
	"time"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

// MaxWait is the maximum time allowed for updating a course's assignments
// and docker image before aborting. It is generous mainly to accommodate the
// Docker image build, which dominates the duration of a full update; the two
// clones it also covers are comparatively quick.
const MaxWait = 15 * time.Minute

// UpdateFromCourseRepositories updates the database record for the course
// assignments and reports the problems found in the course's repositories.
//
// This is called in response to a push event to either the 'tests' or the
// 'assignments' repository, and may also be called manually by a teacher from
// the frontend. It performs the same full update whichever repository was
// pushed: db.UpdateAssignments is idempotent, and the Docker image is rebuilt
// only when the Dockerfile's digest actually changed, which it will not have
// on an assignments push. An assignments push therefore costs one extra clone
// and a no-op database pass, and in return the update becomes self-healing: a
// tests push whose webhook was missed or failed is reconciled by the next
// assignments push.
//
// Note that calling this function concurrently is safe, but it may block the
// caller for an extended period, since it involves cloning both course
// repositories, scanning them for assignments, building the Docker image and
// updating the database.
//
// Both repositories are cloned first, so that the checks below work from a
// snapshot of the two taken at the same time. The tests repository is
// required: failing to clone or read it aborts the update, since neither the
// content check nor the database update is possible without it. The
// assignments repository is only needed for the cross-repository comparison,
// so failing to clone it, or to compare the two, is logged and skips that
// comparison alone, leaving the database up to date.
//
// The returned count reports content problems in the tests repository and
// alignment problems between the two repositories. Issue details are logged
// here so that webhook updates reach the course log; the ctx is expected to
// carry a logger scoped with the course.
// Callers are responsible for logging returned errors.
func UpdateFromCourseRepositories(ctx context.Context, runner ci.Runner, db database.Database, sc scm.SCM, course *qf.Course) (int, error) {
	unlock := course.Lock()
	defer unlock()

	logger := qlog.FromContext(ctx)
	logger.Debug("updating from course repositories")
	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	clonedTestsRepo, clonedAssignmentsRepo, err := cloneCourseRepositories(ctx, sc, course)
	if err != nil {
		return 0, err
	}

	// walk the cloned tests repository and extract the assignments and the course's Dockerfile
	assignments, buildContext, issues, err := readTestsRepositoryContent(clonedTestsRepo, course.GetID())
	if err != nil {
		return 0, fmt.Errorf("reading tests repository content: %w", err)
	}
	issueCount := len(issues)
	logRepositoryIssues(ctx, "tests repository issue", issues)

	if course.UpdateDockerfile(buildContext[ci.Dockerfile]) {
		// Rebuild the Docker image for the course tagged with the course code
		if err = buildDockerImage(ctx, runner, course, buildContext); err != nil {
			return issueCount, fmt.Errorf("building course image: %w", err)
		}
		// Update the course's DockerfileDigest in the database
		if err := db.UpdateCourse(course); err != nil {
			return issueCount, fmt.Errorf("storing course Dockerfile digest: %w", err)
		}
	}

	if err = db.UpdateAssignments(assignments); err != nil {
		for _, assignment := range assignments {
			logger.Debug("assignment not updated in database", label.Assignment, assignment.GetName())
		}
		return issueCount, fmt.Errorf("updating assignments in database: %w", err)
	}
	logger.Debug("assignments successfully updated from tests repository")

	if clonedAssignmentsRepo == "" {
		// The assignments repository could not be cloned, and
		// cloneCourseRepositories has logged why; the update itself is complete.
		return issueCount, nil
	}
	alignmentIssues, err := courseRepositoryIssues(clonedTestsRepo, clonedAssignmentsRepo, assignments)
	if err != nil {
		logger.Error("failed to compare course repositories", label.Error, err)
		return issueCount, nil
	}
	logRepositoryIssues(ctx, "course repository issue", alignmentIssues)
	return issueCount + len(alignmentIssues), nil
}

// cloneCourseRepositories clones the course's tests and assignments
// repositories into the course's clone directory, refreshing any local
// copies are already there.
//
// The tests repository is required. Failure to clone it returns an error
// and the assignments repository is not cloned. The assignments repository
// is only needed for the comparing the tests and assignments repositories.
// Failure to clone the assignments repository is logged and reported as an
// empty assignmentsDir.
func cloneCourseRepositories(ctx context.Context, sc scm.SCM, course *qf.Course) (testsDir, assignmentsDir string, err error) {
	logger := qlog.FromContext(ctx)
	testsDir, err = sc.Clone(ctx, &scm.CloneOptions{
		Organization: course.GetScmOrganizationName(),
		Repository:   qf.TestsRepo,
		DestDir:      course.CloneDir(),
	})
	if err != nil {
		return "", "", fmt.Errorf("cloning tests repository: %w", err)
	}
	logger.Debug("cloned tests repository", label.Path, testsDir)

	assignmentsDir, err = sc.Clone(ctx, &scm.CloneOptions{
		Organization: course.GetScmOrganizationName(),
		Repository:   qf.AssignmentsRepo,
		DestDir:      course.CloneDir(),
	})
	if err != nil {
		logger.Error("failed to clone assignments repository", label.Error, err)
		return testsDir, "", nil
	}
	logger.Debug("cloned assignments repository", label.Path, assignmentsDir)
	return testsDir, assignmentsDir, nil
}

// buildDockerImage builds the Docker image for the given course.
func buildDockerImage(ctx context.Context, runner ci.Runner, course *qf.Course, buildContext map[string]string) error {
	logger := qlog.FromContext(ctx)
	logger.Debug("building course Dockerfile", "dockerfile", course.GetDockerfile())
	out, err := runner.Run(ctx, &ci.Job{
		Name:         course.JobName(),
		Image:        course.DockerImage(),
		BuildContext: buildContext,
		Commands:     []string{`echo -n "Hello from Dockerfile"`},
	})
	logger.Debug("course image build completed", "output", out)
	if err != nil {
		return fmt.Errorf("building image from %s's Dockerfile: %w", course.GetCode(), err)
	}
	return nil
}
