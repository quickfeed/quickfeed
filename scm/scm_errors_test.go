package scm

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

// Run tests with DEBUG_ERR_MSG=1 to print error messages to view them collectively.
// This is useful when comparing error messages and user errors.
// It is best to run without the -v flag to avoid interleaving with test output, e.g.,
//
//	DEBUG_ERR_MSG=1 go test -run TestErrorGetOrganization
var debugErrMsg = os.Getenv("DEBUG_ERR_MSG") != ""

// IgnoreURLPort returns a cmp.Option that compares URLs in strings ignoring port numbers.
func IgnoreURLPort() cmp.Option {
	return cmp.Options{
		cmp.Transformer("RemoveURLPort", func(s string) string {
			parts := strings.Fields(s)
			for i, part := range parts {
				if strings.HasPrefix(part, "http://") || strings.HasPrefix(part, "https://") {
					u, err := url.Parse(part)
					if err != nil {
						continue // ignore invalid URLs
					}
					// Remove the port number
					u.Host = strings.Split(u.Host, ":")[0]
					parts[i] = u.String()
				}
			}
			return strings.Join(parts, " ")
		}),
		// Default equality check for other types of fields
		cmpopts.EquateEmpty(),
	}
}

func TestErrorGetOrganization(t *testing.T) {
	const (
		wantUserErrAlreadyExist     = "foo: course repositories (info, assignments, tests): already exist"
		wantErrAlreadyExist         = "scm.GetOrganization: " + wantUserErrAlreadyExist
		wantErrPermission           = "scm.GetOrganization: bar: permission denied: meling: not an owner of organization"
		wantErrPermissionMembership = "scm.GetOrganization: bar: permission denied: failed to get membership: GET http://127.0.0.1/orgs/bar/memberships/jostein: 404  []"
		wantUserErrPermission       = "bar: permission denied"
		wantUserErrPrefix           = "failed to get organization"
	)

	orgErrFn := func(opt *OrganizationOptions, suffix string) string {
		base := "scm.GetOrganization: failed to get organization"
		if opt.ID != 0 {
			return fmt.Sprintf("%s by ID: %d: %s", base, opt.ID, suffix)
		}
		if opt.Name != "" {
			return fmt.Sprintf("%s %s: %s", base, opt.Name, suffix)
		}
		return base + suffix
	}

	tests := []struct {
		name        string
		opt         *OrganizationOptions // cannot be nil
		wantErr     string
		wantUserErr string
	}{
		{
			name:        "IncompleteRequest",
			opt:         &OrganizationOptions{},
			wantErr:     orgErrFn(&OrganizationOptions{}, ": missing fields: {ID:0 Name: Username: NewCourse:false}"),
			wantUserErr: wantUserErrPrefix,
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &OrganizationOptions{ID: 789},
			wantErr:     orgErrFn(&OrganizationOptions{ID: 789}, "GET http://127.0.0.1/organizations/789: 404  []"),
			wantUserErr: wantUserErrPrefix + " by ID: 789",
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &OrganizationOptions{ID: 789, NewCourse: true},
			wantErr:     orgErrFn(&OrganizationOptions{ID: 789, NewCourse: true}, "GET http://127.0.0.1/organizations/789: 404  []"),
			wantUserErr: wantUserErrPrefix + " by ID: 789",
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &OrganizationOptions{ID: 789, Username: "meling"},
			wantErr:     orgErrFn(&OrganizationOptions{ID: 789, Username: "meling"}, "GET http://127.0.0.1/organizations/789: 404  []"),
			wantUserErr: wantUserErrPrefix + " by ID: 789",
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &OrganizationOptions{Name: "baz"},
			wantErr:     orgErrFn(&OrganizationOptions{Name: "baz"}, "GET http://127.0.0.1/orgs/baz: 404  []"),
			wantUserErr: wantUserErrPrefix + " baz",
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &OrganizationOptions{Name: "baz", NewCourse: true},
			wantErr:     orgErrFn(&OrganizationOptions{Name: "baz", NewCourse: true}, "GET http://127.0.0.1/orgs/baz: 404  []"),
			wantUserErr: wantUserErrPrefix + " baz",
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &OrganizationOptions{Name: "baz", Username: "meling"},
			wantErr:     orgErrFn(&OrganizationOptions{Name: "baz", Username: "meling"}, "GET http://127.0.0.1/orgs/baz: 404  []"),
			wantUserErr: wantUserErrPrefix + " baz",
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &OrganizationOptions{Name: "baz", NewCourse: true, Username: "meling"},
			wantErr:     orgErrFn(&OrganizationOptions{Name: "baz", NewCourse: true, Username: "meling"}, "GET http://127.0.0.1/orgs/baz: 404  []"),
			wantUserErr: wantUserErrPrefix + " baz",
		},
		{
			name:        "CompleteRequest/AlreadyExists",
			opt:         &OrganizationOptions{ID: 123, NewCourse: true},
			wantErr:     wantErrAlreadyExist,
			wantUserErr: wantUserErrAlreadyExist,
		},
		{
			name:        "CompleteRequest/AlreadyExists",
			opt:         &OrganizationOptions{Name: "foo", NewCourse: true},
			wantErr:     wantErrAlreadyExist,
			wantUserErr: wantUserErrAlreadyExist,
		},
		{
			name:        "CompleteRequest/AlreadyExists",
			opt:         &OrganizationOptions{Name: "foo", NewCourse: true, Username: "meling"},
			wantErr:     wantErrAlreadyExist,
			wantUserErr: wantUserErrAlreadyExist,
		},
		{
			name:        "CompleteRequest/NotMember",
			opt:         &OrganizationOptions{ID: 456, Username: "jostein"},
			wantErr:     wantErrPermissionMembership,
			wantUserErr: wantUserErrPermission,
		},
		{
			name:        "CompleteRequest/NotMember",
			opt:         &OrganizationOptions{Name: "bar", Username: "jostein"},
			wantErr:     wantErrPermissionMembership,
			wantUserErr: wantUserErrPermission,
		},
		{
			name:        "CompleteRequest/OnlyMemberNotOwner",
			opt:         &OrganizationOptions{ID: 456, Username: "meling"},
			wantErr:     wantErrPermission,
			wantUserErr: wantUserErrPermission,
		},
		{
			name:        "CompleteRequest/OnlyMemberNotOwner",
			opt:         &OrganizationOptions{Name: "bar", Username: "meling"},
			wantErr:     wantErrPermission,
			wantUserErr: wantUserErrPermission,
		},
		{
			name:        "CompleteRequest/OnlyMemberNotOwner",
			opt:         &OrganizationOptions{Name: "bar", NewCourse: true, Username: "meling"},
			wantErr:     wantErrPermission,
			wantUserErr: wantUserErrPermission,
		},
		{
			name:        "CompleteRequest/Success",
			opt:         &OrganizationOptions{Name: "foo", Username: "meling"},
			wantErr:     "",
			wantUserErr: "",
		},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...), WithMembers(members...))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"ID", "Name", "Username", "NewCourse"}, tt.opt.ID, tt.opt.Name, tt.opt.Username, tt.opt.NewCourse)
		t.Run(name, func(t *testing.T) {
			_, gotErr := s.GetOrganization(context.Background(), tt.opt)
			chkErrMsg(t, "GetOrganization()", gotErr, tt.wantErr, tt.wantUserErr)
		})
	}
}

