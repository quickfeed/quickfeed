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
	ghOrgBuz    = github.Organization{ID: github.Int64(678), Login: buz.Login}
	ghOrgDat320 = github.Organization{ID: github.Int64(789), Login: github.String("dat320")}
)

// mock repositories for organization foo; bar has no repositories
var repos = []github.Repository{
	{ID: github.Int64(1), Organization: &ghOrgFoo, Name: github.String("info")},
	{ID: github.Int64(2), Organization: &ghOrgFoo, Name: github.String("assignments")},
	{ID: github.Int64(3), Organization: &ghOrgFoo, Name: github.String("tests")},
	{ID: github.Int64(4), Organization: &ghOrgFoo, Name: github.String("meling-labs")},
	{ID: github.Int64(5), Organization: &ghOrgFoo, Name: github.String("josie-labs")},
	{ID: github.Int64(6), Organization: &ghOrgFoo, Name: github.String("groupX")},
	{ID: github.Int64(7), Organization: &ghOrgBar, Name: github.String("groupY")},
	{ID: github.Int64(8), Organization: &ghOrgBar, Name: github.String("groupZ")},
}

var (
	meling  = github.User{Login: github.String("meling")}
	leslie  = github.User{Login: github.String("leslie")}
	lamport = github.User{Login: github.String("lamport")}
	jostein = github.User{Login: github.String("jostein")}
	foo     = github.User{Login: github.String("foo")} // organization (user/owner)
	bar     = github.User{Login: github.String("bar")} // organization (user/owner)
	buz     = github.User{Login: github.String("buz")} // organization (user/owner)
)

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

