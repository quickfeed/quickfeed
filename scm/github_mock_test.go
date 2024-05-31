package scm

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v62/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"
	"google.golang.org/protobuf/testing/protocmp"
)

// Note: GetByID uses the undocumented GitHub API endpoint /organizations/:id.
var GetByID mock.EndpointPattern = mock.EndpointPattern{
	Pattern: "/organizations/{id}",
	Method:  "GET",
}

// The original mock.GetReposContentsByOwnerByRepoByPath pattern was incorrectly specified as: {path:.+}
var GetReposContentsByOwnerByRepoByPath mock.EndpointPattern = mock.EndpointPattern{
	Pattern: "/repos/{owner}/{repo}/contents/{path:.*}",
	Method:  "GET",
}

var jsonFolderContent = `[
  {
    "name": "Dockerfile",
    "path": "scripts/Dockerfile",
    "sha": "873c7550c0fc40b07cf173382bc93028f8f87c06",
    "size": 316,
    "type": "file"
  },
  {
    "name": "run.sh",
    "path": "scripts/run.sh",
    "sha": "fa3515649d92a369bb4c212760bf54b5d4d00d4e",
    "size": 1381,
    "type": "file"
  }
]`

var (
	meling  = github.User{Login: github.String("meling")}
	leslie  = github.User{Login: github.String("leslie")}
	lamport = github.User{Login: github.String("lamport")}
	jostein = github.User{Login: github.String("jostein")}
)

// map for testing UpdateGroupMembers; owner -> group_repo -> collaborators
var groups = map[string]map[string][]github.User{
	"foo": {
		"info":        {},
		"assignments": {},
		"tests":       {},
		"meling-labs": {},
		"groupX":      {lamport},
	},
	"bar": {
		"groupY": {leslie},
		"groupZ": {},
	},
}