func TestErrorCreateCourse(t *testing.T) {
	// we need to members (collaborators) with owner role to allow creating a course with meling as course creator
	members := []github.Membership{
		{Organization: &ghOrgFoo, User: &meling, Role: github.String(OrgOwner)},
		{Organization: &ghOrgBar, User: &jostein, Role: github.String(OrgMember)}, // not allowed to create course
		{Organization: &ghOrgBar, User: &meling, Role: github.String(OrgOwner)},
	}

	tests := []struct {
		name        string
		opt         *CourseOptions // cannot be nil
		wantErr     string
		wantUserErr string
	}{
		{
			name:        "IncompleteRequest",
			opt:         &CourseOptions{},
			wantErr:     "scm.CreateCourse: failed to create course: missing fields: {OrganizationID:0 CourseCreator:}",
			wantUserErr: "failed to create course",
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &CourseOptions{OrganizationID: 789, CourseCreator: "meling"},
			wantErr:     "scm.GetOrganization: failed to get organization by ID: 789: GET http://127.0.0.1/organizations/789: 404  []",
			wantUserErr: "failed to get organization by ID: 789",
		},
		{
			name:        "CompleteRequest/FooReposAlreadyExists",
			opt:         &CourseOptions{OrganizationID: 123, CourseCreator: "meling"},
			wantErr:     "scm.GetOrganization: foo: course repositories (info, assignments, tests): already exist",
			wantUserErr: "foo: course repositories (info, assignments, tests): already exist",
		},
		{
			name:        "CompleteRequest/NotOwner",
			opt:         &CourseOptions{OrganizationID: 456, CourseCreator: "jostein"},
			wantErr:     "scm.GetOrganization: bar: permission denied: jostein: not an owner of organization",
			wantUserErr: "bar: permission denied",
		},
		{
			name:        "CompleteRequest/NotMember",
			opt:         &CourseOptions{OrganizationID: 456, CourseCreator: "lamport"},
			wantErr:     "scm.GetOrganization: bar: permission denied: failed to get membership: GET http://127.0.0.1/orgs/bar/memberships/lamport: 404  []",
			wantUserErr: "bar: permission denied",
		},
		{
			name:        "CompleteRequest/Owner/Success",
			opt:         &CourseOptions{OrganizationID: 456, CourseCreator: "meling"},
			wantErr:     "",
			wantUserErr: "",
		},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...), WithMembers(members...))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"OrganizationID", "CourseCreator"}, tt.opt.OrganizationID, tt.opt.CourseCreator)
		t.Run(name, func(t *testing.T) {
			_, gotErr := s.CreateCourse(context.Background(), tt.opt)
			chkErrMsg(t, "CreateCourse()", gotErr, tt.wantErr, tt.wantUserErr)
		})
	}
}

