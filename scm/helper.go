package scm

import (
	"errors"
	"fmt"

	"github.com/quickfeed/quickfeed/qf"
)

const (
	// Organization roles //

	// OrgOwner is organization's owner
	OrgOwner = "admin"
	// OrgMember is organization's member
	OrgMember = "member"

	// Team roles //

	// TeamMaintainer can add and delete team users and repos
	TeamMaintainer = "maintainer"
	// TeamMember is a regular member
	TeamMember = "member"

	// Repository permission levels for organization //

	// OrgPull allows only pull access to organization repositories
	OrgPull = "read"
	// OrgPush allows pull and push access to organization repositories
	OrgPush = "write"
	// OrgFull allows to pull/push, create, remove and update organization repositories
	OrgFull = "admin"
	// OrgNone allows no access to organization repositories
	OrgNone = "none"

	// Repository permission levels for a user //

	// RepoPull allows only pull access to repository
	RepoPull = "pull"
	// RepoPush allows pull and push access to repository
	RepoPush = "push"
	// RepoFull allows full access to repository
	RepoFull = "admin"

	// Standard team names

	// TeachersTeam is the team with all teachers and teaching assistants of a course.
	TeachersTeam = "allteachers"
)

const (
	private = true
	public  = !private
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
	// ErrNotOwner indicates that user has no admin rights in the requested organization.
	ErrNotOwner = errors.New("user is not an owner of the organization")
	// ErrMissingInstallation indicates that GitHub application is not installed on organization.
	ErrMissingInstallation = errors.New("github application is not installed on the course organization")
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