// NewGithubSCMClient returns a new Github client implementing the SCM interface.
func NewMockGithubSCMClient(logger *zap.SugaredLogger) *GithubSCM {
	orgs := []github.Organization{
		{ID: github.Int64(123), Login: github.String("foo")},
		{ID: github.Int64(456), Login: github.String("bar")},
	}
	repos := []github.Repository{
		{Organization: &orgs[0], Name: github.String("info")},
		{Organization: &orgs[0], Name: github.String("assignments")},
		{Organization: &orgs[0], Name: github.String("tests")},
		{Organization: &orgs[0], Name: github.String("meling-labs")},
	}
	memberships := []github.Membership{
		{Organization: &orgs[0], User: &meling, Role: github.String(OrgOwner)},
		{Organization: &orgs[1], User: &meling, Role: github.String(OrgMember)},
	}
	matchFn := func(orgName string, f func(github.Organization)) bool {
		for _, org := range orgs {
			if org.GetLogin() == orgName {
				f(org)
				return true
			}
		}
		return false
	}

	httpClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			GetByID,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				id, err := lookupInt("id", GetByID.Pattern, r.URL.Path)
				if err != nil {
					// Unreachable in this test
					panic(err)
				}
				for _, org := range orgs {
					if org.GetID() == int64(id) {
						_, _ = w.Write(mock.MustMarshal(org))
						return
					}
				}
				w.WriteHeader(http.StatusNotFound)
			}),
		),

		mock.WithRequestMatchHandler(
			mock.GetOrgsByOrg,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				org := lookup("org", mock.GetOrgsByOrg.Pattern, r.URL.Path)
				found := matchFn(org, func(o github.Organization) {
					_, _ = w.Write(mock.MustMarshal(o))
				})
				if !found {
					w.WriteHeader(http.StatusNotFound)
				}
			}),
		),

		mock.WithRequestMatchHandler(
			mock.GetOrgsReposByOrg,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				org := lookup("org", mock.GetOrgsByOrg.Pattern, r.URL.Path)
				found := matchFn(org, func(o github.Organization) {
					foundRepos := make([]github.Repository, 0)
					for _, repo := range repos {
						if repo.GetOrganization().GetLogin() == o.GetLogin() {
							foundRepos = append(foundRepos, repo)
						}
					}
					_, _ = w.Write(mock.MustMarshal(foundRepos))
				})
				if !found {
					w.WriteHeader(http.StatusNotFound)
				}
			}),
		),

		mock.WithRequestMatchHandler(
			mock.GetOrgsMembershipsByOrgByUsername,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				org := lookup("org", mock.GetOrgsMembershipsByOrgByUsername.Pattern, r.URL.Path)
				username := lookup("username", mock.GetOrgsMembershipsByOrgByUsername.Pattern, r.URL.Path)
				found := matchFn(org, func(o github.Organization) {
					for _, m := range memberships {
						if m.GetOrganization().GetLogin() == o.GetLogin() && m.GetUser().GetLogin() == username {
							_, _ = w.Write(mock.MustMarshal(m))
							return
						}
					}
					w.WriteHeader(http.StatusNotFound)
				})
				if !found {
					w.WriteHeader(http.StatusNotFound)
				}
			}),
		),

		mock.WithRequestMatchHandler(
			GetReposContentsByOwnerByRepoByPath,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// we only care about the owner and repo; we ignore the path component
				owner := lookup("owner", GetReposContentsByOwnerByRepoByPath.Pattern, r.URL.Path)
				repo := lookup("repo", GetReposContentsByOwnerByRepoByPath.Pattern, r.URL.Path)
				for _, re := range repos {
					if re.GetOrganization().GetLogin() == owner && re.GetName() == repo {
						_, _ = w.Write([]byte(jsonFolderContent))
						return
					}
				}
				w.WriteHeader(http.StatusNotFound)
			}),
		),

		// Mock handlers for UpdateGroupMembers
		mock.WithRequestMatchHandler(
			mock.GetReposCollaboratorsByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				pattern := mock.GetReposCollaboratorsByOwnerByRepo.Pattern
				owner := lookup("owner", pattern, r.URL.Path)
				repo := lookup("repo", pattern, r.URL.Path)

				collaborators := groups[owner][repo]
				if collaborators == nil {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(mock.MustMarshal(collaborators))
			}),
		),
		mock.WithRequestMatchHandler(
			mock.PutReposCollaboratorsByOwnerByRepoByUsername,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				pattern := mock.PutReposCollaboratorsByOwnerByRepoByUsername.Pattern
				owner := lookup("owner", pattern, r.URL.Path)
				repo := lookup("repo", pattern, r.URL.Path)
				username := lookup("username", pattern, r.URL.Path)

				collaborators := groups[owner][repo]
				if collaborators == nil {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				if slices.ContainsFunc(collaborators, func(u github.User) bool { return u.GetLogin() == username }) {
					// already exists; no need to add again
					w.WriteHeader(http.StatusNoContent)
					return
				}

				ghUser := github.User{Login: github.String(username)}
				groups[owner][repo] = append(collaborators, ghUser)
				invite := github.CollaboratorInvitation{
					Repo:    &github.Repository{Owner: &github.User{Login: github.String(owner)}, Name: github.String(repo)},
					Invitee: &ghUser,
				}
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write(mock.MustMarshal(invite))
			}),
		),
		mock.WithRequestMatchHandler(
			mock.DeleteReposCollaboratorsByOwnerByRepoByUsername,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				pattern := mock.DeleteReposCollaboratorsByOwnerByRepoByUsername.Pattern
				owner := lookup("owner", pattern, r.URL.Path)
				repo := lookup("repo", pattern, r.URL.Path)
				username := lookup("username", pattern, r.URL.Path)

				collaborators := groups[owner][repo]
				if collaborators == nil {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				collaborators = slices.DeleteFunc(collaborators, func(u github.User) bool {
					return u.GetLogin() == username
				})
				groups[owner][repo] = collaborators
				w.WriteHeader(http.StatusNoContent)
				_, _ = w.Write([]byte{})
			}),
		),
	)
	return &GithubSCM{
		logger:      logger,
		client:      github.NewClient(httpClient),
		clientV4:    githubv4.NewClient(httpClient),
		providerURL: "github.com",
	}
}

