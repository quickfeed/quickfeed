package web

import (
	"context"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/scm"
)

func (s *AutograderService) getOrganization(ctx context.Context, sc scm.SCM, org string, user string) (*pb.Organization, error) {
	gitOrg, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{Name: org, Username: user})
	if err != nil {
		return nil, err
	}
	// check payment plan
	if gitOrg.GetPaymentPlan() == FreeOrgPlan {
		return nil, ErrFreePlan
	}
	// check course repos
	repos, err := sc.GetRepositories(ctx, gitOrg)
	if err != nil {
		return nil, err
	}
	if isDirty(repos) {
		return nil, ErrAlreadyExists
	}
	return gitOrg, nil
}
