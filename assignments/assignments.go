package assignments

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"go.uber.org/zap"
)

// MaxWait is the maximum time allowed for updating a course's assignments
// and docker image before aborting.
const MaxWait = 5 * time.Minute

var updateMutex = sync.Mutex{}

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
func UpdateFromTestsRepo(logger *zap.SugaredLogger, runner ci.Runner, db database.Database, sc scm.SCM, course *qf.Course) ([]*qf.Assignment, bool) {
	updateMutex.Lock()
	defer updateMutex.Unlock()

	logger.Debugf("Updating %s from '%s' repository", course.GetCode(), qf.TestsRepo)
	ctx, cancel := context.WithTimeout(context.Background(), MaxWait)
	defer cancel()

	clonedTestsRepo, err := sc.Clone(ctx, &scm.CloneOptions{
		Organization: course.GetScmOrganizationName(),
		Repository:   qf.TestsRepo,
		DestDir:      course.CloneDir(),
	})
	if err != nil {
		logger.Errorf("Failed to clone '%s' repository: %v", qf.TestsRepo, err)
		return nil, false
	}
	logger.Debugf("Successfully cloned tests repository to: %s", clonedTestsRepo)

	// walk the cloned tests repository and extract the assignments and the course's Dockerfile
	assignments, dockerfile, updateTestsScript, err := readTestsRepositoryContent(clonedTestsRepo, course.ID)
	if err != nil {
		logger.Errorf("Failed to parse assignments from '%s' repository: %v", qf.TestsRepo, err)
		return nil, false
	}

	if course.UpdateDockerfile(dockerfile) {
		// Rebuild the Docker image for the course tagged with the course code
		if err = buildDockerImage(ctx, logger, runner, course); err != nil {
			logger.Error(err)
			return nil, false
		}
		// Update the course's DockerfileDigest in the database
		if err := db.UpdateCourse(course); err != nil {
			logger.Errorf("Failed to update Dockerfile for course %s: %v", course.GetCode(), err)
			return nil, false
		}
	}

	// Does not store tasks associated with assignments; tasks are handled separately by synchronizeTasksWithIssues below
	if err = db.UpdateAssignments(assignments); err != nil {
		for _, assignment := range assignments {
			logger.Debugf("Failed to update database for: %v", assignment)
		}
		logger.Errorf("Failed to update assignments in database: %v", err)
		return nil, false
	}
	logger.Debugf("Assignments for %s successfully updated from '%s' repo", course.GetCode(), qf.TestsRepo)

	if err = synchronizeTasksWithIssues(ctx, db, sc, course, assignments); err != nil {
		logger.Errorf("Failed to create tasks on '%s' repository: %v", qf.TestsRepo, err)
		return nil, false
	}
	return assignments, updateTestsScript != ""
}

// buildDockerImage builds the Docker image for the given course.
func buildDockerImage(ctx context.Context, logger *zap.SugaredLogger, runner ci.Runner, course *qf.Course) error {
	logger.Debugf("Building %s's Dockerfile:\n%v", course.GetCode(), course.GetDockerfile())
	out, err := runner.Run(ctx, &ci.Job{
		Name:       course.JobName(),
		Image:      course.DockerImage(),
		Dockerfile: course.GetDockerfile(),
		Commands:   []string{`echo -n "Hello from Dockerfile"`},
	})
	logger.Debugf("Build completed: %s", out)
	if err != nil {
		return fmt.Errorf("failed to build image from %s's Dockerfile: %s", course.GetCode(), err)
	}
	return nil
}