func TestErrorUpdateEnrollment(t *testing.T) {
	// we need to initialize the groups table (collaborators) to allow updating repository accesses
	g := map[string]map[string][]github.User{
		"foo": {
			"assignments": {}, // needed to allow updating PULL access to the repository
			"meling-labs": {}, // needed to allow updating PUSH access to the repository
		},
		"bar": {
			// "assignments": {}, // commented to trigger error updating bar/assignments to grant PULL access to meling and frank
		},
	}
	// two members; no roles defined yet
	members := []github.Membership{
		{Organization: &ghOrgFoo, User: &meling},
		{Organization: &ghOrgBar, User: &meling},
	}

	wantUserErr := "failed to update enrollment"

	tests := []struct {
		name        string
		opt         *UpdateEnrollmentOptions // cannot be nil
		wantErr     string
		wantUserErr string
	}{
		{
			name:        "IncompleteRequest",
			opt:         &UpdateEnrollmentOptions{},
			wantErr:     "scm.UpdateEnrollment: failed to update enrollment: missing fields: {Organization: User: Status:NONE RefreshToken:}",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &UpdateEnrollmentOptions{Organization: "fuzz", User: "meling"},
			wantErr:     "scm.UpdateEnrollment: failed to update enrollment: scm.GetOrganization: failed to get organization fuzz: GET http://127.0.0.1/orgs/fuzz: 404  []",
			wantUserErr: wantUserErr,
		},

		// user frank does not exist, but is added to s.members in github_mock.go
		{
			name:        "CompleteRequest/IgnoredStatus",
			opt:         &UpdateEnrollmentOptions{Organization: "bar", User: "frank", Status: qf.Enrollment_NONE},
			wantErr:     "scm.UpdateEnrollment: failed to update enrollment: invalid enrollment status: NONE",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/IgnoredStatus",
			opt:         &UpdateEnrollmentOptions{Organization: "bar", User: "frank", Status: qf.Enrollment_PENDING},
			wantErr:     "scm.UpdateEnrollment: failed to update enrollment: invalid enrollment status: PENDING",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/CreateStudRepo",
			opt:         &UpdateEnrollmentOptions{Organization: "bar", User: "frank", Status: qf.Enrollment_STUDENT},
			wantErr:     `scm.UpdateEnrollment: failed to enroll frank as student in bar: failed to add with "pull" access: PUT http://127.0.0.1/repos/bar/assignments/collaborators/frank: 404  []`,
			wantUserErr: "failed to enroll frank as student in bar",
		},
		{
			name:        "CompleteRequest/UpdateToTeacher",
			opt:         &UpdateEnrollmentOptions{Organization: "bar", User: "frank", Status: qf.Enrollment_TEACHER},
			wantErr:     `scm.UpdateEnrollment: failed to enroll frank as teacher in bar: failed to update to "admin": PUT http://127.0.0.1/orgs/bar/memberships/frank: 404  []`,
			wantUserErr: "failed to enroll frank as teacher in bar",
		},

		// user meling already exists in s.members in github_mock.go
		{
			name:        "CompleteRequest/None",
			opt:         &UpdateEnrollmentOptions{Organization: "foo", User: "meling", Status: qf.Enrollment_NONE},
			wantErr:     "scm.UpdateEnrollment: failed to update enrollment: invalid enrollment status: NONE",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/Pending",
			opt:         &UpdateEnrollmentOptions{Organization: "foo", User: "meling", Status: qf.Enrollment_PENDING},
			wantErr:     "scm.UpdateEnrollment: failed to update enrollment: invalid enrollment status: PENDING",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/Student/Success",
			opt:         &UpdateEnrollmentOptions{Organization: "foo", User: "meling", Status: qf.Enrollment_STUDENT},
			wantErr:     "",
			wantUserErr: "",
		},
		{
			name:        "CompleteRequest/Teacher/Success",
			opt:         &UpdateEnrollmentOptions{Organization: "foo", User: "meling", Status: qf.Enrollment_TEACHER},
			wantErr:     "",
			wantUserErr: "",
		},
		{
			name:        "CompleteRequest/Student/Fail",
			opt:         &UpdateEnrollmentOptions{Organization: "bar", User: "meling", Status: qf.Enrollment_STUDENT},
			wantErr:     `scm.UpdateEnrollment: failed to enroll meling as student in bar: failed to add with "pull" access: PUT http://127.0.0.1/repos/bar/assignments/collaborators/meling: 404  []`,
			wantUserErr: "failed to enroll meling as student in bar",
		},

		// The following test succeeds because meling is already a member of the bar organization, and updating the role is allowed.
		// On GitHub, only organization owners can update roles, but we do not enforce this in the mock.
		{
			name:        "CompleteRequest/Teacher/TODO",
			opt:         &UpdateEnrollmentOptions{Organization: "bar", User: "meling", Status: qf.Enrollment_TEACHER},
			wantErr:     "",
			wantUserErr: "",
		},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...), WithMembers(members...), WithGroups(g))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Organization", "User", "Status"}, tt.opt.Organization, tt.opt.User, tt.opt.Status)
		t.Run(name, func(t *testing.T) {
			_, gotErr := s.UpdateEnrollment(context.Background(), tt.opt)
			chkErrMsg(t, "UpdateEnrollment()", gotErr, tt.wantErr, tt.wantUserErr)
		})
	}
}

