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

// AcceptRepositoryInvites implements the SCMInvite interface
func (s *GithubSCM) AcceptRepositoryInvites(ctx context.Context, opt *RepositoryInvitationOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "AcceptRepositoryInvites",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

	userSCM := newGithubInviteClient(opt.Token)
	for _, repo := range []string{qf.InfoRepo, qf.AssignmentsRepo, qf.StudentRepoName(opt.Login)} {
		// Important: Get repository invitations using the GitHub App client.
		repoInvites, _, err := s.client.Repositories.ListInvitations(ctx, opt.Owner, repo, &github.ListOptions{})
		if err != nil {
			return ErrFailedSCM{
				GitError: fmt.Errorf("failed to fetch GitHub repository invitations: %w", err),
				Method:   "AcceptRepositoryInvites",
				Message:  "failed to fetch GitHub repository invitations",
			}
		}

		for _, invite := range repoInvites {
			if invite.Invitee.GetLogin() != opt.Login {
				// Ignore unrelated invites
				continue
			}
			// Important: Accept repository invitations using the user-specific GitHub client.
			if _, err := userSCM.Users.AcceptInvitation(ctx, invite.GetID()); err != nil {
				return ErrFailedSCM{
					GitError: fmt.Errorf("failed to accept GitHub repository invitation: %w", err),
					Method:   "AcceptRepositoryInvites",
					Message:  fmt.Sprintf("failed to accept invitation for repo: %s", invite.Repo.GetName()),
				}
			}
		}
	}

	state := "active"
	if _, _, err := userSCM.Organizations.EditOrgMembership(ctx, "", opt.Owner, &github.Membership{State: &state}); err != nil {
		return ErrFailedSCM{
			GitError: fmt.Errorf("failed to accept GitHub organization invitation: %w", err),
			Method:   "AcceptOrganizationInvite",
			Message:  fmt.Sprintf("failed to accept organization invite for org: %s, user: %s", opt.Owner, opt.Login),
		}
	}
	return nil
}
