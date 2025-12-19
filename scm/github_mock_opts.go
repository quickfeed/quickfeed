package scm

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

type mockOptions struct {
	orgs       []github.Organization
	repos      []github.Repository
	members    []github.Membership
	groups     map[string]map[string][]github.User                   // map: owner -> repo -> collaborators
	issues     map[string]map[string][]github.Issue                  // map: owner -> repo -> issues
	comments   map[string]map[string]map[int64][]github.IssueComment // map: owner -> repo -> issue ID -> comments
	reviewers  map[string]map[string]map[int]github.ReviewersRequest // map: owner -> repo -> pull requests ID -> reviewers
	appConfigs map[string]github.AppConfig                           // map: code -> app config
	userID     int64                                                 // counter for generating unique user IDs
}

// DumpState returns a string representation of the mock state.
// This is used for debugging and testing purposes.
func (s mockOptions) DumpState() string {
	b := new(strings.Builder)
	fmt.Fprintln(b, "Mock state:")
	for i, org := range s.orgs {
		fmt.Fprintf(b, "Org[%d]: %v\n", i, org)
	}
	for i := range s.repos {
		fmt.Fprintf(b, "Repo[%d]: %v\n", i, s.repos[i])
	}
	for i, member := range s.members {
		fmt.Fprintf(b, "Member[%d]: %v\n", i, member)
	}
	for owner, repos := range s.groups {
		for repo, members := range repos {
			fmt.Fprintf(b, "Group[%s][%s]: %v\n", owner, repo, members)
		}
	}
	for owner, repos := range s.issues {
		for repo, issues := range repos {
			for i, issue := range issues {
				fmt.Fprintf(b, "Issue[%s][%s][%d]: %v\n", owner, repo, i, issue)
			}
		}
	}
	for owner, repos := range s.comments {
		for repo, issues := range repos {
			for issueID, comments := range issues {
				for i, comment := range comments {
					fmt.Fprintf(b, "Comment[%s][%s][%d][%d]: %v\n", owner, repo, issueID, i, comment)
				}
			}
		}
	}
	for owner, repos := range s.reviewers {
		for repo, prs := range repos {
			for prID, reviewers := range prs {
				fmt.Fprintf(b, "Reviewers[%s][%s][%d]: %v\n", owner, repo, prID, reviewers)
			}
		}
	}
	return b.String()
}

// hasOrgRepo returns true if the given organization and repository exists in the mock data.
func (s mockOptions) hasOrgRepo(orgName, repoName string) bool {
	for i := range s.repos {
		repo := &s.repos[i]
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

// GetComment returns the comment for the given organization, repository, and matching comment ID.
// This is used to inspect the comments created/updated during testing; not part of the SCM interface.
func (s mockOptions) GetComment(orgName, repoName string, commentID int64) *github.IssueComment {
	if s.comments[orgName] == nil || s.comments[orgName][repoName] == nil {
		return nil
	}
	for _, comments := range s.comments[orgName][repoName] {
		for _, comment := range comments {
			if *comment.ID == commentID {
				return &comment
			}
		}
	}
	return nil
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
		userID:    0,
	}
}

// nextUserID returns the next unique user ID and increments the counter.
func (s *mockOptions) nextUserID() int64 {
	s.userID++
	return s.userID
}

// getUserID returns the user ID for the given login, or assigns a new one if not found.
func (s *mockOptions) getUserID(login string) int64 {
	for _, member := range s.members {
		if member.GetUser().GetLogin() == login {
			return member.GetUser().GetID()
		}
	}
	return s.nextUserID()
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
		// Deep clone: owner -> repo -> collaborators
		clonedGroups := make(map[string]map[string][]github.User)
		for owner, repos := range groups {
			clonedRepos := make(map[string][]github.User)
			for repoName, users := range repos {
				clonedUsers := make([]github.User, len(users))
				copy(clonedUsers, users)
				clonedRepos[repoName] = clonedUsers
			}
			clonedGroups[owner] = clonedRepos
		}
		opts.groups = clonedGroups
	}
}

func WithReviewers(reviewers map[string]map[string]map[int]github.ReviewersRequest) MockOption {
	return func(opts *mockOptions) {
		opts.reviewers = reviewers
	}
}

func WithIssues(issues map[string]map[string][]github.Issue) MockOption {
	return func(opts *mockOptions) {
		opts.issues = issues
	}
}

// WithMockOrgs sets up mock data with course organizations and members, if any.
// The first member in the list is the owner of the organization.
func WithMockOrgs(members ...string) MockOption {
	return func(opts *mockOptions) {
		for _, course := range qtest.MockCourses {
			ghOrg := toOrg(course)
			opts.orgs = append(opts.orgs, ghOrg)
			for i, member := range members {
				userID := opts.getUserID(member)
				if i == 0 {
					opts.members = append(opts.members, github.Membership{Organization: &ghOrg, Role: github.String(OrgOwner), User: &github.User{ID: github.Int64(userID), Login: github.String(member)}})
				} else {
					opts.members = append(opts.members, github.Membership{Organization: &ghOrg, Role: github.String(OrgMember), User: &github.User{ID: github.Int64(userID), Login: github.String(member)}})
				}
			}
		}
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

func WithMockAppConfig(configs map[string]github.AppConfig) MockOption {
	return func(opts *mockOptions) {
		opts.appConfigs = configs
	}
}

var toOrg = func(course *qf.Course) github.Organization {
	return github.Organization{
		ID:    github.Int64(int64(course.GetScmOrganizationID())),
		Login: github.String(course.GetScmOrganizationName()),
	}
}

var toRepo = func(org *github.Organization, name string) github.Repository {
	return github.Repository{
		Organization: org,
		Name:         github.String(name),
	}
}
