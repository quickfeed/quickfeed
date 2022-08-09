package assignments

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/rand"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"go.uber.org/zap"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
)

// MaxWait is the maximum time allowed for updating a course's assignments
// and docker image before aborting.
const MaxWait = 5 * time.Minute

// UpdateFromTestsRepo updates the database record for the course assignments.
func UpdateFromTestsRepo(logger *zap.SugaredLogger, db database.Database, mgr *scm.Manager, course *qf.Course) {
	logger.Debugf("Updating %s from '%s' repository", course.GetCode(), qf.TestsRepo)
	ctx, cancel := context.WithTimeout(context.Background(), MaxWait)
	defer cancel()

	scm, err := mgr.SCMWithToken(ctx, logger, course.OrganizationPath)
	if err != nil {
		logger.Errorf("Failed to create SCM Client: %v", err)
		return
	}
	assignments, dockerfile, err := fetchAssignments(ctx, scm, course)
	if err != nil {
		logger.Errorf("Failed to fetch assignments from '%s' repository: %v", qf.TestsRepo, err)
		return
	}
	for _, assignment := range assignments {
		updateGradingCriteria(logger, db, assignment)
	}

	if dockerfile != "" && dockerfile != course.Dockerfile {
		// The course's Dockerfile was added or updated in the tests repository
		course.Dockerfile = dockerfile
		// Rebuild the Docker image for the course tagged with the course code
		if err = buildDockerImage(ctx, logger, course); err != nil {
			logger.Error(err)
			return
		}
		// Update the course's Dockerfile in the database
		if err := db.UpdateCourse(course); err != nil {
			logger.Debugf("Failed to update Dockerfile for course %s: %s", course.GetCode(), err)
			return
		}
	}

	// Does not store tasks associated with assignments; tasks are handled separately by handleTasks below
	if err = db.UpdateAssignments(assignments); err != nil {
		for _, assignment := range assignments {
			logger.Debugf("Failed to update database for: %v", assignment)
		}
		logger.Errorf("Failed to update assignments in database: %v", err)
		return
	}
	logger.Debugf("Assignments for %s successfully updated from '%s' repo", course.GetCode(), qf.TestsRepo)

	if err = synchronizeTasksWithIssues(ctx, db, scm, course, assignments); err != nil {
		logger.Errorf("Failed to create tasks on '%s' repository: %v", qf.TestsRepo, err)
		return
	}
}

// fetchAssignments returns a list of assignments for the given course, by
// cloning the 'tests' repo for the given course and extracting the assignments
// from the 'assignment.yml' files, one for each assignment. If there is a Dockerfile
// in 'tests/script' its content will also be returned.
//
// Note: This will typically be called in response to a push event to the 'tests' repo,
// which should happen infrequently. It may also be called manually by a teacher/admin
// from the frontend. However, even if multiple invocations happen concurrently,
// the function is idempotent. That is, it only reads data from GitHub, processes
// the yml files and returns the assignments. The os.MkdirTemp() function ensures that
// any concurrent calls to this function will always use distinct temp directories.
func fetchAssignments(ctx context.Context, sc scm.SCM, course *qf.Course) ([]*qf.Assignment, string, error) {
	dstDir, err := os.MkdirTemp("", qf.TestsRepo)
	if err != nil {
		return nil, "", err
	}
	defer os.RemoveAll(dstDir)
	cloneDir, err := sc.Clone(ctx, &scm.CloneOptions{
		Organization: course.GetOrganizationPath(),
		Repository:   qf.TestsRepo,
		DestDir:      dstDir,
	})
	if err != nil {
		return nil, "", err
	}
	// walk the cloned tests repository and extract the assignments and the course's Dockerfile
	return readTestsRepositoryContent(cloneDir, course.ID)
}

// buildDockerImage builds the Docker image for the given course.
func buildDockerImage(ctx context.Context, logger *zap.SugaredLogger, course *qf.Course) error {
	docker, err := ci.NewDockerCI(logger)
	if err != nil {
		return fmt.Errorf("failed to set up docker client: %w", err)
	}
	defer func() { _ = docker.Close() }()

	logger.Debugf("Building %s's Dockerfile:\n%v", course.GetCode(), course.GetDockerfile())
	out, err := docker.Run(ctx, &ci.Job{
		Name:       course.GetCode() + "-" + rand.String(),
		Image:      course.GetCode(),
		Dockerfile: course.GetDockerfile(),
		Commands:   []string{`echo -n "Hello from Dockerfile"`},
	})
	logger.Debugf("Build completed: %s", out)
	if err != nil {
		return fmt.Errorf("failed to build image from %s's Dockerfile: %s", course.GetCode(), err)
	}
	return nil
}

// updateGradingCriteria will remove old grading criteria and related reviews when criteria.json gets updated
func updateGradingCriteria(logger *zap.SugaredLogger, db database.Database, assignment *qf.Assignment) {
	if len(assignment.GetGradingBenchmarks()) > 0 {
		gradingBenchmarks, err := db.GetBenchmarks(&qf.Assignment{
			CourseID: assignment.CourseID,
			Order:    assignment.Order,
		})
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// a new assignment, no actions required
				return
			}
			logger.Debugf("Failed to fetch assignment %s from database: %s", assignment.Name, err)
			return
		}
		if len(gradingBenchmarks) > 0 {
			if !cmp.Equal(assignment.GradingBenchmarks, gradingBenchmarks, cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(&qf.GradingBenchmark{}, "ID", "AssignmentID", "ReviewID"),
				protocmp.IgnoreFields(&qf.GradingCriterion{}, "ID", "BenchmarkID"),
				protocmp.IgnoreEnums(),
			}) {
				for _, bm := range gradingBenchmarks {
					for _, c := range bm.Criteria {
						if err := db.DeleteCriterion(c); err != nil {
							logger.Errorf("Failed to delete criteria %v: %s\n", c, err)
							return
						}
					}
					if err := db.DeleteBenchmark(bm); err != nil {
						logger.Errorf("Failed to delete benchmark %v: %s\n", bm, err)
						return
					}
				}
			} else {
				assignment.GradingBenchmarks = nil
			}
		}
	}
}
