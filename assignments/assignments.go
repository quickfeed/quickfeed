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
// and docker image before aborting.
const MaxWait = 15 * time.Minute

// UpdateFromTestsRepo updates the database record for the course assignments.
//
// This will be called in response to a push event to the 'tests' repo, which
// should happen infrequently. It may also be called manually by a teacher from
// the frontend.
//
// Note that calling this function concurrently is safe, but it may block the
// caller for an extended period, since it may involve cloning the tests repository,
// scanning the repository for assignments, building the Docker image, updating the
// database and synchronizing tasks to issues on the students' group repositories.
// The ctx is expected to carry a logger scoped with the course and the tests
// repository, so that the statements below do not repeat those attributes.
//
// The returned issues describe content problems in the tests repository that
// did not abort the update, such as a malformed json file for one assignment.
// Issues are logged here so webhook updates reach the course log.
// Callers are responsible for logging returned errors.
func UpdateFromTestsRepo(ctx context.Context, runner ci.Runner, db database.Database, sc scm.SCM, course *qf.Course) ([]RepoIssue, error) {
	unlock := course.Lock()
	defer unlock()

	logger := qlog.FromContext(ctx)
	logger.Debug("updating from tests repository")
	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	clonedTestsRepo, err := sc.Clone(ctx, &scm.CloneOptions{
		Organization: course.GetScmOrganizationName(),
		Repository:   qf.TestsRepo,
		DestDir:      course.CloneDir(),
	})
	if err != nil {
		return nil, fmt.Errorf("cloning tests repository: %w", err)
	}
	logger.Debug("cloned tests repository", label.Path, clonedTestsRepo)

	// walk the cloned tests repository and extract the assignments and the course's Dockerfile
	assignments, buildContext, issues, err := readTestsRepositoryContent(clonedTestsRepo, course.GetID())
	if err != nil {
		return nil, fmt.Errorf("reading tests repository content: %w", err)
	}
	for _, issue := range issues {
		logger.Warn("tests repository issue",
			label.Assignment, issue.Assignment, label.Path, issue.File, "problem", issue.Problem)
	}

	if course.UpdateDockerfile(buildContext[ci.Dockerfile]) {
		// Rebuild the Docker image for the course tagged with the course code
		if err = buildDockerImage(ctx, runner, course, buildContext); err != nil {
			return issues, fmt.Errorf("building course image: %w", err)
		}
		// Update the course's DockerfileDigest in the database
		if err := db.UpdateCourse(course); err != nil {
			return issues, fmt.Errorf("storing course Dockerfile digest: %w", err)
		}
	}

	if err = db.UpdateAssignments(assignments); err != nil {
		for _, assignment := range assignments {
			logger.Debug("assignment not updated in database", label.Assignment, assignment.GetName())
		}
		return issues, fmt.Errorf("updating assignments in database: %w", err)
	}
	logger.Debug("assignments successfully updated from tests repository")
	return issues, nil
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
