package web

import (
	"context"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"
)

// ListDirectories returns all directories which can be used as a course
// directory from the given provider.
func ListDirectories(ctx context.Context, db database.Database, scm scm.SCM) (*pb.Directories, error) {

	directories, err := scm.ListDirectories(ctx)
	if err != nil {
		return nil, err
	}

	organizations := make([]*pb.Directory, 0)
	for _, directory := range directories {
		plan, err := scm.GetPaymentPlan(ctx, directory.ID)
		if err != nil {
			return nil, err
		}
		repos, err := scm.GetRepositories(ctx, directory)
		if err != nil {
			return nil, err
		}
		// check that course for that organization does not exist in the database
		course, _ := db.GetCourseByDirectoryID(directory.ID)

		if plan.Name != "free" && !hasCourseRepo(repos) && course == nil {
			organizations = append(organizations, directory)
		}
	}

	return &pb.Directories{Directories: organizations}, nil
}

// test function for fake database, can use IsDirty on real database
func hasCourseRepo(repos []*scm.Repository) bool {
	hasRepo := false
	for _, repo := range repos {
		if repo.Path == "course-info" {
			hasRepo = true
		}
	}
	return hasRepo
}