func TestGetOrganization2(t *testing.T) {
	orgFoo := &qf.Organization{ScmOrganizationID: 123, ScmOrganizationName: "foo"}
	orgBar := &qf.Organization{ScmOrganizationID: 456, ScmOrganizationName: "bar"}

	tests := []struct {
		name    string
		org     *OrganizationOptions // cannot be nil
		wantOrg *qf.Organization
		wantErr bool
	}{
		{name: "GetOrganization/IncompleteRequest", org: &OrganizationOptions{}, wantOrg: nil, wantErr: true},
		{name: "GetOrganization/IncompleteRequest/Username", org: &OrganizationOptions{Username: "meling"}, wantOrg: nil, wantErr: true},
		{name: "GetOrganization/IncompleteRequest/NewCourse", org: &OrganizationOptions{NewCourse: true}, wantOrg: nil, wantErr: true},
		{name: "GetOrganization/IncompleteRequest/NewCourse/Username", org: &OrganizationOptions{NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},

		{name: "GetOrganization", org: &OrganizationOptions{ID: 123}, wantOrg: orgFoo, wantErr: false},
		{name: "GetOrganization", org: &OrganizationOptions{ID: 456}, wantOrg: orgBar, wantErr: false},
		{name: "GetOrganization/Missing", org: &OrganizationOptions{ID: 789}, wantOrg: nil, wantErr: true}, // 789 does not exist

		{name: "GetOrganization/NewCourse", org: &OrganizationOptions{ID: 123, NewCourse: true}, wantOrg: nil, wantErr: true},
		{name: "GetOrganization/NewCourse", org: &OrganizationOptions{ID: 456, NewCourse: true}, wantOrg: orgBar, wantErr: false},
		{name: "GetOrganization/NewCourse/Missing", org: &OrganizationOptions{ID: 789, NewCourse: true}, wantOrg: nil, wantErr: true}, // 789 does not exist

		{name: "GetOrganization/Username", org: &OrganizationOptions{ID: 123, Username: "meling"}, wantOrg: orgFoo, wantErr: false},     // meling is owner of foo
		{name: "GetOrganization/Username", org: &OrganizationOptions{ID: 456, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is only member of bar, not owner
		{name: "GetOrganization/Username/Missing", org: &OrganizationOptions{ID: 789, Username: "meling"}, wantOrg: nil, wantErr: true}, // 789 does not exist

		{name: "GetOrganization/NewCourse/Username", org: &OrganizationOptions{ID: 123, NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is owner of foo, but foo is not empty (not new course)
		{name: "GetOrganization/NewCourse/Username", org: &OrganizationOptions{ID: 456, NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is only member of bar, not owner
		{name: "GetOrganization/NewCourse/Username/Missing", org: &OrganizationOptions{ID: 789, NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true}, // 789 does not exist

		{name: "GetOrganization", org: &OrganizationOptions{Name: "foo"}, wantOrg: orgFoo, wantErr: false},
		{name: "GetOrganization", org: &OrganizationOptions{Name: "bar"}, wantOrg: orgBar, wantErr: false},
		{name: "GetOrganization/Missing", org: &OrganizationOptions{Name: "baz"}, wantOrg: nil, wantErr: true}, // baz does not exist

		{name: "GetOrganization/NewCourse", org: &OrganizationOptions{Name: "foo", NewCourse: true}, wantOrg: nil, wantErr: true},
		{name: "GetOrganization/NewCourse", org: &OrganizationOptions{Name: "bar", NewCourse: true}, wantOrg: orgBar, wantErr: false},
		{name: "GetOrganization/NewCourse/Missing", org: &OrganizationOptions{Name: "baz", NewCourse: true}, wantOrg: nil, wantErr: true}, // baz does not exist

		{name: "GetOrganization/Username", org: &OrganizationOptions{Name: "foo", Username: "meling"}, wantOrg: orgFoo, wantErr: false},     // meling is owner of foo
		{name: "GetOrganization/Username", org: &OrganizationOptions{Name: "bar", Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is only member of bar, not owner
		{name: "GetOrganization/Username/Missing", org: &OrganizationOptions{Name: "baz", Username: "meling"}, wantOrg: nil, wantErr: true}, // baz does not exist

		{name: "GetOrganization/NewCourse/Username", org: &OrganizationOptions{Name: "foo", NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is owner of foo
		{name: "GetOrganization/NewCourse/Username", org: &OrganizationOptions{Name: "bar", NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is only member of bar, not owner
		{name: "GetOrganization/NewCourse/Username/Missing", org: &OrganizationOptions{Name: "baz", NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true}, // baz does not exist
	}
	s := NewMockGithubSCMClient(qtest.Logger(t))
	for _, tt := range tests {
		var name string
		if tt.org.ID == 0 && tt.org.Name == "" {
			name = tt.name
		} else if tt.org.ID != 0 {
			name = fmt.Sprintf("%s/ID=%d", tt.name, tt.org.ID)
		} else {
			name = fmt.Sprintf("%s/Name=%s", tt.name, tt.org.Name)
		}
		t.Run(name, func(t *testing.T) {
			gotOrg, gotErr := s.GetOrganization(context.Background(), tt.org)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetOrganization() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.wantOrg, gotOrg, protocmp.Transform()); diff != "" {
				t.Errorf("GetOrganization() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetRepositories(t *testing.T) {
	tests := []struct {
		name    string
		org     *qf.Organization
		want    []*Repository
		wantErr bool
	}{
		{name: "GetRepositories", org: &qf.Organization{ScmOrganizationName: "foo"}, want: []*Repository{
			{OrgID: 123, Path: "info"},
			{OrgID: 123, Path: "assignments"},
			{OrgID: 123, Path: "tests"},
			{OrgID: 123, Path: "meling-labs"},
		}},
		{name: "GetRepositoriesNilOrg", org: nil, want: nil, wantErr: true},
		{name: "GetRepositoriesNoOrgName", org: &qf.Organization{}, want: nil, wantErr: true},
		{name: "GetRepositoriesNotFound", org: &qf.Organization{ScmOrganizationName: "bar"}, want: []*Repository{}, wantErr: false},
	}
	s := NewMockGithubSCMClient(qtest.Logger(t))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetRepositories(context.Background(), tt.org)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepositories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("GetRepositories() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRepositoryIsEmpty(t *testing.T) {
	tests := []struct {
		name      string
		opt       *RepositoryOptions
		wantEmpty bool
	}{
		{name: "RepositoryIsEmpty/IncompleteRequest/Owner=%s/Repo=%s", opt: &RepositoryOptions{}, wantEmpty: true},
		{name: "RepositoryIsEmpty/IncompleteRequest/Owner=%s/Repo=%s", opt: &RepositoryOptions{Owner: "foo"}, wantEmpty: true},
		{name: "RepositoryIsEmpty/IncompleteRequest/Owner=%s/Repo=%s", opt: &RepositoryOptions{Path: "info"}, wantEmpty: true},

		{name: "RepositoryIsEmpty/NonEmpty/Owner=%s/Repo=%s", opt: &RepositoryOptions{Owner: "foo", Path: "info"}, wantEmpty: false},
		{name: "RepositoryIsEmpty/NonEmpty/Owner=%s/Repo=%s", opt: &RepositoryOptions{Owner: "foo", Path: "assignments"}, wantEmpty: false},
		{name: "RepositoryIsEmpty/NonEmpty/Owner=%s/Repo=%s", opt: &RepositoryOptions{Owner: "foo", Path: "tests"}, wantEmpty: false},
		{name: "RepositoryIsEmpty/NonEmpty/Owner=%s/Repo=%s", opt: &RepositoryOptions{Owner: "foo", Path: "meling-labs"}, wantEmpty: false},

		{name: "RepositoryIsEmpty/Empty/Owner=%s/Repo=%s", opt: &RepositoryOptions{Owner: "bar", Path: "info"}, wantEmpty: true},
		{name: "RepositoryIsEmpty/Empty/Owner=%s/Repo=%s", opt: &RepositoryOptions{Owner: "bar", Path: "assignments"}, wantEmpty: true},
		{name: "RepositoryIsEmpty/Empty/Owner=%s/Repo=%s", opt: &RepositoryOptions{Owner: "bar", Path: "tests"}, wantEmpty: true},
		{name: "RepositoryIsEmpty/Empty/Owner=%s/Repo=%s", opt: &RepositoryOptions{Owner: "bar", Path: "meling-labs"}, wantEmpty: true},
	}
	s := NewMockGithubSCMClient(qtest.Logger(t))
	for _, tt := range tests {
		name := fmt.Sprintf(tt.name, tt.opt.Owner, tt.opt.Path)
		t.Run(name, func(t *testing.T) {
			gotIsEmpty := s.RepositoryIsEmpty(context.Background(), tt.opt)
			if gotIsEmpty != tt.wantEmpty {
				t.Errorf("RepositoryIsEmpty() = %v, want %v", gotIsEmpty, tt.wantEmpty)
			}
		})
	}
}

func TestUpdateGroupMembers(t *testing.T) {
	tests := []struct {
		name      string
		org       *GroupOptions
		wantUsers []github.User
		wantErr   bool
	}{
		{name: "IncompleteRequest", org: &GroupOptions{}, wantErr: true},
		{name: "IncompleteRequest", org: &GroupOptions{Organization: "foo"}, wantErr: true},
		{name: "IncompleteRequest", org: &GroupOptions{GroupName: "a"}, wantErr: true},
		{name: "IncompleteRequest", org: &GroupOptions{Users: []string{"meling"}}, wantErr: true},
		{name: "IncompleteRequest", org: &GroupOptions{Organization: "foo", Users: []string{"meling"}}, wantErr: true},
		{name: "IncompleteRequest", org: &GroupOptions{GroupName: "a", Users: []string{"meling"}}, wantErr: true},

		{name: "CompleteRequest/NotFound", org: &GroupOptions{Organization: "foo", GroupName: "a"}, wantErr: true},
		{name: "CompleteRequest/NotFound", org: &GroupOptions{Organization: "x", GroupName: "info"}, wantErr: true},
		{name: "CompleteRequest/NotFound", org: &GroupOptions{Organization: "foo", GroupName: "a", Users: []string{"meling"}}, wantErr: true},
		{name: "CompleteRequest/NotFound", org: &GroupOptions{Organization: "x", GroupName: "info", Users: []string{"meling"}}, wantErr: true},

		{name: "CompleteRequest", org: &GroupOptions{Organization: "foo", GroupName: "info", Users: []string{}}, wantErr: false, wantUsers: []github.User{}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "foo", GroupName: "groupX", Users: []string{"meling"}}, wantErr: false, wantUsers: []github.User{meling}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "foo", GroupName: "groupX", Users: []string{"meling", "leslie"}}, wantErr: false, wantUsers: []github.User{meling, leslie}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "foo", GroupName: "groupX", Users: []string{"meling", "leslie", "lamport"}}, wantErr: false, wantUsers: []github.User{meling, leslie, lamport}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "bar", GroupName: "groupY", Users: []string{"leslie", "lamport"}}, wantErr: false, wantUsers: []github.User{leslie, lamport}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "bar", GroupName: "groupY", Users: []string{"leslie"}}, wantErr: false, wantUsers: []github.User{leslie}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "bar", GroupName: "groupY", Users: []string{}}, wantErr: false, wantUsers: []github.User{}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "bar", GroupName: "groupZ", Users: []string{"leslie"}}, wantErr: false, wantUsers: []github.User{leslie}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "bar", GroupName: "groupZ", Users: []string{}}, wantErr: false, wantUsers: []github.User{}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "bar", GroupName: "groupZ", Users: []string{"leslie", "lamport"}}, wantErr: false, wantUsers: []github.User{leslie, lamport}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "bar", GroupName: "groupZ", Users: []string{"jostein"}}, wantErr: false, wantUsers: []github.User{jostein}},
	}
	s := NewMockGithubSCMClient(qtest.Logger(t))
	for _, tt := range tests {
		users := "nil"
		if tt.org.Users != nil {
			users = fmt.Sprintf("%v", tt.org.Users)
		}
		name := fmt.Sprintf("%s/Organization=%s/GroupName=%s/Users=%v", tt.name, tt.org.Organization, tt.org.GroupName, users)
		t.Run(name, func(t *testing.T) {
			if err := s.UpdateGroupMembers(context.Background(), tt.org); (err != nil) != tt.wantErr {
				t.Errorf("UpdateGroupMembers() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantUsers == nil {
				return
			}
			// verify the state of the groups after the test
			if diff := cmp.Diff(tt.wantUsers, groups[tt.org.Organization][tt.org.GroupName], protocmp.Transform()); diff != "" {
				t.Errorf("UpdateGroupMembers() mismatch (-want +got):\n%s", diff)
			}
		})
	}

	// expected state after calling the sequence of UpdateGroupMembers
	// owner -> repo -> collaborators
	wantGroups := map[string]map[string][]github.User{
		"foo": {
			"info":        {},
			"assignments": {},
			"tests":       {},
			"meling-labs": {},
			"groupX":      {meling, leslie, lamport},
		},
		"bar": {
			"groupY": {},
			"groupZ": {jostein},
		},
	}
	// verify the state of the groups after the sequence of UpdateGroupMembers
	if diff := cmp.Diff(wantGroups, groups, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateGroupMembers() mismatch (-want +got):\n%s", diff)
	}
}
