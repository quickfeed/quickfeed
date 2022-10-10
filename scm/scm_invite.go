package scm

import (
	"context"
)

type SCMInvite interface {
	AcceptInvite(ctx context.Context, inviteID int64) error
	AcceptOrganizationInvite(ctx context.Context, orgName string) error
}

// RepositoryInvitationOptions contains information on which organization and user to accept invitations for.
type RepositoryInvitationOptions struct {
	Login   string    // GitHub username.
	Owner   string    // Name of the organization.
	UserSCM SCMInvite // SCM client for the user.
}

// NewInviteOnlySCMClient returns a new provider client implementing the SCM interface.
func NewInviteOnlySCMClient(token string) SCMInvite {
	return newGithubInviteClient(token)
}