func TestReplaceArgs(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		args    []any
		want    string
	}{
		{name: "NoArgs", pattern: "foo", args: nil, want: "foo"},
		{name: "OneArg", pattern: "foo/{bar}", args: []any{"baz"}, want: "foo/bar=baz"},
		{name: "TwoArgs", pattern: "foo/{bar}/{baz}", args: []any{123, "qux"}, want: "foo/bar=123/baz=qux"},
		{name: "ThreeArgs", pattern: "foo/{bar}/{baz}/{qux}", args: []any{123, "qux", "quux"}, want: "foo/bar=123/baz=qux/qux=quux"},
		{name: "WithPathElemInBetween", pattern: "foo/{bar}/baz/{qux}", args: []any{"baz", "quux"}, want: "foo/bar=baz/baz/qux=quux"},
		{name: "WithNumbers", pattern: "foo/{bar}/baz/{qux}", args: []any{123, 456}, want: "foo/bar=123/baz/qux=456"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := replaceArgs(tt.pattern, tt.args...); got != tt.want {
				t.Errorf("replaceArgs(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestMockGetOrganization(t *testing.T) {
	orgFoo := &qf.Organization{ScmOrganizationID: 123, ScmOrganizationName: *foo.Login}
	orgBar := &qf.Organization{ScmOrganizationID: 456, ScmOrganizationName: *bar.Login}

	tests := []struct {
		name    string
		opt     *OrganizationOptions // cannot be nil
		wantOrg *qf.Organization
		wantErr bool
	}{
		{name: "IncompleteRequest", opt: &OrganizationOptions{}, wantOrg: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &OrganizationOptions{Username: "meling"}, wantOrg: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &OrganizationOptions{NewCourse: true}, wantOrg: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &OrganizationOptions{NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},

		{name: "CompleteRequest", opt: &OrganizationOptions{ID: 123}, wantOrg: orgFoo, wantErr: false},
		{name: "CompleteRequest", opt: &OrganizationOptions{ID: 456}, wantOrg: orgBar, wantErr: false},
		{name: "CompleteRequest/Missing", opt: &OrganizationOptions{ID: 789}, wantOrg: nil, wantErr: true}, // 789 does not exist

		{name: "CompleteRequest", opt: &OrganizationOptions{ID: 123, NewCourse: true}, wantOrg: nil, wantErr: true},
		{name: "CompleteRequest", opt: &OrganizationOptions{ID: 456, NewCourse: true}, wantOrg: orgBar, wantErr: false},
		{name: "CompleteRequest/Missing", opt: &OrganizationOptions{ID: 789, NewCourse: true}, wantOrg: nil, wantErr: true}, // 789 does not exist

		{name: "CompleteRequest", opt: &OrganizationOptions{ID: 123, Username: "meling"}, wantOrg: orgFoo, wantErr: false},     // meling is owner of foo
		{name: "CompleteRequest", opt: &OrganizationOptions{ID: 456, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is only member of bar, not owner
		{name: "CompleteRequest/Missing", opt: &OrganizationOptions{ID: 789, Username: "meling"}, wantOrg: nil, wantErr: true}, // 789 does not exist

		{name: "CompleteRequest", opt: &OrganizationOptions{ID: 123, NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is owner of foo, but foo is not empty (not new course)
		{name: "CompleteRequest", opt: &OrganizationOptions{ID: 456, NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is only member of bar, not owner
		{name: "CompleteRequest/Missing", opt: &OrganizationOptions{ID: 789, NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true}, // 789 does not exist

		{name: "CompleteRequest", opt: &OrganizationOptions{Name: "foo"}, wantOrg: orgFoo, wantErr: false},
		{name: "CompleteRequest", opt: &OrganizationOptions{Name: "bar"}, wantOrg: orgBar, wantErr: false},
		{name: "CompleteRequest/Missing", opt: &OrganizationOptions{Name: "baz"}, wantOrg: nil, wantErr: true}, // baz does not exist

		{name: "CompleteRequest", opt: &OrganizationOptions{Name: "foo", NewCourse: true}, wantOrg: nil, wantErr: true},
		{name: "CompleteRequest", opt: &OrganizationOptions{Name: "bar", NewCourse: true}, wantOrg: orgBar, wantErr: false},
		{name: "CompleteRequest/Missing", opt: &OrganizationOptions{Name: "baz", NewCourse: true}, wantOrg: nil, wantErr: true}, // baz does not exist

		{name: "CompleteRequest", opt: &OrganizationOptions{Name: "foo", Username: "meling"}, wantOrg: orgFoo, wantErr: false},     // meling is owner of foo
		{name: "CompleteRequest", opt: &OrganizationOptions{Name: "bar", Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is only member of bar, not owner
		{name: "CompleteRequest/Missing", opt: &OrganizationOptions{Name: "baz", Username: "meling"}, wantOrg: nil, wantErr: true}, // baz does not exist

		{name: "CompleteRequest", opt: &OrganizationOptions{Name: "foo", NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is owner of foo
		{name: "CompleteRequest", opt: &OrganizationOptions{Name: "bar", NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is only member of bar, not owner
		{name: "CompleteRequest/Missing", opt: &OrganizationOptions{Name: "baz", NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true}, // baz does not exist
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...), WithMembers(members...))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"ID", "Name", "Username", "NewCourse"}, tt.opt.ID, tt.opt.Name, tt.opt.Username, tt.opt.NewCourse)
		t.Run(name, func(t *testing.T) {
			gotOrg, gotErr := s.GetOrganization(context.Background(), tt.opt)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetOrganization() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.wantOrg, gotOrg, protocmp.Transform()); diff != "" {
				t.Errorf("GetOrganization() mismatch (-want +got):\n%s", diff)
			}
		})
	}

	s = NewMockedGithubSCMClient(qtest.Logger(t), WithMockOrgs())
	for _, course := range qtest.MockCourses {
		name := qtest.Name(course.Name, []string{"ScmOrgID", "ScmOrgName"}, course.ScmOrganizationID, course.ScmOrganizationName)
		t.Run(name, func(t *testing.T) {
			gotOrg, err := s.GetOrganization(context.Background(), &OrganizationOptions{Name: course.ScmOrganizationName})
			if err != nil {
				t.Errorf("GetOrganization() error = %v, want <nil>", err)
			}
			if gotOrg == nil {
				t.Errorf("GetOrganization() = <nil>, want non-nil organization")
			}
			gotOrg, err = s.GetOrganization(context.Background(), &OrganizationOptions{ID: course.ScmOrganizationID})
			if err != nil {
				t.Errorf("GetOrganization() error = %v, want <nil>", err)
			}
			if gotOrg == nil {
				t.Errorf("GetOrganization() = <nil>, want non-nil organization")
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
		{name: "CompleteRequest/NotFound", org: &qf.Organization{ScmOrganizationName: "buz"}, want: []*Repository{}, wantErr: false},
		{name: "CompleteRequest/SixRepos", org: &qf.Organization{ScmOrganizationName: "foo"}, want: []*Repository{
			{ID: 1, OrgID: 123, Path: "info"},
			{ID: 2, OrgID: 123, Path: "assignments"},
			{ID: 3, OrgID: 123, Path: "tests"},
			{ID: 4, OrgID: 123, Path: "meling-labs"},
			{ID: 5, OrgID: 123, Path: "josie-labs"},
			{ID: 6, OrgID: 123, Path: "groupX"},
		}},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar, ghOrgBuz), WithRepos(repos...))
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
			if empty := s.RepositoryIsEmpty(context.Background(), tt.opt); empty != tt.wantEmpty {
				t.Errorf("RepositoryIsEmpty(%+v) = %t, want %t", *tt.opt, empty, tt.wantEmpty)
			}
		})
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
	// we need to members (collaborators) with owner role to allow creating a course with meling as course creator
	members := []github.Membership{
		{Organization: &ghOrgFoo, User: &meling, Role: github.String(OrgOwner)},
		{Organization: &ghOrgBar, User: &jostein, Role: github.String(OrgMember)}, // not allowed to create course
		{Organization: &ghOrgBar, User: &meling, Role: github.String(OrgOwner)},
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

		{name: "CompleteRequest/OrgNotFound", opt: &CourseOptions{OrganizationID: 789, CourseCreator: "meling"}, wantRepos: nil, wantErr: true},           // 789 does not exist
		{name: "CompleteRequest/FooReposAlreadyExists", opt: &CourseOptions{OrganizationID: 123, CourseCreator: "meling"}, wantRepos: nil, wantErr: true}, // foo already has repositories

		{name: "CompleteRequest/CourseBarReposCreated", opt: &CourseOptions{OrganizationID: 456, CourseCreator: "jostein"}, wantRepos: nil, wantErr: true}, // jostein is not owner and cannot create course
		{name: "CompleteRequest/CourseBarReposCreated", opt: &CourseOptions{OrganizationID: 456, CourseCreator: "meling"}, wantRepos: wantBarRepos, wantErr: false},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...), WithMembers(members...))
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
	wantBarFrankRepo := &Repository{OrgID: 456, Owner: "bar", Path: "frank-labs"}

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

		// user frank does not exist, but is added to s.members in github_mock.go
		{name: "CompleteRequest/IgnoredStatus", opt: &UpdateEnrollmentOptions{Organization: "bar", User: "frank", Status: qf.Enrollment_NONE}, wantRepo: nil, wantErr: true},                   // ignored
		{name: "CompleteRequest/IgnoredStatus", opt: &UpdateEnrollmentOptions{Organization: "bar", User: "frank", Status: qf.Enrollment_PENDING}, wantRepo: nil, wantErr: true},                // ignored
		{name: "CompleteRequest/CreateStudRepo", opt: &UpdateEnrollmentOptions{Organization: "bar", User: "frank", Status: qf.Enrollment_STUDENT}, wantRepo: wantBarFrankRepo, wantErr: false}, // allowed; returns newly created repo (actual creation)
		{name: "CompleteRequest/UpdateToTeacher", opt: &UpdateEnrollmentOptions{Organization: "bar", User: "frank", Status: qf.Enrollment_TEACHER}, wantRepo: nil, wantErr: false},             // does not return a repo since repo is not created

		// user meling already exists in s.members in github_mock.go
		{name: "CompleteRequest/None", opt: &UpdateEnrollmentOptions{Organization: "foo", User: "meling", Status: qf.Enrollment_NONE}, wantRepo: nil, wantErr: true},                // ignored
		{name: "CompleteRequest/Pending", opt: &UpdateEnrollmentOptions{Organization: "foo", User: "meling", Status: qf.Enrollment_PENDING}, wantRepo: nil, wantErr: true},          // ignored
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

func TestMockCreateGroup(t *testing.T) {
	tests := []struct {
		name     string
		opt      *GroupOptions
		wantRepo *Repository
		wantErr  bool
	}{
		{name: "IncompleteRequest", opt: &GroupOptions{}, wantRepo: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &GroupOptions{Organization: "foo"}, wantRepo: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &GroupOptions{GroupName: "a"}, wantRepo: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &GroupOptions{Users: []string{"meling"}}, wantRepo: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &GroupOptions{Organization: "foo", Users: []string{"meling"}}, wantRepo: nil, wantErr: true},
		{name: "IncompleteRequest", opt: &GroupOptions{GroupName: "a", Users: []string{"meling"}}, wantRepo: nil, wantErr: true},

		{name: "CompleteRequest/OrgNotFound", opt: &GroupOptions{Organization: "x", GroupName: "sphinx", Users: []string{"meling"}}, wantRepo: nil, wantErr: true},
		{name: "CompleteRequest/RepoAlreadyExists", opt: &GroupOptions{Organization: "foo", Users: []string{"meling"}, GroupName: "info"}, wantRepo: nil, wantErr: true},
		{name: "CompleteRequest/RepoAlreadyExists", opt: &GroupOptions{Organization: "foo", Users: []string{"meling"}, GroupName: "assignments"}, wantRepo: nil, wantErr: true},
		{name: "CompleteRequest/RepoAlreadyExists", opt: &GroupOptions{Organization: "foo", Users: []string{"meling"}, GroupName: "tests"}, wantRepo: nil, wantErr: true},
		{name: "CompleteRequest/RepoAlreadyExists", opt: &GroupOptions{Organization: "foo", Users: []string{"meling"}, GroupName: "meling-labs"}, wantRepo: nil, wantErr: true},
		{name: "CompleteRequest/RepoAlreadyExists", opt: &GroupOptions{Organization: "foo", Users: []string{"meling"}, GroupName: "groupX"}, wantRepo: nil, wantErr: true},

		{name: "CompleteRequest/GroupCreated", opt: &GroupOptions{Organization: "foo", Users: []string{"meling"}, GroupName: "yes-minister"}, wantRepo: &Repository{OrgID: 123, Owner: "foo", Path: "yes-minister"}, wantErr: false},
		{name: "CompleteRequest/GroupCreated", opt: &GroupOptions{Organization: "foo", Users: []string{"leslie", "lamport"}, GroupName: "paxos"}, wantRepo: &Repository{OrgID: 123, Owner: "foo", Path: "paxos"}, wantErr: false},
		{name: "CompleteRequest/GroupCreated", opt: &GroupOptions{Organization: "bar", Users: []string{"jostein", "meling"}, GroupName: "raft"}, wantRepo: &Repository{OrgID: 456, Owner: "bar", Path: "raft"}, wantErr: false},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...), WithGroups(groups))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Organization", "GroupName"}, tt.opt.Organization, tt.opt.GroupName)
		t.Run(name, func(t *testing.T) {
			repo, err := s.CreateGroup(context.Background(), tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantRepo == nil {
				return
			}
			if diff := cmp.Diff(tt.wantRepo, repo, cmpopts.IgnoreFields(Repository{}, "ID")); diff != "" {
				t.Errorf("CreateGroup() mismatch (-want +got):\n%s", diff)
			}
			// verify the state of the groups after the test
			if _, ok := s.groups[tt.opt.Organization][tt.opt.GroupName]; !ok {
				t.Errorf("CreateGroup() group not created")
			}
		})
	}
}

func TestMockUpdateGroupMembers(t *testing.T) {
	push := map[string]bool{"push": true}
	var (
		meling  = github.User{Login: github.String("meling"), Permissions: push}
		leslie  = github.User{Login: github.String("leslie"), Permissions: push}
		lamport = github.User{Login: github.String("lamport"), Permissions: push}
		jostein = github.User{Login: github.String("jostein"), Permissions: push}
	)
	tests := []struct {
		name      string
		opt       *GroupOptions
		wantUsers []github.User
		wantErr   bool
	}{
		{name: "IncompleteRequest", opt: &GroupOptions{}, wantErr: true},
		{name: "IncompleteRequest", opt: &GroupOptions{Organization: "foo"}, wantErr: true},
		{name: "IncompleteRequest", opt: &GroupOptions{GroupName: "a"}, wantErr: true},
		{name: "IncompleteRequest", opt: &GroupOptions{Users: []string{"meling"}}, wantErr: true},
		{name: "IncompleteRequest", opt: &GroupOptions{Organization: "foo", Users: []string{"meling"}}, wantErr: true},
		{name: "IncompleteRequest", opt: &GroupOptions{GroupName: "a", Users: []string{"meling"}}, wantErr: true},

		{name: "CompleteRequest/NotFound", opt: &GroupOptions{Organization: "foo", GroupName: "a"}, wantErr: true},
		{name: "CompleteRequest/NotFound", opt: &GroupOptions{Organization: "x", GroupName: "info"}, wantErr: true},
		{name: "CompleteRequest/NotFound", opt: &GroupOptions{Organization: "foo", GroupName: "a", Users: []string{"meling"}}, wantErr: true},
		{name: "CompleteRequest/NotFound", opt: &GroupOptions{Organization: "x", GroupName: "info", Users: []string{"meling"}}, wantErr: true},

		{name: "CompleteRequest", opt: &GroupOptions{Organization: "foo", GroupName: "info", Users: []string{}}, wantErr: false, wantUsers: []github.User{}},
		{name: "CompleteRequest", opt: &GroupOptions{Organization: "foo", GroupName: "groupX", Users: []string{"meling"}}, wantErr: false, wantUsers: []github.User{meling}},
		{name: "CompleteRequest", opt: &GroupOptions{Organization: "foo", GroupName: "groupX", Users: []string{"meling", "leslie"}}, wantErr: false, wantUsers: []github.User{meling, leslie}},
		{name: "CompleteRequest", opt: &GroupOptions{Organization: "foo", GroupName: "groupX", Users: []string{"meling", "leslie", "lamport"}}, wantErr: false, wantUsers: []github.User{meling, leslie, lamport}},
		{name: "CompleteRequest", opt: &GroupOptions{Organization: "bar", GroupName: "groupY", Users: []string{"leslie", "lamport"}}, wantErr: false, wantUsers: []github.User{leslie, lamport}},
		{name: "CompleteRequest", opt: &GroupOptions{Organization: "bar", GroupName: "groupY", Users: []string{"leslie"}}, wantErr: false, wantUsers: []github.User{leslie}},
		{name: "CompleteRequest", opt: &GroupOptions{Organization: "bar", GroupName: "groupY", Users: []string{}}, wantErr: false, wantUsers: []github.User{}},
		{name: "CompleteRequest", opt: &GroupOptions{Organization: "bar", GroupName: "groupZ", Users: []string{"leslie"}}, wantErr: false, wantUsers: []github.User{leslie}},
		{name: "CompleteRequest", opt: &GroupOptions{Organization: "bar", GroupName: "groupZ", Users: []string{}}, wantErr: false, wantUsers: []github.User{}},
		{name: "CompleteRequest", opt: &GroupOptions{Organization: "bar", GroupName: "groupZ", Users: []string{"leslie", "lamport"}}, wantErr: false, wantUsers: []github.User{leslie, lamport}},
		{name: "CompleteRequest", opt: &GroupOptions{Organization: "bar", GroupName: "groupZ", Users: []string{"jostein"}}, wantErr: false, wantUsers: []github.User{jostein}},
	}
	groups["bar"]["groupY"] = []github.User{leslie}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithGroups(groups))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Organization", "GroupName", "Users"}, tt.opt.Organization, tt.opt.GroupName, tt.opt.Users)
		t.Run(name, func(t *testing.T) {
			if err := s.UpdateGroupMembers(context.Background(), tt.opt); (err != nil) != tt.wantErr {
				t.Errorf("UpdateGroupMembers() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantUsers == nil {
				return
			}
			// verify the state of the groups after the test
			if diff := cmp.Diff(tt.wantUsers, s.groups[tt.opt.Organization][tt.opt.GroupName]); diff != "" {
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

func TestMockDeleteGroup(t *testing.T) {
	tests := []struct {
		name    string
		opt     *RepositoryOptions
		wantErr bool
	}{
		{name: "IncompleteRequest", opt: &RepositoryOptions{}, wantErr: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Owner: "foo"}, wantErr: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Path: "info"}, wantErr: true},

		{name: "CompleteRequest/NotFound", opt: &RepositoryOptions{Owner: "foo", Path: "bar"}, wantErr: true},
		{name: "CompleteRequest/NotFound", opt: &RepositoryOptions{Owner: "bar", Path: "foo"}, wantErr: true},
		{name: "CompleteRequest/NotFound", opt: &RepositoryOptions{ID: 432}, wantErr: true}, // Repo ID 432 does not exist

		{name: "CompleteRequest/GroupDeleted", opt: &RepositoryOptions{Owner: "foo", Path: "groupX"}, wantErr: false}, // ID 6
		{name: "CompleteRequest/GroupDeleted", opt: &RepositoryOptions{ID: 5}, wantErr: false},                        // ID 5 is josie-labs
		{name: "CompleteRequest/GroupDeleted", opt: &RepositoryOptions{Owner: "bar", Path: "groupY"}, wantErr: false}, // ID 7
		{name: "CompleteRequest/GroupDeleted", opt: &RepositoryOptions{ID: 8}, wantErr: false},                        // ID 8 is groupZ

		{name: "CompleteRequest/AlreadyDeleted", opt: &RepositoryOptions{ID: 6}, wantErr: true},                        // ID 6 already deleted
		{name: "CompleteRequest/AlreadyDeleted", opt: &RepositoryOptions{ID: 7}, wantErr: true},                        // ID 7 already deleted
		{name: "CompleteRequest/AlreadyDeleted", opt: &RepositoryOptions{Owner: "bar", Path: "groupZ"}, wantErr: true}, // ID 8 already deleted
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...), WithGroups(groups))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Owner", "Path"}, tt.opt.Owner, tt.opt.Path)
		t.Run(name, func(t *testing.T) {
			if err := s.DeleteGroup(context.Background(), tt.opt); (err != nil) != tt.wantErr {
				t.Errorf("DeleteGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
			// verify the state of the groups after the test
			if _, ok := s.groups[tt.opt.Owner][tt.opt.Path]; ok {
				t.Errorf("DeleteGroup() group not deleted")
			}
		})
	}
}
