package scm

import (
	"context"
	"fmt"

	"github.com/google/go-github/v62/github"
)

// acceptOrgInvitation accepts an organization membership invitation.
func (s *GithubSCM) acceptOrgInvitation(ctx context.Context, opt *InvitationOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid options: %+v", opt)
	}
	userSCM := s.createInviteClientFn(opt.AccessToken)
	state := "active"
	if _, _, err := userSCM.Organizations.EditOrgMembership(ctx, "", opt.Owner, &github.Membership{State: &state}); err != nil {
		return fmt.Errorf("failed to accept invitation to organization %s: %w", opt.Owner, err)
	}
	return nil
}
