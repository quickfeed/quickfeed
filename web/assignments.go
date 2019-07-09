package web

import (
	"context"
	"io/ioutil"
	"os"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/scm"
)

// fetchAssignments returns a list of assignments for the given course, by
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
func fetchAssignments(c context.Context, s scm.SCM, course *pb.Course) ([]*pb.Assignment, error) {
	ctx, cancel := context.WithTimeout(c, MaxWait)
	defer cancel()

	org, err := s.GetOrganization(ctx, course.OrganizationID)
	if err != nil {
		return nil, err
	}

	cloneURL := s.CreateCloneURL(&scm.CreateClonePathOptions{
		Organization: org.Path,
		Repository:   pb.TestsRepo,
	})

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

	runner := ci.Local{}
	_, err = runner.Run(ctx, job)
	if err != nil {
		return nil, err
	}

	// parse assignments found in the cloned tests directory
	return ParseAssignments(cloneDir, course.ID)
}
