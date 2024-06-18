package scm

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

// mock organizations foo and bar
var (
	ghOrgFoo    = github.Organization{ID: github.Int64(123), Login: foo.Login}
	ghOrgBar    = github.Organization{ID: github.Int64(456), Login: bar.Login}
	ghOrgDat320 = github.Organization{ID: github.Int64(789), Login: github.String("dat320")}
)

// mock repositories for organization foo; bar has no repositories
var repos = []github.Repository{
	{ID: github.Int64(1), Organization: &ghOrgFoo, Name: github.String("info")},
	{ID: github.Int64(2), Organization: &ghOrgFoo, Name: github.String("assignments")},
	{ID: github.Int64(3), Organization: &ghOrgFoo, Name: github.String("tests")},
	{ID: github.Int64(4), Organization: &ghOrgFoo, Name: github.String("meling-labs")},
	{ID: github.Int64(5), Organization: &ghOrgFoo, Name: github.String("josie-labs")},
}

// memberships: user -> role; two members; one owner, one member
var members = []github.Membership{
	{Organization: &ghOrgFoo, User: &meling, Role: github.String(OrgOwner)},
	{Organization: &ghOrgBar, User: &meling, Role: github.String(OrgMember)},
}

// groups map: owner -> repo -> collaborators (only group repos should have collaborators)
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

// reviewers map: owner -> repo -> pull requests ID -> reviewers
var reviewers = map[string]map[string]map[int]github.ReviewersRequest{
	"foo": {
		"meling-labs": {
			1: {Reviewers: []string{"meling", "leslie"}},
			2: {Reviewers: []string{"lamport", "jostein"}},
		},
		"josie-labs": {
			1: {Reviewers: []string{"meling", "leslie"}},
			2: {Reviewers: []string{"lamport", "jostein"}},
		},
	},
}

