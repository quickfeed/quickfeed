package scm

import (
	"context"
	"fmt"

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/qf"
	"golang.org/x/oauth2"
)

// GithubSCM implements the SCM interface.
type GithubInviteSCM struct {
	client *github.Client
}

// newGithubSCMClient returns a new Github client implementing the SCMInvite interface.
func newGithubInviteClient(token string) *GithubInviteSCM {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	return &GithubInviteSCM{
		client: github.NewClient(httpClient),
	}
}

func (inviteSCM *GithubInviteSCM) AcceptInvite(ctx context.Context, inviteID int64) error {
	_, err := inviteSCM.client.Users.AcceptInvitation(ctx, inviteID)
	if err != nil {
		return err
	}
	return nil
}

func (inviteSCM *GithubInviteSCM) AcceptOrganizationInvite(ctx context.Context, orgName string) error {
	state := "active"
	_, _, err := inviteSCM.client.Organizations.EditOrgMembership(ctx, "", orgName, &github.Membership{State: &state})
	if err != nil {
		return err
	}
	return nil
}

// AcceptRepositoryInvites implements the SCMInvite interface
func (s *GithubSCM) AcceptRepositoryInvites(ctx context.Context, opt *RepositoryInvitationOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "AcceptRepositoryInvites",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

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
			if err := opt.UserSCM.AcceptInvite(ctx, invite.GetID()); err != nil {
				return ErrFailedSCM{
					GitError: fmt.Errorf("failed to accept GitHub repository invitation: %w", err),
					Method:   "AcceptRepositoryInvites",
					Message:  fmt.Sprintf("failed to accept invitation for repo: %s", invite.Repo.GetName()),
				}
			}
		}
	}
	if err := opt.UserSCM.AcceptOrganizationInvite(ctx, opt.Owner); err != nil {
		return ErrFailedSCM{
			GitError: fmt.Errorf("failed to accept GitHub organization invitation: %w", err),
			Method:   "AcceptOrganizationInvite",
			Message:  fmt.Sprintf("failed to accept organization invite for org: %s, user: %s", opt.Owner, opt.Login),
		}
	}
	return nil
}
