package web

import (
	"context"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"
)

// ListOrganizations returns all directories which can be used as a course
// directory from the given provider.
func ListOrganizations(ctx context.Context, db database.Database, scm scm.SCM) (*pb.Organizations, error) {

	orgs, err := scm.ListOrganizations(ctx)
	if err != nil {
		return nil, err
	}

	organizations := make([]*pb.Organization, 0)
	for _, org := range orgs {
		plan, err := scm.GetPaymentPlan(ctx, org.ID)
		if err != nil {
			return nil, err
		}
		repos, err := scm.GetRepositories(ctx, org)
		if err != nil {
			return nil, err
		}
		// check that course for that organization does not exist in the database
		course, _ := db.GetCourseByOrganizationID(org.ID)

		if plan.Name != "free" && !hasCourseRepo(repos) && course == nil {
			organizations = append(organizations, org)
		}
	}

	return &pb.Organizations{Organizations: organizations}, nil
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
