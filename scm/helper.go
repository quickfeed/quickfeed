package scm

import (
	"context"
	"fmt"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/qf"
	"golang.org/x/oauth2"
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
	member = &github.Membership{Role: github.String(OrgMember)}
	admin  = &github.Membership{Role: github.String(OrgOwner)}
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
)

// newUserGithubClient creates a GitHub client using the provided user access token.
// This client is used to perform actions on behalf of the user, such as accepting invitations.
func newUserGithubClient(token string) *github.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	return github.NewClient(httpClient)
}

// isDirty returns true if the list of provided repositories contains
// any of the repositories that QuickFeed wants to create.
func isDirty(repos []*Repository) bool {
	if len(repos) == 0 {
		return false
	}
	for _, repo := range repos {
		if _, exists := RepoPaths[repo.Repo]; exists {
			return true
		}
	}
	return false
}
