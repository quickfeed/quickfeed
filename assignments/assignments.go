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

// UpdateResult reports the outcome of UpdateFromCourseRepositories. Its fields
// are meaningful only when the update returned no error.
type UpdateResult struct {
	// IssueCount is the number of content problems found in the tests
	// repository and alignment problems found between the two repositories.
	IssueCount int
	// AssignmentsCloneReady reports whether the local clone of the assignments
	// repository was refreshed. When false, the clone is missing or stale, and
	// callers, such as the test environment canary, must not use it.
	AssignmentsCloneReady bool
}

// UpdateFromCourseRepositories updates the database record for the course
// assignments and reports any problems found in the course's repositories.
//
// This is called in response to a push event to either the 'tests' or the
// 'assignments' repository, and may also be called manually by a teacher from
// the frontend. It performs a full update whichever repository was pushed, so
// that a missed webhook notification is reconciled by the next push to either
// repository. This works because the database update is idempotent, and the
// Docker image is rebuilt only when the Dockerfile's digest actually changed.
//
// The checks below are based on the latest content of both repositories.
// Failing to refresh the tests repository returns an error, while failing to
// refresh the assignments repository is only logged and reported in the
// returned UpdateResult; it does not prevent updating the database from the
// tests repository.
//
// Updates for a course are serialized, so the call may block the caller for an
// extended period. Issues are logged here, ensuring they reach the course log;
// callers are responsible for logging the returned error.
func UpdateFromCourseRepositories(ctx context.Context, runner ci.Runner, db database.Database, sc scm.SCM, course *qf.Course) (UpdateResult, error) {
	unlock := course.Lock()
	defer unlock()

	logger := qlog.FromContext(ctx)
	logger.Debug("updating from course repositories")
	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	clonedTestsRepo, clonedAssignmentsRepo, err := refreshCourseRepositories(ctx, sc, course)
	if err != nil {
		return UpdateResult{}, err
	}
	result := UpdateResult{AssignmentsCloneReady: clonedAssignmentsRepo != ""}

	// walk the cloned tests repository and extract the assignments and the course's Dockerfile
	assignments, buildContext, issues, err := ReadTestsRepository(clonedTestsRepo, course.GetID())
	if err != nil {
		return result, fmt.Errorf("reading tests repository content: %w", err)
	}
	result.IssueCount = len(issues)
	logRepositoryIssues(ctx, "tests repository issue", issues)

	if course.UpdateDockerfile(buildContext[ci.Dockerfile]) {
		// Rebuild the Docker image for the course tagged with the course code
		if err = BuildDockerImage(ctx, runner, course, buildContext); err != nil {
			return result, fmt.Errorf("building course image: %w", err)
		}
		// Update the course's DockerfileDigest in the database
		if err := db.UpdateCourse(course); err != nil {
			return result, fmt.Errorf("storing course Dockerfile digest: %w", err)
		}
	}

	if err = db.UpdateAssignments(assignments); err != nil {
		for _, assignment := range assignments {
			logger.Debug("assignment not updated in database", label.Assignment, assignment.GetName())
		}
		return result, fmt.Errorf("updating assignments in database: %w", err)
	}
	logger.Debug("assignments successfully updated from tests repository")

	if clonedAssignmentsRepo == "" {
		// The assignments repository could not be cloned, and
		// refreshCourseRepositories has logged why; the update itself is complete.
		return result, nil
	}
	alignmentIssues, err := CourseRepositoryIssues(clonedTestsRepo, clonedAssignmentsRepo, assignments)
	if err != nil {
		logger.Error("failed to compare course repositories", label.Error, err)
		return result, nil
	}
	logRepositoryIssues(ctx, "course repository issue", alignmentIssues)
	result.IssueCount += len(alignmentIssues)
	return result, nil
}

// refreshCourseRepositories refreshes the course's local tests and assignments
// clones, creating them if they are missing.
//
// The tests repository is required: failing to refresh it returns an error.
// The assignments repository is only needed for comparing the two repositories,
// so failing to refresh it is logged and reported as an empty assignmentsDir.
func refreshCourseRepositories(ctx context.Context, sc scm.SCM, course *qf.Course) (testsDir, assignmentsDir string, err error) {
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

// BuildDockerImage builds the Docker image for the given course, tagged with
// the course code, from the Dockerfile in the given build context.
func BuildDockerImage(ctx context.Context, runner ci.Runner, course *qf.Course, buildContext map[string]string) error {
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
