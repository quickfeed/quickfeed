package scm

import (
	"context"
	"errors"
)

// SCM is a common interface for different source code management solutions,
// i.e., GitHub and GitLab.
type SCM interface {
	// Lists directories which can be used as a course directory.
	ListDirectories(context.Context) ([]*Directory, error)
	// Creates a new directory.
	CreateDirectory(context.Context, *CreateDirectoryOptions) (*Directory, error)
	// Gets a directory.
	GetDirectory(context.Context, uint64) (*Directory, error)
}

// NewSCMClient returns a new provider client implementing the SCM interface.
func NewSCMClient(provider, token string) (SCM, error) {
	switch provider {
	case "github":
		return NewGithubSCMClient(token), nil
	case "gitlab":
		return NewGitlabSCMClient(token), nil
	}
	return nil, errors.New("invalid provider: " + provider)
}

// Directory represents an entity which is capable of managing source code
// repositories as well as user access to those repositories.
type Directory struct {
	ID   uint64
	Name string
}

// CreateDirectoryOptions contains information on how a directory should be
// created.
type CreateDirectoryOptions struct {
	Name string
}
