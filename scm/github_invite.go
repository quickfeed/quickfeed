package scm

import (
	"context"
	"fmt"

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/qf"
	"golang.org/x/oauth2"
)

// newGithubSCMClient returns a new Github client implementing the SCMInvite interface.
func newGithubInviteClient(token string) *github.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	return github.NewClient(httpClient)
}

// AcceptRepositoryInvites accepts course invites.
func (s *GithubSCM) AcceptRepositoryInvites(ctx context.Context, opt *InvitationOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid options: %+v", opt)
	}

	userSCM := newGithubInviteClient(opt.Token)
	for _, repo := range []string{qf.InfoRepo, qf.AssignmentsRepo, qf.StudentRepoName(opt.Login)} {
		// Important: Get repository invitations using the GitHub App client.
		repoInvites, _, err := s.client.Repositories.ListInvitations(ctx, opt.Owner, repo, &github.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to fetch invitations for repository %s: %w", repo, err)
		}

		for _, invite := range repoInvites {
			if invite.Invitee.GetLogin() != opt.Login {
				// Ignore unrelated invites
				continue
			}
			// Important: Accept repository invitations using the user-specific GitHub client.
			if _, err := userSCM.Users.AcceptInvitation(ctx, invite.GetID()); err != nil {
				return fmt.Errorf("failed to accept invitation for repository %s: %w", invite.Repo.GetName(), err)
			}
		}
	}

	state := "active"
	if _, _, err := userSCM.Organizations.EditOrgMembership(ctx, "", opt.Owner, &github.Membership{State: &state}); err != nil {
		return fmt.Errorf("failed to accept invitation to organization %s: %w", opt.Owner, err)
	}
	return nil
}
