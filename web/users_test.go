package web_test

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetUserExpectUnknownUser(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := web.NewMockClient(t, db, scm.WithMockOrgs())
	_, err := client.GetUser(t.Context(), &qf.Void{})
	qtest.CheckError(t, err, connect.NewError(connect.CodeNotFound, errors.New("unknown user")))
}

func TestGetUsers(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := web.NewMockClient(t, db, scm.WithMockOrgs())
	ctx := t.Context()

	unexpectedUsers, err := client.GetUsers(ctx, &qf.Void{})
	if err == nil && unexpectedUsers != nil && len(unexpectedUsers.GetUsers()) > 0 {
		t.Fatalf("found unexpected users %+v", unexpectedUsers)
	}

	admin := qtest.CreateFakeUser(t, db)
	user2 := qtest.CreateFakeUser(t, db)

	foundUsers, err := client.GetUsers(ctx, &qf.Void{})
	if err != nil {
		t.Fatal(err)
	}

	wantUsers := make([]*qf.User, 0)
	wantUsers = append(wantUsers, admin, user2)
	gotUsers := foundUsers.GetUsers()
	if diff := cmp.Diff(wantUsers, gotUsers, protocmp.Transform()); diff != "" {
		t.Errorf("GetUsers() mismatch (-wantUsers +gotUsers):\n%s", diff)
	}
}

func TestGetEnrollmentsByCourse(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := web.NewMockClient(t, db, scm.WithMockOrgs())

	var users []*qf.User
	for range 10 {
		user := qtest.CreateFakeUser(t, db)
		users = append(users, user)
	}
	admin := users[0]
	for _, course := range qtest.MockCourses {
		qtest.CreateCourse(t, db, admin, course)
	}

	// users to enroll in course DAT520 Distributed Systems
	// (excluding admin because admin is enrolled on creation)
	wantUsers := users[0:6]
	for i, user := range wantUsers {
		if i == 0 {
			// skip enrolling admin as student
			continue
		}
		qtest.EnrollStudent(t, db, user, qtest.MockCourses[0])
	}

	// users to enroll in course DAT320 Operating Systems
	// (excluding admin because admin is enrolled on creation)
	osUsers := users[3:7]
	for _, user := range osUsers {
		qtest.EnrollStudent(t, db, user, qtest.MockCourses[1])
	}

	request := &qf.EnrollmentRequest{
		FetchMode: &qf.EnrollmentRequest_CourseID{
			CourseID: qtest.MockCourses[0].GetID(),
		},
	}

	gotEnrollments, err := client.GetEnrollments(t.Context(), request)
	if err != nil {
		t.Error(err)
	}
	var gotUsers []*qf.User
	for _, e := range gotEnrollments.GetEnrollments() {
		gotUsers = append(gotUsers, e.GetUser())
	}
	if diff := cmp.Diff(wantUsers, gotUsers, protocmp.Transform()); diff != "" {
		t.Errorf("GetEnrollmentsByCourse() mismatch (-wantUsers +gotUsers):\n%s", diff)
	}
}

func TestUpdateUser(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors())

	firstAdminUser := qtest.CreateFakeUser(t, db)
	nonAdminUser := qtest.CreateFakeUser(t, db)

	firstAdminCtx := client.Context(t, firstAdminUser)

	// we want to update nonAdminUser to become admin
	nonAdminUser.IsAdmin = true
	err := db.UpdateUser(nonAdminUser)
	if err != nil {
		t.Fatal(err)
	}

	// we expect the nonAdminUser to now be admin
	admin, err := db.GetUser(nonAdminUser.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if !admin.GetIsAdmin() {
		t.Error("expected nonAdminUser to have become admin")
	}

	nameChangeRequest := &qf.User{
		ID:        nonAdminUser.GetID(),
		IsAdmin:   nonAdminUser.GetIsAdmin(),
		Name:      "Scrooge McDuck",
		StudentID: "99",
		Email:     "test@test.com",
		AvatarURL: "www.hello.com",
	}

	_, err = client.UpdateUser(firstAdminCtx, nameChangeRequest)
	if err != nil {
		t.Error(err)
	}
	gotUser, err := db.GetUser(nonAdminUser.GetID())
	if err != nil {
		t.Fatal(err)
	}
	wantUser := &qf.User{
		ID:           gotUser.GetID(),
		Name:         "Scrooge McDuck",
		IsAdmin:      true,
		StudentID:    "99",
		Email:        "test@test.com",
		AvatarURL:    "www.hello.com",
		RefreshToken: nonAdminUser.GetRefreshToken(),
		ScmRemoteID:  nonAdminUser.GetScmRemoteID(),
	}
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateUser() mismatch (-wantUser +gotUser):\n%s", diff)
	}
}

func TestUpdateUserFailures(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors())

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "admin", Login: "admin"})
	if !admin.GetIsAdmin() {
		t.Fatalf("expected user %v to be admin", admin)
	}
	user := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "user", Login: "user"})
	if user.GetIsAdmin() {
		t.Fatalf("expected user %v to be non-admin", user)
	}
	userCtx := client.Context(t, user)
	adminCtx := client.Context(t, admin)
	tests := []struct {
		name     string
		ctx      context.Context
		req      *qf.User
		wantUser *qf.User
		wantErr  bool
	}{
		{
			name: "user demotes admin, must fail",
			ctx:  userCtx,
			req: &qf.User{
				ID:        admin.GetID(),
				IsAdmin:   false,
				Name:      admin.GetName(),
				Email:     admin.GetEmail(),
				StudentID: admin.GetStudentID(),
				AvatarURL: admin.GetAvatarURL(),
			},
			wantErr: true,
		},
		{
			name: "user promotes self to admin, must fail",
			ctx:  userCtx,
			req: &qf.User{
				ID:        user.GetID(),
				Name:      user.GetName(),
				Email:     user.GetEmail(),
				StudentID: user.GetStudentID(),
				AvatarURL: user.GetAvatarURL(),
				IsAdmin:   true,
			},
			wantErr: true,
		},
		{
			name: "admin changes own name, must pass",
			ctx:  adminCtx,
			req: &qf.User{
				ID:   admin.GetID(),
				Name: "super user",
			},
			wantUser: &qf.User{
				ID:           admin.GetID(),
				IsAdmin:      true,
				Login:        admin.GetLogin(),
				Name:         "super user",
				Email:        admin.GetEmail(),
				StudentID:    admin.GetStudentID(),
				RefreshToken: admin.GetRefreshToken(),
				ScmRemoteID:  admin.GetScmRemoteID(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// UpdateUser returns void, so we cannot check that the user was updated
			_, err := client.UpdateUser(tt.ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
			}
			if !tt.wantErr {
				// Instead (for success cases), get all users and check that the user was updated
				users, err := client.GetUsers(tt.ctx, &qf.Void{})
				if err != nil {
					t.Fatal(err)
				}
				for _, u := range users.GetUsers() {
					if u.GetID() == tt.wantUser.GetID() {
						if diff := cmp.Diff(tt.wantUser, u, protocmp.Transform()); diff != "" {
							t.Errorf("%s: UpdateUser() mismatch (-wantUser +gotUser):\n%s", tt.name, diff)
						}
					}
				}
			}
		})
	}
}