func TestMockGetOrganization(t *testing.T) {
	orgFoo := &qf.Organization{ScmOrganizationID: 123, ScmOrganizationName: *foo.Login}
	orgBar := &qf.Organization{ScmOrganizationID: 456, ScmOrganizationName: *bar.Login}

	tests := []struct {
		name    string
		org     *OrganizationOptions // cannot be nil
		wantOrg *qf.Organization
		wantErr bool
	}{
		{name: "IncompleteRequest", org: &OrganizationOptions{}, wantOrg: nil, wantErr: true},
		{name: "IncompleteRequest", org: &OrganizationOptions{Username: "meling"}, wantOrg: nil, wantErr: true},
		{name: "IncompleteRequest", org: &OrganizationOptions{NewCourse: true}, wantOrg: nil, wantErr: true},
		{name: "IncompleteRequest", org: &OrganizationOptions{NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},

		{name: "CompleteRequest", org: &OrganizationOptions{ID: 123}, wantOrg: orgFoo, wantErr: false},
		{name: "CompleteRequest", org: &OrganizationOptions{ID: 456}, wantOrg: orgBar, wantErr: false},
		{name: "CompleteRequest/Missing", org: &OrganizationOptions{ID: 789}, wantOrg: nil, wantErr: true}, // 789 does not exist

		{name: "CompleteRequest", org: &OrganizationOptions{ID: 123, NewCourse: true}, wantOrg: nil, wantErr: true},
		{name: "CompleteRequest", org: &OrganizationOptions{ID: 456, NewCourse: true}, wantOrg: orgBar, wantErr: false},
		{name: "CompleteRequest/Missing", org: &OrganizationOptions{ID: 789, NewCourse: true}, wantOrg: nil, wantErr: true}, // 789 does not exist

		{name: "CompleteRequest", org: &OrganizationOptions{ID: 123, Username: "meling"}, wantOrg: orgFoo, wantErr: false},     // meling is owner of foo
		{name: "CompleteRequest", org: &OrganizationOptions{ID: 456, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is only member of bar, not owner
		{name: "CompleteRequest/Missing", org: &OrganizationOptions{ID: 789, Username: "meling"}, wantOrg: nil, wantErr: true}, // 789 does not exist

		{name: "CompleteRequest", org: &OrganizationOptions{ID: 123, NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is owner of foo, but foo is not empty (not new course)
		{name: "CompleteRequest", org: &OrganizationOptions{ID: 456, NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is only member of bar, not owner
		{name: "CompleteRequest/Missing", org: &OrganizationOptions{ID: 789, NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true}, // 789 does not exist

		{name: "CompleteRequest", org: &OrganizationOptions{Name: "foo"}, wantOrg: orgFoo, wantErr: false},
		{name: "CompleteRequest", org: &OrganizationOptions{Name: "bar"}, wantOrg: orgBar, wantErr: false},
		{name: "CompleteRequest/Missing", org: &OrganizationOptions{Name: "baz"}, wantOrg: nil, wantErr: true}, // baz does not exist

		{name: "CompleteRequest", org: &OrganizationOptions{Name: "foo", NewCourse: true}, wantOrg: nil, wantErr: true},
		{name: "CompleteRequest", org: &OrganizationOptions{Name: "bar", NewCourse: true}, wantOrg: orgBar, wantErr: false},
		{name: "CompleteRequest/Missing", org: &OrganizationOptions{Name: "baz", NewCourse: true}, wantOrg: nil, wantErr: true}, // baz does not exist

		{name: "CompleteRequest", org: &OrganizationOptions{Name: "foo", Username: "meling"}, wantOrg: orgFoo, wantErr: false},     // meling is owner of foo
		{name: "CompleteRequest", org: &OrganizationOptions{Name: "bar", Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is only member of bar, not owner
		{name: "CompleteRequest/Missing", org: &OrganizationOptions{Name: "baz", Username: "meling"}, wantOrg: nil, wantErr: true}, // baz does not exist

		{name: "CompleteRequest", org: &OrganizationOptions{Name: "foo", NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is owner of foo
		{name: "CompleteRequest", org: &OrganizationOptions{Name: "bar", NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is only member of bar, not owner
		{name: "CompleteRequest/Missing", org: &OrganizationOptions{Name: "baz", NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true}, // baz does not exist
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...), WithMembers(members...))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"ID", "Name", "Username", "NewCourse"}, tt.org.ID, tt.org.Name, tt.org.Username, tt.org.NewCourse)
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

func TestMockGetRepositories(t *testing.T) {
	tests := []struct {
		name    string
		org     *qf.Organization
		want    []*Repository
		wantErr bool
	}{
		{name: "IncompleteRequest/NilOrg", org: nil, want: nil, wantErr: true},
		{name: "IncompleteRequest/NoOrgName", org: &qf.Organization{}, want: nil, wantErr: true},
		{name: "CompleteRequest/NotFound", org: &qf.Organization{ScmOrganizationName: "bar"}, want: []*Repository{}, wantErr: false},
		{name: "CompleteRequest/FiveRepos", org: &qf.Organization{ScmOrganizationName: "foo"}, want: []*Repository{
			{ID: 1, OrgID: 123, Path: "info"},
			{ID: 2, OrgID: 123, Path: "assignments"},
			{ID: 3, OrgID: 123, Path: "tests"},
			{ID: 4, OrgID: 123, Path: "meling-labs"},
			{ID: 5, OrgID: 123, Path: "josie-labs"},
		}},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetRepositories(context.Background(), tt.org)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepositories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetRepositories() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMockRepositoryIsEmpty(t *testing.T) {
	tests := []struct {
		name      string
		opt       *RepositoryOptions
		wantEmpty bool
	}{
		{name: "IncompleteRequest", opt: &RepositoryOptions{}, wantEmpty: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Owner: "foo"}, wantEmpty: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Path: "info"}, wantEmpty: true},

		{name: "CompleteRequest/Empty", opt: &RepositoryOptions{Owner: "bar", Path: "info"}, wantEmpty: true},
		{name: "CompleteRequest/Empty", opt: &RepositoryOptions{Owner: "bar", Path: "assignments"}, wantEmpty: true},
		{name: "CompleteRequest/Empty", opt: &RepositoryOptions{Owner: "bar", Path: "tests"}, wantEmpty: true},
		{name: "CompleteRequest/Empty", opt: &RepositoryOptions{Owner: "bar", Path: "meling-labs"}, wantEmpty: true},

		{name: "CompleteRequest/NonEmpty", opt: &RepositoryOptions{Owner: "foo", Path: "info"}, wantEmpty: false},
		{name: "CompleteRequest/NonEmpty", opt: &RepositoryOptions{Owner: "foo", Path: "assignments"}, wantEmpty: false},
		{name: "CompleteRequest/NonEmpty", opt: &RepositoryOptions{Owner: "foo", Path: "tests"}, wantEmpty: false},
		{name: "CompleteRequest/NonEmpty", opt: &RepositoryOptions{Owner: "foo", Path: "meling-labs"}, wantEmpty: false},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Owner", "Path"}, tt.opt.Owner, tt.opt.Path)
		t.Run(name, func(t *testing.T) {
			gotIsEmpty := s.RepositoryIsEmpty(context.Background(), tt.opt)
			if gotIsEmpty != tt.wantEmpty {
				t.Errorf("RepositoryIsEmpty() = %v, want %v", gotIsEmpty, tt.wantEmpty)
			}
		})
	}
}

func TestMockUpdateGroupMembers(t *testing.T) {
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
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithGroups(groups))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Organization", "GroupName", "Users"}, tt.org.Organization, tt.org.GroupName, tt.org.Users)
		t.Run(name, func(t *testing.T) {
			if err := s.UpdateGroupMembers(context.Background(), tt.org); (err != nil) != tt.wantErr {
				t.Errorf("UpdateGroupMembers() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantUsers == nil {
				return
			}
			// verify the state of the groups after the test
			if diff := cmp.Diff(tt.wantUsers, s.groups[tt.org.Organization][tt.org.GroupName]); diff != "" {
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
	if diff := cmp.Diff(wantGroups, s.groups); diff != "" {
		t.Errorf("UpdateGroupMembers() mismatch (-want +got):\n%s", diff)
	}
}

func TestMockCreateCourse(t *testing.T) {
	// repositories that should be created for bar; manually sorted by path
	wantBarRepos := []*Repository{
		{OrgID: 456, Owner: "bar", Path: "assignments"},
		{OrgID: 456, Owner: "bar", Path: "info"},
		{OrgID: 456, Owner: "bar", Path: "meling-labs"},
		{OrgID: 456, Owner: "bar", Path: "tests"},
	}
	// we need to initialize the groups table (collaborators) to allow creating a course with meling as course creator
	g := map[string]map[string][]github.User{
		"bar": {
			"meling-labs": {},
		},
	}

	tests := []struct {
		name      string
		opt       *CourseOptions // cannot be nil
		wantRepos []*Repository
		wantErr   bool
	}{
		{name: "IncompleteRequest", opt: &CourseOptions{}, wantRepos: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &CourseOptions{OrganizationID: 123}, wantRepos: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &CourseOptions{CourseCreator: "meling"}, wantRepos: nil, wantErr: true},

		{name: "CompleteRequest/OrgNotFound", opt: &CourseOptions{OrganizationID: 789, CourseCreator: "meling"}, wantRepos: nil, wantErr: true},              // 789 does not exist
		{name: "CompleteRequest/FooReposAlreadyExists", opt: &CourseOptions{OrganizationID: 123, CourseCreator: "meling"}, wantRepos: nil, wantErr: true},    // foo already has repositories
		{name: "CompleteRequest/CourseCreatorDoesNotExist", opt: &CourseOptions{OrganizationID: 456, CourseCreator: "frank"}, wantRepos: nil, wantErr: true}, // frank is not a member of bar
		{name: "CompleteRequest/CourseBarReposCreated", opt: &CourseOptions{OrganizationID: 456, CourseCreator: "meling"}, wantRepos: wantBarRepos, wantErr: false},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...), WithGroups(g))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"OrganizationID", "CourseCreator"}, tt.opt.OrganizationID, tt.opt.CourseCreator)
		t.Run(name, func(t *testing.T) {
			got, err := s.CreateCourse(context.Background(), tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCourse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// sort repositories to make comparison easier since the order is not guaranteed
			slices.SortFunc(got, func(a, b *Repository) int {
				return strings.Compare(a.Path, b.Path)
			})
			if diff := cmp.Diff(tt.wantRepos, got, cmpopts.IgnoreFields(Repository{}, "ID")); diff != "" {
				t.Errorf("CreateCourse() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMockUpdateEnrollment(t *testing.T) {
	// we need to initialize the groups table (collaborators) to allow updating repository accesses
	g := map[string]map[string][]github.User{
		"foo": {
			"assignments": {}, // needed to allow updating PULL access to the repository
			"meling-labs": {}, // needed to allow updating PUSH access to the repository
		},
		"bar": {
			"assignments": {}, // needed to allow updating PULL access to the repository
			"meling-labs": {}, // needed to allow updating PUSH access to the repository
		},
	}
	// two members; no roles defined yet
	members := []github.Membership{
		{Organization: &ghOrgFoo, User: &meling},
		{Organization: &ghOrgBar, User: &meling},
	}
	wantFooRepo := &Repository{OrgID: 123, Owner: "foo", Path: "meling-labs"}
	wantBarRepo := &Repository{OrgID: 456, Owner: "bar", Path: "meling-labs"}

	tests := []struct {
		name     string
		opt      *UpdateEnrollmentOptions // cannot be nil
		wantRepo *Repository
		wantErr  bool
	}{
		{name: "IncompleteRequest", opt: &UpdateEnrollmentOptions{}, wantRepo: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &UpdateEnrollmentOptions{Status: qf.Enrollment_PENDING}, wantRepo: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &UpdateEnrollmentOptions{Status: qf.Enrollment_STUDENT}, wantRepo: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &UpdateEnrollmentOptions{Status: qf.Enrollment_TEACHER}, wantRepo: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &UpdateEnrollmentOptions{Organization: "foo"}, wantRepo: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &UpdateEnrollmentOptions{Organization: "foo", Status: qf.Enrollment_PENDING}, wantRepo: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &UpdateEnrollmentOptions{Organization: "foo", Status: qf.Enrollment_STUDENT}, wantRepo: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &UpdateEnrollmentOptions{Organization: "foo", Status: qf.Enrollment_TEACHER}, wantRepo: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &UpdateEnrollmentOptions{User: "meling"}, wantRepo: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &UpdateEnrollmentOptions{User: "meling", Status: qf.Enrollment_PENDING}, wantRepo: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &UpdateEnrollmentOptions{User: "meling", Status: qf.Enrollment_STUDENT}, wantRepo: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &UpdateEnrollmentOptions{User: "meling", Status: qf.Enrollment_TEACHER}, wantRepo: nil, wantErr: true},

		{name: "CompleteRequest/OrgNotFound", opt: &UpdateEnrollmentOptions{Organization: "fuzz", User: "meling"}, wantRepo: nil, wantErr: true},
		{name: "CompleteRequest/UserNotFound", opt: &UpdateEnrollmentOptions{Organization: "bar", User: "frank", Status: qf.Enrollment_NONE}, wantRepo: nil, wantErr: true},
		{name: "CompleteRequest/UserNotFound", opt: &UpdateEnrollmentOptions{Organization: "bar", User: "frank", Status: qf.Enrollment_PENDING}, wantRepo: nil, wantErr: true},
		{name: "CompleteRequest/UserNotFound", opt: &UpdateEnrollmentOptions{Organization: "bar", User: "frank", Status: qf.Enrollment_STUDENT}, wantRepo: nil, wantErr: true},
		{name: "CompleteRequest/UserNotFound", opt: &UpdateEnrollmentOptions{Organization: "bar", User: "frank", Status: qf.Enrollment_TEACHER}, wantRepo: nil, wantErr: true},
		{name: "CompleteRequest/None", opt: &UpdateEnrollmentOptions{Organization: "foo", User: "meling", Status: qf.Enrollment_NONE}, wantRepo: nil, wantErr: true},                // not allowed
		{name: "CompleteRequest/Pending", opt: &UpdateEnrollmentOptions{Organization: "foo", User: "meling", Status: qf.Enrollment_PENDING}, wantRepo: nil, wantErr: true},          // not allowed
		{name: "CompleteRequest/Student", opt: &UpdateEnrollmentOptions{Organization: "foo", User: "meling", Status: qf.Enrollment_STUDENT}, wantRepo: wantFooRepo, wantErr: false}, // allowed; returns already created repo (skip creation)
		{name: "CompleteRequest/Teacher", opt: &UpdateEnrollmentOptions{Organization: "foo", User: "meling", Status: qf.Enrollment_TEACHER}, wantRepo: nil, wantErr: false},         // does not return a repo since repo is not created
		{name: "CompleteRequest/Student", opt: &UpdateEnrollmentOptions{Organization: "bar", User: "meling", Status: qf.Enrollment_STUDENT}, wantRepo: wantBarRepo, wantErr: false}, // allowed; returns newly created repo (actual creation)
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...), WithMembers(members...), WithGroups(g))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Organization", "User", "Status"}, tt.opt.Organization, tt.opt.User, tt.opt.Status)
		t.Run(name, func(t *testing.T) {
			got, err := s.UpdateEnrollment(context.Background(), tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateEnrollment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.wantRepo, got, cmpopts.IgnoreFields(Repository{}, "ID")); diff != "" {
				t.Errorf("UpdateEnrollment() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMockRejectEnrollment(t *testing.T) {
	members := []github.Membership{
		{Organization: &ghOrgFoo, User: &meling},
		{Organization: &ghOrgFoo, User: &jostein},
		{Organization: &ghOrgBar, User: &meling},
	}

	tests := []struct {
		name    string
		opt     *RejectEnrollmentOptions // cannot be nil
		wantErr bool
	}{
		{name: "IncompleteRequest", opt: &RejectEnrollmentOptions{}, wantErr: true},
		{name: "IncompleteRequest", opt: &RejectEnrollmentOptions{OrganizationID: 123}, wantErr: true},
		{name: "IncompleteRequest", opt: &RejectEnrollmentOptions{RepositoryID: 1}, wantErr: true},
		{name: "IncompleteRequest", opt: &RejectEnrollmentOptions{User: "meling"}, wantErr: true},
		{name: "IncompleteRequest", opt: &RejectEnrollmentOptions{OrganizationID: 123, RepositoryID: 1}, wantErr: true},
		{name: "IncompleteRequest", opt: &RejectEnrollmentOptions{OrganizationID: 123, User: "meling"}, wantErr: true},
		{name: "IncompleteRequest", opt: &RejectEnrollmentOptions{RepositoryID: 1, User: "meling"}, wantErr: true},

		{name: "CompleteRequest/OrgNotFound", opt: &RejectEnrollmentOptions{OrganizationID: 789, RepositoryID: 1, User: "meling"}, wantErr: true},     // 789 does not exist
		{name: "CompleteRequest/RepoNotFound", opt: &RejectEnrollmentOptions{OrganizationID: 123, RepositoryID: 999, User: "jostein"}, wantErr: true}, // 999 does not exist; note that jostein will be removed from foo
		{name: "CompleteRequest/UserNotFound", opt: &RejectEnrollmentOptions{OrganizationID: 123, RepositoryID: 1, User: "frank"}, wantErr: true},     // frank is not a member of foo
		{name: "CompleteRequest/SuccessfullyRejected", opt: &RejectEnrollmentOptions{OrganizationID: 123, RepositoryID: 4, User: "meling"}, wantErr: false},
		{name: "CompleteRequest/SuccessfullyRejected", opt: &RejectEnrollmentOptions{OrganizationID: 123, RepositoryID: 5, User: "jostein"}, wantErr: true}, // jostein was already removed
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...), WithMembers(members...))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"OrganizationID", "RepositoryID", "User"}, tt.opt.OrganizationID, tt.opt.RepositoryID, tt.opt.User)
		t.Run(name, func(t *testing.T) {
			if err := s.RejectEnrollment(context.Background(), tt.opt); (err != nil) != tt.wantErr {
				t.Errorf("RejectEnrollment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMockDemoteTeacherToStudent(t *testing.T) {
	members := []github.Membership{
		{Organization: &ghOrgFoo, User: &meling, Role: github.String(OrgOwner)},
		{Organization: &ghOrgFoo, User: &jostein, Role: github.String(OrgMember)},
		{Organization: &ghOrgBar, User: &meling, Role: github.String(OrgOwner)},
	}

	tests := []struct {
		name    string
		opt     *UpdateEnrollmentOptions // cannot be nil; the Status field is not used by DemoteTeacherToStudent
		wantErr bool
	}{
		{name: "IncompleteRequest", opt: &UpdateEnrollmentOptions{}, wantErr: true},
		{name: "IncompleteRequest", opt: &UpdateEnrollmentOptions{Organization: "foo"}, wantErr: true},
		{name: "IncompleteRequest", opt: &UpdateEnrollmentOptions{User: "meling"}, wantErr: true},

		{name: "CompleteRequest/OrgNotFound", opt: &UpdateEnrollmentOptions{Organization: "fuzz", User: "meling"}, wantErr: true},
		{name: "CompleteRequest/UserNotFound", opt: &UpdateEnrollmentOptions{Organization: "bar", User: "frank"}, wantErr: true},
		{name: "CompleteRequest/FooStudent", opt: &UpdateEnrollmentOptions{Organization: "foo", User: "jostein"}, wantErr: false}, // jostein is already a student
		{name: "CompleteRequest/FooTeacher", opt: &UpdateEnrollmentOptions{Organization: "foo", User: "meling"}, wantErr: false},  // meling is demoted from teacher to student
		{name: "CompleteRequest/BarTeacher", opt: &UpdateEnrollmentOptions{Organization: "bar", User: "meling"}, wantErr: false},  // meling is demoted from teacher to student
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...), WithMembers(members...))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Organization", "User"}, tt.opt.Organization, tt.opt.User)
		t.Run(name, func(t *testing.T) {
			if err := s.DemoteTeacherToStudent(context.Background(), tt.opt); (err != nil) != tt.wantErr {
				t.Errorf("DemoteTeacherToStudent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
