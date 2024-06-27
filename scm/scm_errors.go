package scm

import "errors"

var (
	// ErrNotMember indicates that the requested organization exists, but the current user
	// is not its member.
	ErrNotMember = errors.New("user is not a member of the organization")
	// ErrAlreadyExists indicates that one or more QuickFeed repositories
	// already exists for the directory (or GitHub organization).
	ErrAlreadyExists = errors.New("course repositories already exist for that organization: " + repoNames)
)

// SCMError is returned to provide detailed information
// to user about source of the error and possible solution
type SCMError struct {
	Method   string
	Message  string
	GitError error
}

// Error message includes name of the failed method and the original error message
// from GitHub, to make it suitable for informative back-end logging
func (e SCMError) Error() string {
	return "github method " + e.Method + " failed: " + e.GitError.Error() + "\n" + e.Message
}
