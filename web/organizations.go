package web

import (
	"context"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

func (s *QuickFeedService) getOrganization(ctx context.Context, sc scm.SCM, org string, user string) (*qf.Organization, error) {
	return sc.GetOrganization(ctx, &scm.GetOrgOptions{Name: org, Username: user})
}
