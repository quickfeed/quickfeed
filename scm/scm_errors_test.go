package scm

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

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
	const wantErrPrefix = "scm.GetOrganization: failed to get organization: "
	const wantErrPrefix2 = "scm.GetOrganization: foo: course repositories already exist"
	wantErrPrefix3 := "scm.GetOrganization: bar/meling: " + ErrNotOwner.Error()
	const wantUserErrPrefix = "failed to get organization"
	const wantUserErrSuffix = ": permission denied for "
	const wantUserErrPrefix2 = "course repositories (info, assignments, tests) already exist for "

	tests := []struct {
		name        string
		opt         *OrganizationOptions // cannot be nil
		wantErr     string
		wantUserErr string
	}{
		{
			name:        "IncompleteRequest",
			opt:         &OrganizationOptions{},
			wantErr:     "scm.GetOrganization: missing fields: {ID:0 Name: Username: NewCourse:false}",
			wantUserErr: wantUserErrPrefix,
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &OrganizationOptions{ID: 789},
			wantErr:     wantErrPrefix + "GET http://127.0.0.1:61390/organizations/789: 404  []",
			wantUserErr: wantUserErrPrefix + " by ID: 789",
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &OrganizationOptions{ID: 789, NewCourse: true},
			wantErr:     wantErrPrefix + "GET http://127.0.0.1:61730/organizations/789: 404  []",
			wantUserErr: wantUserErrPrefix + " by ID: 789",
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &OrganizationOptions{ID: 789, Username: "meling"},
			wantErr:     wantErrPrefix + "GET http://127.0.0.1/organizations/789: 404  []",
			wantUserErr: wantUserErrPrefix + " by ID: 789",
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &OrganizationOptions{Name: "baz"},
			wantErr:     wantErrPrefix + "GET http://127.0.0.1:49915/orgs/baz: 404  []",
			wantUserErr: wantUserErrPrefix + " baz",
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &OrganizationOptions{Name: "baz", NewCourse: true},
			wantErr:     wantErrPrefix + "GET http://127.0.0.1:49915/orgs/baz: 404  []",
			wantUserErr: wantUserErrPrefix + " baz",
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &OrganizationOptions{Name: "baz", Username: "meling"},
			wantErr:     wantErrPrefix + "GET http://127.0.0.1:49915/orgs/baz: 404  []",
			wantUserErr: wantUserErrPrefix + " baz",
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &OrganizationOptions{Name: "baz", NewCourse: true, Username: "meling"},
			wantErr:     wantErrPrefix + "GET http://127.0.0.1:51062/orgs/baz: 404  []",
			wantUserErr: wantUserErrPrefix + " baz",
		},
		{
			name:        "CompleteRequest/AlreadyExists",
			opt:         &OrganizationOptions{ID: 123, NewCourse: true},
			wantErr:     wantErrPrefix2,
			wantUserErr: wantUserErrPrefix2 + "foo",
		},
		{
			name:        "CompleteRequest/AlreadyExists",
			opt:         &OrganizationOptions{Name: "foo", NewCourse: true},
			wantErr:     wantErrPrefix2,
			wantUserErr: wantUserErrPrefix2 + "foo",
		},
		{
			name:        "CompleteRequest/AlreadyExists",
			opt:         &OrganizationOptions{Name: "foo", NewCourse: true, Username: "meling"},
			wantErr:     wantErrPrefix2,
			wantUserErr: wantUserErrPrefix2 + "foo",
		},
		{
			name:        "CompleteRequest/NotMember",
			opt:         &OrganizationOptions{ID: 456, Username: "jostein"},
			wantErr:     "scm.GetOrganization: failed to get membership: GET http://127.0.0.1:62169/orgs/bar/memberships/jostein: 404  []",
			wantUserErr: "bar" + wantUserErrSuffix + "jostein",
		},
		{
			name:        "CompleteRequest/NotMember",
			opt:         &OrganizationOptions{Name: "bar", Username: "jostein"},
			wantErr:     "scm.GetOrganization: failed to get membership: GET http://127.0.0.1:62169/orgs/bar/memberships/jostein: 404  []",
			wantUserErr: "bar" + wantUserErrSuffix + "jostein",
		},
		{
			name:        "CompleteRequest/OnlyMemberNotOwner",
			opt:         &OrganizationOptions{ID: 456, Username: "meling"},
			wantErr:     wantErrPrefix3,
			wantUserErr: "bar" + wantUserErrSuffix + "meling",
		},
		{
			name:        "CompleteRequest/OnlyMemberNotOwner",
			opt:         &OrganizationOptions{Name: "bar", Username: "meling"},
			wantErr:     wantErrPrefix3,
			wantUserErr: "bar" + wantUserErrSuffix + "meling",
		},
		{
			name:        "CompleteRequest/OnlyMemberNotOwner",
			opt:         &OrganizationOptions{Name: "bar", NewCourse: true, Username: "meling"},
			wantErr:     wantErrPrefix3,
			wantUserErr: "bar" + wantUserErrSuffix + "meling",
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
			if gotErr == nil {
				if tt.wantErr != "" {
					t.Errorf("GetOrganization() error = nil, want %q", tt.wantErr)
				}
				if tt.wantUserErr != "" {
					t.Errorf("GetOrganization() user error = nil, want %q", tt.wantUserErr)
				}
				return
			}
			if diff := cmp.Diff(tt.wantErr, gotErr.Error(), IgnoreURLPort()); diff != "" {
				t.Logf(gotErr.Error())
				t.Errorf("GetOrganization() error mismatch (-want +got):\n%s", diff)
			}
			var userErr *UserError
			if errors.As(gotErr, &userErr) {
				gotUserErr := userErr.Error()
				if diff := cmp.Diff(tt.wantUserErr, gotUserErr); diff != "" {
					t.Logf(gotUserErr)
					t.Errorf("GetOrganization() user error mismatch (-want +got):\n%s", diff)
				}
			}
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
			wantErr:     "scm.CreateCourse: missing fields: {OrganizationID:0 CourseCreator:}",
			wantUserErr: "failed to create course",
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &CourseOptions{OrganizationID: 789, CourseCreator: "meling"},
			wantErr:     "scm.GetOrganization: failed to get organization: GET http://127.0.0.1/organizations/789: 404  []",
			wantUserErr: "failed to get organization by ID: 789",
		},
		{
			name:        "CompleteRequest/FooReposAlreadyExists",
			opt:         &CourseOptions{OrganizationID: 123, CourseCreator: "meling"},
			wantErr:     "scm.GetOrganization: foo: course repositories already exist",
			wantUserErr: "course repositories (info, assignments, tests) already exist for foo",
		},
		{
			name:        "CompleteRequest/NotOwner",
			opt:         &CourseOptions{OrganizationID: 456, CourseCreator: "jostein"},
			wantErr:     "scm.GetOrganization: bar/jostein: " + ErrNotOwner.Error(),
			wantUserErr: "bar: permission denied for jostein",
		},
		{
			name:        "CompleteRequest/NotMember",
			opt:         &CourseOptions{OrganizationID: 456, CourseCreator: "lamport"},
			wantErr:     "scm.GetOrganization: failed to get membership: GET http://127.0.0.1:57346/orgs/bar/memberships/lamport: 404  []",
			wantUserErr: "bar: permission denied for lamport",
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
			if gotErr == nil {
				if tt.wantErr != "" {
					t.Errorf("CreateCourse() error = nil, want %q", tt.wantErr)
				}
				if tt.wantUserErr != "" {
					t.Errorf("CreateCourse() user error = nil, want %q", tt.wantUserErr)
				}
				return
			}
			if diff := cmp.Diff(tt.wantErr, gotErr.Error(), IgnoreURLPort()); diff != "" {
				t.Logf(gotErr.Error())
				t.Errorf("CreateCourse() error mismatch (-want +got):\n%s", diff)
			}
			var userErr *UserError
			if errors.As(gotErr, &userErr) {
				gotUserErr := userErr.Error()
				if diff := cmp.Diff(tt.wantUserErr, gotUserErr); diff != "" {
					t.Logf(gotUserErr)
					t.Errorf("GetOrganization() user error mismatch (-want +got):\n%s", diff)
				}
			}
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
			wantErr:     "scm.UpdateEnrollment: missing fields: {Organization: User: Status:NONE}",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &UpdateEnrollmentOptions{Organization: "fuzz", User: "meling"},
			wantErr:     "scm.UpdateEnrollment: failed to update enrollment: scm.GetOrganization: failed to get organization: GET http://127.0.0.1:50580/orgs/fuzz: 404  []",
			wantUserErr: wantUserErr,
		},

		// user frank does not exist, but is added to s.members in github_mock.go
		{
			name:        "CompleteRequest/IgnoredStatus",
			opt:         &UpdateEnrollmentOptions{Organization: "bar", User: "frank", Status: qf.Enrollment_NONE},
			wantErr:     "scm.UpdateEnrollment: invalid enrollment status: NONE",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/IgnoredStatus",
			opt:         &UpdateEnrollmentOptions{Organization: "bar", User: "frank", Status: qf.Enrollment_PENDING},
			wantErr:     "scm.UpdateEnrollment: invalid enrollment status: PENDING",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/CreateStudRepo",
			opt:         &UpdateEnrollmentOptions{Organization: "bar", User: "frank", Status: qf.Enrollment_STUDENT},
			wantErr:     `scm.UpdateEnrollment: failed to add frank with "pull" access to bar/assignments: PUT http://127.0.0.1:55626/repos/bar/assignments/collaborators/frank: 404  []`,
			wantUserErr: "failed to enroll frank as student in bar",
		},
		{
			name:        "CompleteRequest/UpdateToTeacher",
			opt:         &UpdateEnrollmentOptions{Organization: "bar", User: "frank", Status: qf.Enrollment_TEACHER},
			wantErr:     `scm.UpdateEnrollment: failed to update frank's role to "admin" in organization bar: PUT http://127.0.0.1:55626/orgs/bar/memberships/frank: 404  []`,
			wantUserErr: "failed to enroll frank as teacher in bar",
		},

		// user meling already exists in s.members in github_mock.go
		{
			name:        "CompleteRequest/None",
			opt:         &UpdateEnrollmentOptions{Organization: "foo", User: "meling", Status: qf.Enrollment_NONE},
			wantErr:     "scm.UpdateEnrollment: invalid enrollment status: NONE",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/Pending",
			opt:         &UpdateEnrollmentOptions{Organization: "foo", User: "meling", Status: qf.Enrollment_PENDING},
			wantErr:     "scm.UpdateEnrollment: invalid enrollment status: PENDING",
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
			wantErr:     `scm.UpdateEnrollment: failed to add meling with "pull" access to bar/assignments: PUT http://127.0.0.1:55783/repos/bar/assignments/collaborators/meling: 404  []`,
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
			if gotErr == nil {
				if tt.wantErr != "" {
					t.Errorf("UpdateEnrollment() error = nil, want %q", tt.wantErr)
				}
				if tt.wantUserErr != "" {
					t.Errorf("UpdateEnrollment() user error = nil, want %q", tt.wantUserErr)
				}
				return
			}
			if diff := cmp.Diff(tt.wantErr, gotErr.Error(), IgnoreURLPort()); diff != "" {
				t.Logf(gotErr.Error())
				t.Errorf("UpdateEnrollment() error mismatch (-want +got):\n%s", diff)
			}
			var userErr *UserError
			if errors.As(gotErr, &userErr) {
				gotUserErr := userErr.Error()
				if diff := cmp.Diff(tt.wantUserErr, gotUserErr); diff != "" {
					t.Logf(gotUserErr)
					t.Errorf("UpdateEnrollment() user error mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestErrorRejectEnrollment(t *testing.T) {
	members := []github.Membership{
		{Organization: &ghOrgFoo, User: &meling},
		{Organization: &ghOrgFoo, User: &jostein},
		{Organization: &ghOrgBar, User: &meling},
	}
	const userErrPrefix = "failed to reject enrollment for "

	tests := []struct {
		name        string
		opt         *RejectEnrollmentOptions // cannot be nil
		wantErr     string
		wantUserErr string
	}{
		{
			name:        "IncompleteRequest",
			opt:         &RejectEnrollmentOptions{},
			wantErr:     "scm.RejectEnrollment: missing fields: {OrganizationID:0 RepositoryID:0 User:}",
			wantUserErr: userErrPrefix,
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &RejectEnrollmentOptions{OrganizationID: 789, RepositoryID: 1, User: "meling"},
			wantErr:     "scm.RejectEnrollment: failed to reject enrollment for meling: scm.GetOrganization: failed to get organization: GET http://127.0.0.1/organizations/789: 404  []",
			wantUserErr: userErrPrefix + "meling",
		},
		{
			name:        "CompleteRequest/RepoNotFound",
			opt:         &RejectEnrollmentOptions{OrganizationID: 123, RepositoryID: 999, User: "jostein"},
			wantErr:     "scm.RejectEnrollment: failed to reject enrollment for jostein: scm.deleteRepository: failed to get repository 999: GET http://127.0.0.1/repositories/999: 404  []",
			wantUserErr: userErrPrefix + "jostein",
		},
		{
			name:        "CompleteRequest/UserNotFound",
			opt:         &RejectEnrollmentOptions{OrganizationID: 123, RepositoryID: 1, User: "frank"},
			wantErr:     "scm.RejectEnrollment: failed to remove user: DELETE http://127.0.0.1/orgs/foo/members/frank: 404  []",
			wantUserErr: userErrPrefix + "frank",
		},
		{
			name:        "CompleteRequest/UserAlreadyRemoved",
			opt:         &RejectEnrollmentOptions{OrganizationID: 123, RepositoryID: 5, User: "jostein"},
			wantErr:     "scm.RejectEnrollment: failed to remove user: DELETE http://127.0.0.1/orgs/foo/members/jostein: 404  []",
			wantUserErr: userErrPrefix + "jostein",
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
			if gotErr == nil {
				if tt.wantErr != "" {
					t.Errorf("RejectEnrollment() error = nil, want %q", tt.wantErr)
				}
				return
			}
			if diff := cmp.Diff(tt.wantErr, gotErr.Error(), IgnoreURLPort()); diff != "" {
				t.Logf(gotErr.Error())
				t.Errorf("RejectEnrollment() error mismatch (-want +got):\n%s", diff)
			}
			var userErr *UserError
			if errors.As(gotErr, &userErr) {
				gotUserErr := userErr.Error()
				if diff := cmp.Diff(tt.wantUserErr, gotUserErr); diff != "" {
					t.Logf(gotUserErr)
					t.Errorf("RejectEnrollment() user error mismatch (-want +got):\n%s", diff)
				}
			}
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
			wantErr:     "scm.DemoteTeacherToStudent: missing fields: {Organization: User: Status:NONE}",
			wantUserErr: "failed to demote teacher to student",
		},

		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &UpdateEnrollmentOptions{Organization: "fuzz", User: "meling"},
			wantErr:     `scm.DemoteTeacherToStudent: failed to update meling's role to "member" in organization fuzz: PUT http://127.0.0.1:59502/orgs/fuzz/memberships/meling: 404  []`,
			wantUserErr: "failed to demote teacher meling to student in fuzz",
		},
		{
			name:        "CompleteRequest/UserNotFound",
			opt:         &UpdateEnrollmentOptions{Organization: "bar", User: "frank"},
			wantErr:     `scm.DemoteTeacherToStudent: failed to update frank's role to "member" in organization bar: PUT http://127.0.0.1:59685/orgs/bar/memberships/frank: 404  []`,
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
			if gotErr == nil {
				if tt.wantErr != "" {
					t.Errorf("DemoteTeacherToStudent() error = nil, want %q", tt.wantErr)
				}
				return
			}
			if diff := cmp.Diff(tt.wantErr, gotErr.Error(), IgnoreURLPort()); diff != "" {
				t.Logf(gotErr.Error())
				t.Errorf("DemoteTeacherToStudent() error mismatch (-want +got):\n%s", diff)
			}
			var userErr *UserError
			if errors.As(gotErr, &userErr) {
				gotUserErr := userErr.Error()
				if diff := cmp.Diff(tt.wantUserErr, gotUserErr); diff != "" {
					t.Logf(gotUserErr)
					t.Errorf("DemoteTeacherToStudent() user error mismatch (-want +got):\n%s", diff)
				}
			}
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
			wantErr:     "scm.CreateGroup: missing fields: {Organization: GroupName: Users:[]}",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/OrgNotFound",
			opt:         &GroupOptions{Organization: "x", GroupName: "sphinx", Users: []string{"meling"}},
			wantErr:     "scm.CreateGroup: failed to create group: scm.GetOrganization: failed to get organization: GET http://127.0.0.1:61071/orgs/x: 404  []",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/RepoAlreadyExists",
			opt:         &GroupOptions{Organization: "foo", Users: []string{"meling"}, GroupName: "tests"},
			wantErr:     "scm.CreateGroup: repository foo/tests already exists: " + ErrAlreadyExists.Error(),
			wantUserErr: wantUserErr,
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
			if gotErr == nil {
				if tt.wantErr != "" {
					t.Errorf("CreateGroup() error = nil, want %q", tt.wantErr)
				}
				if tt.wantUserErr != "" {
					t.Errorf("CreateGroup() user error = nil, want %q", tt.wantUserErr)
				}
				return
			}
			if diff := cmp.Diff(tt.wantErr, gotErr.Error(), IgnoreURLPort()); diff != "" {
				t.Logf(gotErr.Error())
				t.Errorf("CreateGroup() error mismatch (-want +got):\n%s", diff)
			}
			var userErr *UserError
			if errors.As(gotErr, &userErr) {
				gotUserErr := userErr.Error()
				if diff := cmp.Diff(tt.wantUserErr, gotUserErr); diff != "" {
					t.Logf(gotUserErr)
					t.Errorf("CreateGroup() user error mismatch (-want +got):\n%s", diff)
				}
			}
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
			wantErr:     "scm.UpdateGroupMembers: missing fields: {Organization: GroupName: Users:[]}",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/NotFound",
			opt:         &GroupOptions{Organization: "foo", GroupName: "a"},
			wantErr:     "scm.UpdateGroupMembers: failed to get members for foo/a: GET http://127.0.0.1:63851/repos/foo/a/collaborators: 404  []",
			wantUserErr: wantUserErr,
		},
		{
			name:        "CompleteRequest/NotFound",
			opt:         &GroupOptions{Organization: "x", GroupName: "info"},
			wantErr:     "scm.UpdateGroupMembers: failed to get members for x/info: GET http://127.0.0.1:63851/repos/x/info/collaborators: 404  []",
			wantUserErr: wantUserErr,
		},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithGroups(groups))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Organization", "GroupName", "Users"}, tt.opt.Organization, tt.opt.GroupName, tt.opt.Users)
		t.Run(name, func(t *testing.T) {
			gotErr := s.UpdateGroupMembers(context.Background(), tt.opt)
			if gotErr == nil {
				if tt.wantErr != "" {
					t.Errorf("UpdateGroupMembers() error = nil, want %q", tt.wantErr)
				}
				if tt.wantUserErr != "" {
					t.Errorf("UpdateGroupMembers() user error = nil, want %q", tt.wantUserErr)
				}
				return
			}
			if diff := cmp.Diff(tt.wantErr, gotErr.Error(), IgnoreURLPort()); diff != "" {
				t.Logf(gotErr.Error())
				t.Errorf("UpdateGroupMembers() error mismatch (-want +got):\n%s", diff)
			}
			var userErr *UserError
			if errors.As(gotErr, &userErr) {
				gotUserErr := userErr.Error()
				if diff := cmp.Diff(tt.wantUserErr, gotUserErr); diff != "" {
					t.Logf(gotUserErr)
					t.Errorf("UpdateGroupMembers() user error mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestErrorE(t *testing.T) {
	e1 := E(Op("GetOrganization"), M("organization not found"), errors.New("organization not found"))
	e2 := E(Op("GetRepository"), M("repository not found"), e1)
	e3 := E(Op("GetUser"), M("user not found"), e2)

	tests := []struct {
		name        string
		err         error
		wantErr     string
		wantUserErr string
	}{
		{
			name:        "E1",
			err:         e1,
			wantErr:     "scm.GetOrganization: organization not found",
			wantUserErr: "organization not found",
		},
		{
			name:        "E2",
			err:         e2,
			wantErr:     "scm.GetRepository: repository not found: scm.GetOrganization: organization not found",
			wantUserErr: "repository not found",
		},
		{
			name:        "E3",
			err:         e3,
			wantErr:     "scm.GetUser: user not found: scm.GetRepository: repository not found: scm.GetOrganization: organization not found",
			wantUserErr: "user not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.wantErr, tt.err.Error()); diff != "" {
				t.Errorf("%s() error mismatch (-want +got):\n%s", tt.name, diff)
			}
			var userErr *UserError
			if errors.As(tt.err, &userErr) {
				gotUserErr := userErr.Error()
				if diff := cmp.Diff(tt.wantUserErr, gotUserErr); diff != "" {
					t.Errorf("%s() user error mismatch (-want +got):\n%s", tt.name, diff)
				}
			}
		})
	}
}