func TestErrorRejectEnrollment(t *testing.T) {
	members := []github.Membership{
		{Organization: &ghOrgFoo, User: &meling},
		{Organization: &ghOrgFoo, User: &jostein},
		{Organization: &ghOrgBar, User: &meling},
	}
	const userErrPrefix = "failed to reject enrollment"

	tests := []struct {
		name        string
		opt         *RejectEnrollmentOptions // cannot be nil
		wantErr     string
		wantUserErr string
	}{
		{
			name:        "IncompleteRequest",
			opt:         &RejectEnrollmentOptions{},
			wantErr:     "scm.RejectEnrollment: failed to reject enrollment: missing fields: {OrganizationID:0 RepositoryID:0 User:}",
			wantUserErr: userErrPrefix,
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &RejectEnrollmentOptions{OrganizationID: 789, RepositoryID: 1, User: "meling"},
			wantErr:     "scm.RejectEnrollment: failed to reject enrollment for meling: scm.GetOrganization: failed to get organization by ID: 789: GET http://127.0.0.1/organizations/789: 404  []",
			wantUserErr: userErrPrefix + " for meling",
		},
		{
			name:        "CompleteRequest/RepoNotFound",
			opt:         &RejectEnrollmentOptions{OrganizationID: 123, RepositoryID: 999, User: "jostein"},
			wantErr:     "scm.RejectEnrollment: failed to reject enrollment for jostein: scm.deleteRepository: failed to delete repository: failed to get repository 999: GET http://127.0.0.1/repositories/999: 404  []",
			wantUserErr: userErrPrefix + " for jostein",
		},
		{
			name:        "CompleteRequest/UserNotFound",
			opt:         &RejectEnrollmentOptions{OrganizationID: 123, RepositoryID: 1, User: "frank"},
			wantErr:     "scm.RejectEnrollment: failed to reject enrollment for frank: failed to remove user: DELETE http://127.0.0.1/orgs/foo/members/frank: 404  []",
			wantUserErr: userErrPrefix + " for frank",
		},
		{
			name:        "CompleteRequest/UserAlreadyRemoved",
			opt:         &RejectEnrollmentOptions{OrganizationID: 123, RepositoryID: 5, User: "jostein"},
			wantErr:     "scm.RejectEnrollment: failed to reject enrollment for jostein: failed to remove user: DELETE http://127.0.0.1/orgs/foo/members/jostein: 404  []",
			wantUserErr: userErrPrefix + " for jostein",
		},
		{
			name:        "CompleteRequest/Success",
			opt:         &RejectEnrollmentOptions{OrganizationID: 123, RepositoryID: 1, User: "meling"},
			wantErr:     "",
			wantUserErr: "",
		},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...), WithMembers(members...))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"OrganizationID", "RepositoryID", "User"}, tt.opt.OrganizationID, tt.opt.RepositoryID, tt.opt.User)
		t.Run(name, func(t *testing.T) {
			gotErr := s.RejectEnrollment(context.Background(), tt.opt)
			chkErrMsg(t, "RejectEnrollment()", gotErr, tt.wantErr, tt.wantUserErr)
		})
	}
}

