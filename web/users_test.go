package web_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetUsers(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := web.MockClient(t, db, nil)
	ctx := context.Background()

	unexpectedUsers, err := client.GetUsers(ctx, &connect.Request[qf.Void]{Msg: &qf.Void{}})
	if err == nil && unexpectedUsers != nil && len(unexpectedUsers.Msg.GetUsers()) > 0 {
		t.Fatalf("found unexpected users %+v", unexpectedUsers)
	}

	admin := qtest.CreateFakeUser(t, db)
	user2 := qtest.CreateFakeUser(t, db)

	foundUsers, err := client.GetUsers(ctx, &connect.Request[qf.Void]{Msg: &qf.Void{}})
	if err != nil {
		t.Fatal(err)
	}

	wantUsers := make([]*qf.User, 0)
	wantUsers = append(wantUsers, admin, user2)
	gotUsers := foundUsers.Msg.GetUsers()
	if diff := cmp.Diff(wantUsers, gotUsers, protocmp.Transform()); diff != "" {
		t.Errorf("GetUsers() mismatch (-wantUsers +gotUsers):\n%s", diff)
	}
}

func TestGetEnrollmentsByCourse(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := web.MockClient(t, db, nil)
	ctx := context.Background()

	var users []*qf.User
	for i := 0; i < 10; i++ {
		user := qtest.CreateFakeUser(t, db)
		users = append(users, user)
	}
	admin := users[0]
	for _, course := range qtest.MockCourses {
		err := db.CreateCourse(admin.ID, course)
		if err != nil {
			t.Fatal(err)
		}
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

	request := &connect.Request[qf.EnrollmentRequest]{
		Msg: &qf.EnrollmentRequest{
			FetchMode: &qf.EnrollmentRequest_CourseID{
				CourseID: qtest.MockCourses[0].ID,
			},
		},
	}
	gotEnrollments, err := client.GetEnrollments(ctx, request)
	if err != nil {
		t.Error(err)
	}
	var gotUsers []*qf.User
	for _, e := range gotEnrollments.Msg.Enrollments {
		gotUsers = append(gotUsers, e.User)
	}
	if diff := cmp.Diff(wantUsers, gotUsers, protocmp.Transform()); diff != "" {
		t.Errorf("GetEnrollmentsByCourse() mismatch (-wantUsers +gotUsers):\n%s", diff)
	}
}

func TestUpdateUser(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qtest.Logger(t)

	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}
	client := web.MockClient(t, db, connect.WithInterceptors(
		interceptor.NewUserInterceptor(logger, tm),
	))
	ctx := context.Background()

	firstAdminUser := qtest.CreateFakeUser(t, db)
	nonAdminUser := qtest.CreateFakeUser(t, db)

	firstAdminCookie, err := tm.NewAuthCookie(firstAdminUser.ID)
	if err != nil {
		t.Fatal(err)
	}

	// we want to update nonAdminUser to become admin
	nonAdminUser.IsAdmin = true
	err = db.UpdateUser(nonAdminUser)
	if err != nil {
		t.Fatal(err)
	}

	// we expect the nonAdminUser to now be admin
	admin, err := db.GetUser(nonAdminUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !admin.IsAdmin {
		t.Error("expected nonAdminUser to have become admin")
	}

	nameChangeRequest := connect.NewRequest(&qf.User{
		ID:        nonAdminUser.ID,
		IsAdmin:   nonAdminUser.IsAdmin,
		Name:      "Scrooge McDuck",
		StudentID: "99",
		Email:     "test@test.com",
		AvatarURL: "www.hello.com",
	})

	nameChangeRequest.Header().Set(auth.Cookie, firstAdminCookie.String())
	_, err = client.UpdateUser(ctx, nameChangeRequest)
	if err != nil {
		t.Error(err)
	}
	gotUser, err := db.GetUser(nonAdminUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantUser := &qf.User{
		ID:           gotUser.ID,
		Name:         "Scrooge McDuck",
		IsAdmin:      true,
		StudentID:    "99",
		Email:        "test@test.com",
		AvatarURL:    "www.hello.com",
		RefreshToken: nonAdminUser.RefreshToken,
		ScmRemoteID:  nonAdminUser.ScmRemoteID,
	}
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateUser() mismatch (-wantUser +gotUser):\n%s", diff)
	}
}

func TestUpdateUserFailures(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client, tm := web.MockClientWithOption(t, db, scm.WithMockOrgs())
	ctx := context.Background()

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "admin", Login: "admin"})
	if !admin.IsAdmin {
		t.Fatalf("expected user %v to be admin", admin)
	}
	user := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "user", Login: "user"})
	if user.IsAdmin {
		t.Fatalf("expected user %v to be non-admin", user)
	}
	userCookie := Cookie(t, tm, user)
	adminCookie := Cookie(t, tm, admin)
	tests := []struct {
		name     string
		cookie   string
		req      *qf.User
		wantUser *qf.User
		wantErr  bool
	}{
		{
			name:   "user demotes admin, must fail",
			cookie: userCookie,
			req: &qf.User{
				ID:        admin.ID,
				IsAdmin:   false,
				Name:      admin.Name,
				Email:     admin.Email,
				StudentID: admin.StudentID,
				AvatarURL: admin.AvatarURL,
			},
			wantErr: true,
		},
		{
			name:   "user promotes self to admin, must fail",
			cookie: userCookie,
			req: &qf.User{
				ID:        user.ID,
				Name:      user.Name,
				Email:     user.Email,
				StudentID: user.StudentID,
				AvatarURL: user.AvatarURL,
				IsAdmin:   true,
			},
			wantErr: true,
		},
		{
			name:   "admin changes own name, must pass",
			cookie: adminCookie,
			req: &qf.User{
				ID:   admin.ID,
				Name: "super user",
			},
			wantUser: &qf.User{
				ID:           admin.ID,
				IsAdmin:      true,
				Login:        admin.Login,
				Name:         "super user",
				RefreshToken: admin.RefreshToken,
				ScmRemoteID:  admin.ScmRemoteID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// UpdateUser returns void, so we cannot check that the user was updated
			_, err := client.UpdateUser(ctx, qtest.RequestWithCookie(tt.req, tt.cookie))
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: expected error: %v, got = %v", tt.name, tt.wantErr, err)
			}
			if !tt.wantErr {
				// Instead (for success cases), get all users and check that the user was updated
				users, err := client.GetUsers(ctx, qtest.RequestWithCookie(&qf.Void{}, tt.cookie))
				if err != nil {
					t.Fatal(err)
				}
				for _, u := range users.Msg.GetUsers() {
					if u.ID == tt.wantUser.ID {
						if diff := cmp.Diff(tt.wantUser, u, protocmp.Transform()); diff != "" {
							t.Errorf("%s: UpdateUser() mismatch (-wantUser +gotUser):\n%s", tt.name, diff)
						}
					}
				}
			}
		})
	}
}
