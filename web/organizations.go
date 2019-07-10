package web

import (
	"context"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
	"github.com/jinzhu/gorm"
)

// getAvailableOrganizations returns all organizations that can be used as a course
// organization from the given SCM provider.
func (s *AutograderService) getAvailableOrganizations(ctx context.Context, sc scm.SCM) (*pb.Organizations, error) {
	orgs, err := sc.ListOrganizations(ctx)
	if err != nil {
		return nil, err
	}

	organizations := make([]*pb.Organization, 0)
	for _, org := range orgs {
		plan, err := sc.GetPaymentPlan(ctx, org.ID)
		if err != nil {
			s.logger.Errorf("couldn't fetch payment plan: %v", err)
			continue
		}
		repos, err := sc.GetRepositories(ctx, org)
		if err != nil {
			s.logger.Errorf("couldn't fetch repos: %v", err)
			continue
		}
		course, err := s.db.GetCourseByOrganizationID(org.ID)
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				continue
			}
			s.logger.Errorf("couldn't fetch course from database: %v", err)
		}

		// only include organizations with non-free plan,
		// that are not already used for another course (has Autograder Repos), and
		// that doesn't already exist in the database.
		if plan.Name != "free" && !isDirty(repos) && course == nil {
			organizations = append(organizations, org)
		}
	}

	return &pb.Organizations{Organizations: organizations}, nil
}
