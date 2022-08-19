package scm

import (
	"context"

	"go.uber.org/zap"
)

type SCMInvite interface {
	// Accepts repository invites.
	AcceptRepositoryInvites(context.Context, *RepositoryInvitationOptions) error
}

// RepositoryInvitationOptions contains information on which organization and user to accept invitations for.
type RepositoryInvitationOptions struct {
	Login  string // GitHub username.
	Owner  string // Name of the organization.
	ScmApp SCM
}

func (r *RepositoryInvitationOptions) appClient() *GithubSCM {
	return r.ScmApp.(*GithubSCM)
}

// NewInviteOnlySCMClient returns a new provider client implementing the SCM interface.
func NewInviteOnlySCMClient(logger *zap.SugaredLogger, token string) SCMInvite {
	return NewGithubSCMClient(logger, token)
}
