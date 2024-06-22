package scm

import (
	"errors"
	"fmt"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/qf"
)

// Organization roles
const (
	// OrgOwner is organization's owner
	OrgOwner = "admin"
	// OrgMember is organization's member
	OrgMember = "member"
)

const (
	private = true
	public  = !private
)

// Repository permission levels for users
var (
	// pullAccess allows only pull access to repository
	pullAccess = &github.RepositoryAddCollaboratorOptions{Permission: "pull"}
	// pushAccess allows pull and push access to repository
	pushAccess = &github.RepositoryAddCollaboratorOptions{Permission: "push"}
)

var (
	// RepoPaths maps from QuickFeed repository path names to a boolean indicating
	// whether or not the repository should be create as public or private.
	RepoPaths = map[string]bool{
		qf.InfoRepo:        public,
		qf.AssignmentsRepo: private,
		qf.TestsRepo:       private,
	}
	repoNames = fmt.Sprintf("(%s, %s, %s)",
		qf.InfoRepo, qf.AssignmentsRepo, qf.TestsRepo)

	// ErrNotMember indicates that the requested organization exists, but the current user
	// is not its member.
	ErrNotMember = errors.New("user is not a member of the organization")
	// ErrAlreadyExists indicates that one or more QuickFeed repositories
	// already exists for the directory (or GitHub organization).
	ErrAlreadyExists = errors.New("course repositories already exist for that organization: " + repoNames)
)

// Errors //

// ErrNotSupported is returned when the source code management solution used
// does not provide a sufficient API for the method called.
type ErrNotSupported struct {
	SCM    string
	Method string
}

func (e ErrNotSupported) Error() string {
	return "method " + e.Method + " not supported by " + e.SCM + " SCM"
}

// ErrFailedSCM is returned to provide detailed information
// to user about source of the error and possible solution
type ErrFailedSCM struct {
	Method   string
	Message  string
	GitError error
}

// Error message includes name of the failed method and the original error message
// from GitHub, to make it suitable for informative back-end logging
func (e ErrFailedSCM) Error() string {
	return "github method " + e.Method + " failed: " + e.GitError.Error() + "\n" + e.Message
}

// isDirty returns true if the list of provided repositories contains
// any of the repositories that QuickFeed wants to create.
func isDirty(repos []*Repository) bool {
	if len(repos) == 0 {
		return false
	}
	for _, repo := range repos {
		if _, exists := RepoPaths[repo.Path]; exists {
			return true
		}
	}
	return false
}
