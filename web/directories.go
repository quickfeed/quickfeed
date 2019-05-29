package web

import (
	"context"
	"log"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"
)

// ListDirectories returns all directories which can be used as a course
// directory from the given provider.
func ListDirectories(ctx context.Context, db database.Database, scm scm.SCM) (*pb.Directories, error) {

	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	directories, err := scm.ListDirectories(ctx)
	if err != nil {
		return nil, err
	}

	organizations := make([]*pb.Directory, 0)
	for i, directory := range directories {
		log.Println("ListDirectories: organization ", i, " found: ", directory.ID)
		plan, err := scm.GetPaymentPlan(ctx, directory.ID)
		if err != nil {
			log.Println("ListDirectories: Error getting payment plan for organization: ", directory.GetPath())
			return nil, err
		}
		log.Println("ListDirectories: plan for Organization: ", directory.Path, " is ", plan.Name, " includes ", plan.PrivateRepos, " private repos")
		repos, err := scm.GetRepositories(ctx, directory)
		if err != nil {
			log.Println("ListDirectories: Error getting repos for organization: ", directory.GetPath())
			return nil, err
		}
		for i, repo := range repos {
			log.Println(i, ": repo ", repo.Path, " with url ", repo.WebURL)
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
		log.Println("hasCourse repo checks repo path: ", repo.Path, " on url ", repo.WebURL)
		if repo.Path == "course-info" {
			hasRepo = true
		}
	}
	return hasRepo
}
