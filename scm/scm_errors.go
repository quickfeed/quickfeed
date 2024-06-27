package scm

import (
	"errors"
	"fmt"
)

var (
	// ErrNotMember indicates that the requested organization exists, but the current user
	// is not its member.
	ErrNotMember = errors.New("user is not a member of the organization")
	// ErrAlreadyExists indicates that one or more QuickFeed repositories
	// already exists for the directory (or GitHub organization).
	ErrAlreadyExists = errors.New("course repositories already exist for that organization: " + repoNames)
)

// Op describes an operation, such as "GetOrganization".
type Op string

// SCMError is returned to provide detailed information
// to user about source of the error and possible solution
type SCMError struct {
	Op      Op
	Message string
	Err     error
}

func (e SCMError) Error() string {
	return fmt.Sprintf("scm.%s: %v", e.Op, e.Message)
}