func TestErrorDemoteTeacherToStudent(t *testing.T) {
	members := []github.Membership{
		{Organization: &ghOrgFoo, User: &meling, Role: github.String(OrgOwner)},
		{Organization: &ghOrgFoo, User: &jostein, Role: github.String(OrgMember)},
		{Organization: &ghOrgBar, User: &meling, Role: github.String(OrgOwner)},
	}

	tests := []struct {
		name        string
		opt         *UpdateEnrollmentOptions // cannot be nil; the Status field is not used by DemoteTeacherToStudent
		wantErr     string
		wantUserErr string
	}{
		{
			name:        "IncompleteRequest",
			opt:         &UpdateEnrollmentOptions{},
			wantErr:     "scm.DemoteTeacherToStudent: failed to demote teacher to student: missing fields: {Organization: User: Status:NONE RefreshToken:}",
			wantUserErr: "failed to demote teacher to student",
		},

		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &UpdateEnrollmentOptions{Organization: "fuzz", User: "meling"},
			wantErr:     `scm.DemoteTeacherToStudent: failed to demote teacher meling to student in fuzz: failed to update to "member": PUT http://127.0.0.1/orgs/fuzz/memberships/meling: 404  []`,
			wantUserErr: "failed to demote teacher meling to student in fuzz",
		},
		{
			name:        "CompleteRequest/UserNotFound",
			opt:         &UpdateEnrollmentOptions{Organization: "bar", User: "frank"},
			wantErr:     `scm.DemoteTeacherToStudent: failed to demote teacher frank to student in bar: failed to update to "member": PUT http://127.0.0.1/orgs/bar/memberships/frank: 404  []`,
			wantUserErr: "failed to demote teacher frank to student in bar",
		},
		{
			name:        "CompleteRequest/FooStudent/Success",
			opt:         &UpdateEnrollmentOptions{Organization: "foo", User: "jostein"},
			wantErr:     ``,
			wantUserErr: "",
		},
		{
			name:        "CompleteRequest/FooTeacher/Success",
			opt:         &UpdateEnrollmentOptions{Organization: "foo", User: "meling"},
			wantErr:     ``,
			wantUserErr: "",
		},
		{
			name:        "CompleteRequest/BarTeacher/Success",
			opt:         &UpdateEnrollmentOptions{Organization: "bar", User: "meling"},
			wantErr:     ``,
			wantUserErr: "",
		},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...), WithMembers(members...))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Organization", "User"}, tt.opt.Organization, tt.opt.User)
		t.Run(name, func(t *testing.T) {
			gotErr := s.DemoteTeacherToStudent(context.Background(), tt.opt)
			chkErrMsg(t, "DemoteTeacherToStudent()", gotErr, tt.wantErr, tt.wantUserErr)
		})
	}
}

