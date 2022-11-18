package scm

import (
	"context"
	"fmt"

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/qf"
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

// AcceptInvitations accepts course invites.
func (s *GithubSCM) AcceptInvitations(ctx context.Context, opt *InvitationOptions) (string, error) {
	if !opt.valid() {
		return "", fmt.Errorf("invalid options: %+v", opt)
	}
	exchangeToken, err := s.config.ExchangeToken(opt.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to exchange refresh token for %s: %w", opt.Login, err)
	}

	userSCM := newGithubInviteClient(exchangeToken.AccessToken)
	for _, repo := range []string{qf.InfoRepo, qf.AssignmentsRepo, qf.StudentRepoName(opt.Login)} {
		// Important: Get repository invitations using the GitHub App client.
		repoInvites, _, err := s.client.Repositories.ListInvitations(ctx, opt.Owner, repo, &github.ListOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to fetch invitations for repository %s: %w", repo, err)
		}

		for _, invite := range repoInvites {
			if invite.Invitee.GetLogin() != opt.Login {
				// Ignore unrelated invites
				continue
			}
			// Important: Accept repository invitations using the user-specific GitHub client.
			if _, err := userSCM.Users.AcceptInvitation(ctx, invite.GetID()); err != nil {
				return "", fmt.Errorf("failed to accept invitation for repository %s: %w", invite.Repo.GetName(), err)
			}
		}
	}

	state := "active"
	if _, _, err := userSCM.Organizations.EditOrgMembership(ctx, "", opt.Owner, &github.Membership{State: &state}); err != nil {
		return "", fmt.Errorf("failed to accept invitation to organization %s: %w", opt.Owner, err)
	}
	// Return the new refresh token.
	return exchangeToken.RefreshToken, nil
}
