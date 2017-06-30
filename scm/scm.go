package scm

import "context"

// SCM is a common interface for different source code management solutions,
// i.e., GitHub and GitLab.
type SCM interface {
	// Lists directories which can be used as a course directory.
	ListDirectories(context.Context) ([]*Directory, error)
	// Creates a new directory.
	CreateDirectory(context.Context, *CreateDirectoryOptions) (*Directory, error)
	// Gets a directory.
	GetDirectory(context.Context, int) (*Directory, error)
}

// Directory represents an entity which is capable of managing source code
// repositories as well as user access to those repositories.
type Directory struct {
	ID   int
	Name string
}

// CreateDirectoryOptions contains information on how a directory should be
// created.
type CreateDirectoryOptions struct {
	Name string
}