func TestErrorCreateGroup(t *testing.T) {
	const wantUserErr = "failed to create group"

	tests := []struct {
		name        string
		opt         *GroupOptions // cannot be nil
		wantErr     string
		wantUserErr string
	}{
		{
			name:        "IncompleteRequest",
			opt:         &GroupOptions{},
			wantErr:     "scm.CreateGroup: failed to create group: missing fields: {Organization: GroupName: Users:[]}",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &GroupOptions{Organization: "x", GroupName: "sphinx", Users: []string{"meling"}},
			wantErr:     "scm.CreateGroup: failed to create group: scm.GetOrganization: failed to get organization x: GET http://127.0.0.1/orgs/x: 404  []",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/RepoAlreadyExists",
			opt:         &GroupOptions{Organization: "foo", Users: []string{"meling"}, GroupName: "tests"},
			wantErr:     "scm.CreateGroup: foo: repository tests already exist",
			wantUserErr: "foo: repository tests already exist",
		},

		// This test cannot be implemented until the mock can check if a user exists.
		{
			name:        "CompleteRequest/UserDoesNotExists/TODO",
			opt:         &GroupOptions{Organization: "foo", Users: []string{"frank"}, GroupName: "franks-group"},
			wantErr:     "",
			wantUserErr: "",
		},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo, ghOrgBar), WithRepos(repos...), WithGroups(groups))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Organization", "GroupName"}, tt.opt.Organization, tt.opt.GroupName)
		t.Run(name, func(t *testing.T) {
			_, gotErr := s.CreateGroup(context.Background(), tt.opt)
			chkErrMsg(t, "CreateGroup()", gotErr, tt.wantErr, tt.wantUserErr)
		})
	}
}

func TestErrorUpdateGroupMembers(t *testing.T) {
	const wantUserErr = "failed to update group members"

	tests := []struct {
		name        string
		opt         *GroupOptions // cannot be nil
		wantErr     string
		wantUserErr string
	}{
		{
			name:        "IncompleteRequest",
			opt:         &GroupOptions{},
			wantErr:     "scm.UpdateGroupMembers: failed to update group members: missing fields: {Organization: GroupName: Users:[]}",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/RepoNotFound",
			opt:         &GroupOptions{Organization: "foo", GroupName: "a"},
			wantErr:     "scm.UpdateGroupMembers: failed to update group members: failed to get members: GET http://127.0.0.1/repos/foo/a/collaborators: 404  []",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &GroupOptions{Organization: "x", GroupName: "info"},
			wantErr:     "scm.UpdateGroupMembers: failed to update group members: failed to get members: GET http://127.0.0.1/repos/x/info/collaborators: 404  []",
			wantUserErr: wantUserErr,
		},
		// TODO: Add more tests to check error handling when updating group members.
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithGroups(groups))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Organization", "GroupName", "Users"}, tt.opt.Organization, tt.opt.GroupName, tt.opt.Users)
		t.Run(name, func(t *testing.T) {
			gotErr := s.UpdateGroupMembers(context.Background(), tt.opt)
			chkErrMsg(t, "UpdateGroupMembers()", gotErr, tt.wantErr, tt.wantUserErr)
		})
	}
}

