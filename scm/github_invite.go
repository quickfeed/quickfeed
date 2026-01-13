package scm

import (
	"context"
	"fmt"

	"github.com/google/go-github/v62/github"
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

// acceptRepositoryInvitation accepts a repository invitation for a specific repository.
// Returns the new refresh token if successful.
func (s *GithubSCM) acceptRepositoryInvitation(ctx context.Context, opt *InvitationOptions) (string, error) {
	if !opt.valid() {
		return "", fmt.Errorf("invalid options: %+v", opt)
	}
	if opt.Repository == "" {
		return "", fmt.Errorf("repository name is required")
	}
	exchangeToken, err := s.config.ExchangeToken(opt.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to exchange refresh token for %s: %w", opt.Login, err)
	}

	userSCM := newGithubInviteClient(exchangeToken.AccessToken)
	// Important: Get repository invitations using the GitHub App client.
	repoInvites, _, err := s.client.Repositories.ListInvitations(ctx, opt.Owner, opt.Repository, &github.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to fetch invitations for repository %s: %w", opt.Repository, err)
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
	// Return the new refresh token.
	return exchangeToken.RefreshToken, nil
}

// AcceptInvitations accepts assignments repository invitation and organization membership invitation.
// Note: Student repository invitations are accepted directly in createStudentRepo during enrollment.
// With the fork-based approach, the assignments repo invitation must be accepted before
// creating the student fork (which requires assignments access for private forks).
func (s *GithubSCM) AcceptInvitations(ctx context.Context, opt *InvitationOptions) (string, error) {
	if !opt.valid() {
		return "", fmt.Errorf("invalid options: %+v", opt)
	}

	// Accept assignments repo invitation (info repo is public, so no invitation needed)
	newRefreshToken, err := s.acceptRepositoryInvitation(ctx, &InvitationOptions{
		Login:        opt.Login,
		Owner:        opt.Owner,
		Repository:   qf.AssignmentsRepo,
		RefreshToken: opt.RefreshToken,
	})
	if err != nil {
		return "", err
	}

	// Accept organization membership invitation
	exchangeToken, err := s.config.ExchangeToken(newRefreshToken)
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
