package scm

import (
	"context"
	"fmt"

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/qf"
)

// AcceptRepositoryInvites implements the SCMInvite interface
func (s *GithubSCM) AcceptRepositoryInvites(ctx context.Context, opt *RepositoryInvitationOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "AcceptRepositoryInvites",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

	for _, repo := range []string{qf.InfoRepo, qf.AssignmentRepo, qf.StudentRepoName(opt.Login)} {
		// Important: Get repository invitations using the GitHub App client.
		repoInvites, _, err := opt.appClient().client.Repositories.ListInvitations(ctx, opt.Owner, repo, &github.ListOptions{})
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
			_, err := s.client.Users.AcceptInvitation(ctx, invite.GetID())
			if err != nil {
				return ErrFailedSCM{
					GitError: fmt.Errorf("failed to accept GitHub repository invitation: %w", err),
					Method:   "AcceptRepositoryInvites",
					Message:  fmt.Sprintf("failed to accept invitation for repo: %s", invite.Repo.GetName()),
				}
			}
		}
	}
	return nil
}
