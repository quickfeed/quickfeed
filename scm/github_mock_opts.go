package scm

import (
	"github.com/google/go-github/v62/github"
	"github.com/gosimple/slug"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

type mockOptions struct {
	orgs      []github.Organization
	repos     []github.Repository
	members   []github.Membership
	groups    map[string]map[string][]github.User                   // map: owner -> repo -> collaborators
	issues    map[string]map[string][]github.Issue                  // map: owner -> repo -> issues
	comments  map[string]map[string]map[int64][]github.IssueComment // map: owner -> repo -> issue ID -> comments
	reviewers map[string]map[string]map[int]github.ReviewersRequest // map: owner -> repo -> pull requests ID -> reviewers
}

// hasOrgRepo returns true if the given organization and repository exists in the mock data.
func (s mockOptions) hasOrgRepo(orgName, repoName string) bool {
	for _, repo := range s.repos {
		if repo.GetOrganization().GetLogin() == orgName && repo.GetName() == repoName {
			return true
		}
	}
	return false
}

// matchOrgFunc calls f with the organization that matches orgName and returns true if found.
func (s mockOptions) matchOrgFunc(orgName string, f func(github.Organization)) bool {
	for _, org := range s.orgs {
		if org.GetLogin() == orgName {
			f(org)
			return true
		}
	}
	return false
}

func newMockOptions() *mockOptions {
	return &mockOptions{
		orgs:      make([]github.Organization, 0),
		repos:     make([]github.Repository, 0),
		members:   make([]github.Membership, 0),
		groups:    map[string]map[string][]github.User{},
		issues:    map[string]map[string][]github.Issue{},
		comments:  map[string]map[string]map[int64][]github.IssueComment{},
		reviewers: map[string]map[string]map[int]github.ReviewersRequest{},
	}
}

type MockOption func(*mockOptions)

func WithOrgs(orgs ...github.Organization) MockOption {
	return func(opts *mockOptions) {
		opts.orgs = append(opts.orgs, orgs...)
	}
}

func WithRepos(repos ...github.Repository) MockOption {
	return func(opts *mockOptions) {
		opts.repos = append(opts.repos, repos...)
	}
}

func WithMembers(members ...github.Membership) MockOption {
	return func(opts *mockOptions) {
		opts.members = append(opts.members, members...)
	}
}

func WithGroups(groups map[string]map[string][]github.User) MockOption {
	return func(opts *mockOptions) {
		opts.groups = groups
	}
}

func WithReviewers(reviewers map[string]map[string]map[int]github.ReviewersRequest) MockOption {
	return func(opts *mockOptions) {
		opts.reviewers = reviewers
	}
}

// WithMockCourses sets up mock data based on qtest.MockCourses with complete
// course organizations and four repositories.
func WithMockCourses() MockOption {
	return func(opts *mockOptions) {
		for _, course := range qtest.MockCourses {
			ghOrg := toOrg(course)
			opts.orgs = append(opts.orgs, ghOrg)
			for _, repo := range []string{"info", "assignments", "tests", qf.StudentRepoName("meling")} {
				opts.repos = append(opts.repos, toRepo(&ghOrg, repo))
			}
		}
	}
}

var toOrg = func(course *qf.Course) github.Organization {
	return github.Organization{
		ID:    github.Int64(int64(course.ScmOrganizationID)),
		Login: github.String(slug.Make(course.ScmOrganizationName)),
	}
}

var toRepo = func(org *github.Organization, name string) github.Repository {
	return github.Repository{
		Organization: org,
		Name:         github.String(name),
	}
}