func chkErrMsg(t *testing.T, m string, gotErr error, wantErr, wantUserErr string) {
	t.Helper()
	if gotErr == nil {
		if wantErr != "" {
			t.Errorf("%s error = nil, want %q", m, wantErr)
		}
		if wantUserErr != "" {
			t.Errorf("%s user error = nil, want %q", m, wantUserErr)
		}
		return
	}
	if diff := cmp.Diff(wantErr, gotErr.Error(), IgnoreURLPort()); diff != "" {
		t.Log(" got error:", gotErr.Error())
		t.Log("want error:", wantErr)
		t.Errorf("%s error mismatch (-want +got):\n%s", m, diff)
	}
	var userErr *UserError
	if errors.As(gotErr, &userErr) {
		gotUserErr := userErr.Error()
		if debugErrMsg {
			fmt.Printf("%s      error: %v\n", m, gotErr)
			fmt.Printf("%s user error: %v\n", m, gotUserErr)
		}
		if diff := cmp.Diff(wantUserErr, gotUserErr); diff != "" {
			t.Log(" got user error:", gotUserErr)
			t.Log("want user error:", wantUserErr)
			t.Errorf("%s user error mismatch (-want +got):\n%s", m, diff)
		}
	} else {
		if wantUserErr != "" {
			t.Errorf("%s() user error = nil, want %q", m, wantUserErr)
		}
	}
}

func TestErrorCheckSentinel(t *testing.T) {
	const op1 Op = Op("op1")
	const op2 Op = Op("op2")
	tests := []struct {
		name         string
		err          error
		wantSentinel bool
	}{
		{name: "nil", err: nil, wantSentinel: false},
		{name: "simple", err: ErrAlreadyExists, wantSentinel: true},
		{name: "wrapped", err: fmt.Errorf("wrapped: %w", ErrAlreadyExists), wantSentinel: true},
		{name: "wrapped2", err: fmt.Errorf("wrapped: %w", fmt.Errorf("wrapped2: %w", ErrAlreadyExists)), wantSentinel: true},
		{name: "E_func", err: E(op1, M("u1"), ErrAlreadyExists), wantSentinel: true},
		{name: "E_func2", err: E(op2, M("u2"), E(op1, M("u1"), ErrAlreadyExists)), wantSentinel: true},
		{name: "E_func_nested", err: E(op1, M("%s: course repositories %s:", "foo", repoNames), ErrAlreadyExists), wantSentinel: true},
		{name: "E_func_nested2", err: E(op1, M("%s: course repositories %s: %w", "foo", repoNames, ErrAlreadyExists)), wantSentinel: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSentinel := errors.Is(tt.err, ErrAlreadyExists)
			if gotSentinel != tt.wantSentinel {
				t.Errorf("CheckSentinel() = %v, want %v", gotSentinel, tt.wantSentinel)
			}
		})
	}
}

func TestErrorE(t *testing.T) {
	e1 := E(Op("op1"), M("u1"), errors.New("e1"))
	e2 := E(Op("op2"), M("u2"), e1)
	e3 := E(Op("op3"), M("u3"), e2)

	tests := []struct {
		name        string
		err         error
		wantErr     string
		wantUserErr string
	}{
		{
			name:        "E1",
			err:         e1,
			wantErr:     "scm.op1: u1: e1",
			wantUserErr: "u1",
		},
		{
			name:        "E2",
			err:         e2,
			wantErr:     "scm.op2: u2: scm.op1: u1: e1",
			wantUserErr: "u2",
		},
		{
			name:        "E3",
			err:         e3,
			wantErr:     "scm.op3: u3: scm.op2: u2: scm.op1: u1: e1",
			wantUserErr: "u3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.wantErr, tt.err.Error()); diff != "" {
				t.Logf("got err:  %s", tt.err.Error())
				t.Logf("want err: %s", tt.wantErr)
				t.Errorf("%s() error mismatch (-want +got):\n%s", tt.name, diff)
			}

			var userErr *UserError
			hasUserErr := errors.As(tt.err, &userErr)
			if hasUserErr {
				gotUserErr := userErr.Error()
				if diff := cmp.Diff(tt.wantUserErr, gotUserErr); diff != "" {
					t.Logf("got user error:  %s", gotUserErr)
					t.Logf("want user error: %s", tt.wantUserErr)
					t.Errorf("%s() user error mismatch (-want +got):\n%s", tt.name, diff)
				}
			} else {
				if tt.wantUserErr != "" {
					t.Errorf("%s() user error = nil, want %q", tt.name, tt.wantUserErr)
				}
			}
		})
	}
}
