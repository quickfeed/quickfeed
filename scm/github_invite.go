package scm

import (
	"context"
	"fmt"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

// newGithubSCMClient returns a new Github client for accepting invitations
// on a user's behalf.
func newGithubInviteClient(token string) *github.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	return github.NewClient(httpClient)
}

// acceptOrgInvitation accepts an organization membership invitation.
// Returns the new refresh token if successful.
func (s *GithubSCM) acceptOrgInvitation(ctx context.Context, opt *InvitationOptions) (string, error) {
	if !opt.valid() {
		return "", fmt.Errorf("invalid options: %+v", opt)
	}
	exchangeToken, err := s.config.ExchangeToken(opt.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to exchange refresh token for %s: %w", opt.Login, err)
	}
	userSCM := newGithubInviteClient(exchangeToken.AccessToken)
	state := "active"
	if _, _, err := userSCM.Organizations.EditOrgMembership(ctx, "", opt.Owner, &github.Membership{State: &state}); err != nil {
		return "", fmt.Errorf("failed to accept invitation to organization %s: %w", opt.Owner, err)
	}
	// Return the new refresh token.
	return exchangeToken.RefreshToken, nil
}
