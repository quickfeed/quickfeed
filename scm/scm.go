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
	// Create a new repository.
	CreateRepository(context.Context, *CreateRepositoryOptions) (*Repository, error)
	// Get repositories within directory.
	GetRepositories(context.Context, *Directory) ([]*Repository, error)
}

// NewSCMClient returns a new provider client implementing the SCM interface.
func NewSCMClient(provider, token string) (SCM, error) {
	switch provider {
	case "github":
		return NewGithubSCMClient(token), nil
	case "gitlab":
		return NewGitlabSCMClient(token), nil
	case "fake":
		return &FakeSCM{}, nil
	}
	return nil, errors.New("invalid provider: " + provider)
}

// Directory represents an entity which is capable of managing source code
// repositories as well as user access to those repositories.
type Directory struct {
	ID     uint64 `json:"id"`
	Path   string `json:"path"`
	Avatar string `json:"avatar,omitempty"`
}

// CreateDirectoryOptions contains information on how a directory should be
// created.
type CreateDirectoryOptions struct {
	Path string
	Name string
}

// Repository represents a git remote repository.
type Repository struct {
	ID   uint64
	Path string

	// Repository website.
	WebURL string
	// SSH clone URL.
	SSHURL string
	// HTTP(S) clone URL.
	HTTPURL string

	DirectoryID uint64
}

// CreateRepositoryOptions contains information on how a repository should be
// created.
type CreateRepositoryOptions struct {
	Path      string
	Directory *Directory
}

// ErrNotSupported is returned when the source code management solution used
// does not provide a sufficient API for the method called.
type ErrNotSupported struct {
	SCM    string
	Method string
}

func (e ErrNotSupported) Error() string {
	return "method" + e.Method + " not supported by " + e.SCM + " SCM"
}
