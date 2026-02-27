package scm

import (
	"context"
	"fmt"

	"github.com/google/go-github/v62/github"
)

// acceptOrgInvitation accepts an organization membership invitation on behalf of the user.
func (s *GithubSCM) acceptOrgInvitation(ctx context.Context, opt *InvitationOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid options: %+v", opt)
	}
	userSCM := s.createUserClientFn(opt.AccessToken)
	state := "active"
	if _, _, err := userSCM.Organizations.EditOrgMembership(ctx, "", opt.Owner, &github.Membership{State: &state}); err != nil {
		return fmt.Errorf("failed to accept invitation for %s to organization %s: %w", opt.Login, opt.Owner, err)
	}
	return nil
}
