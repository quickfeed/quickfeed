package assignments

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/scm"
	"go.uber.org/zap"
)

// UpdateFromTestsRepo updates the database record for the course assignments.
func UpdateFromTestsRepo(logger *zap.SugaredLogger, db database.Database, repo *pb.Repository, course *pb.Course) {
	logger.Debugf("Updating %s from '%s' repository", course.GetCode(), pb.TestsRepo)
	s, err := scm.NewSCMClient(logger, course.GetProvider(), course.GetAccessToken())
	if err != nil {
		logger.Errorf("Failed to create SCM Client: %w", err)
		return
	}
	assignments, err := FetchAssignments(context.Background(), s, course)
	if err != nil {
		logger.Errorf("Failed to fetch assignments from '%s' repository: %w", pb.TestsRepo, err)
		return
	}

	for _, assignment := range assignments {
		logger.Debugf("Found assignment in '%s' repository: %v", pb.TestsRepo, assignment)
	}

	if err = db.UpdateAssignments(assignments); err != nil {
		for _, assignment := range assignments {
			logger.Debugf("Failed to update database for: %v", assignment)
		}
		logger.Errorf("Failed to update assignments in database: %w", err)
		return
	}

	logger.Debugf("Assignments for %s successfully updated from '%s' repo", course.GetCode(), pb.TestsRepo)
}

// FetchAssignments returns a list of assignments for the given course, by
// cloning the 'tests' repo for the given course and extracting the assignments
// from the 'assignment.yml' files, one for each assignment.
//
// Note: This will typically be called on a push event to the 'tests' repo,
// which should happen infrequently. It may also be called manually by a
// teacher/admin from the frontend. However, even if multiple invocations
// happen concurrently, the function is idempotent. That is, it only reads
// data from GitHub, processes the yml files and returns the assignments.
// The TempDir() function ensures that cloning is done in distinct temp
// directories, should there be concurrent calls to this function.
func FetchAssignments(c context.Context, sc scm.SCM, course *pb.Course) ([]*pb.Assignment, error) {
	ctx, cancel := context.WithTimeout(c, pb.MaxWait)
	defer cancel()

	// ensuring compatibility with the old database:
	if course.OrganizationPath == "" {
		org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{ID: course.OrganizationID})
		if err != nil {
			return nil, err
		}
		course.OrganizationPath = org.GetPath()
	}

	log.Printf("org %s\n", course.GetOrganizationPath())

	cloneURL := sc.CreateCloneURL(&scm.CreateClonePathOptions{
		Organization: course.OrganizationPath,
		Repository:   pb.TestsRepo,
	})
	log.Printf("cloneURL %v\n", cloneURL)

	cloneDir, err := ioutil.TempDir("", pb.TestsRepo)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(cloneDir)

	// clone the tests repository to cloneDir
	job := &ci.Job{
		Commands: []string{
			"cd " + cloneDir,
			"git clone " + cloneURL,
		},
	}
	log.Printf("cd %v\n", cloneDir)
	log.Printf("git clone %v\n", cloneURL)

	runner := ci.Local{}
	_, err = runner.Run(ctx, job)
	if err != nil {
		return nil, err
	}

	// parse assignments found in the cloned tests directory
	return parseAssignments(cloneDir, course.ID)
}
