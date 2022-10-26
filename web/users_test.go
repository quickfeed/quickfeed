package web_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetUsers(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := MockClient(t, db, nil)
	ctx := context.Background()

	unexpectedUsers, err := client.GetUsers(ctx, &connect.Request[qf.Void]{Msg: &qf.Void{}})
	if err == nil && unexpectedUsers != nil && len(unexpectedUsers.Msg.GetUsers()) > 0 {
		t.Fatalf("found unexpected users %+v", unexpectedUsers)
	}

	admin := qtest.CreateFakeUser(t, db, 1)
	user2 := qtest.CreateFakeUser(t, db, 2)

	ctx = auth.WithUserContext(ctx, admin)
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

var allUsers = []struct {
	provider string
	remoteID uint64
	secret   string
}{
	{"github", 1, "123"},
	{"github", 2, "123"},
	{"github", 3, "456"},
	{"gitlab", 4, "789"},
	{"gitlab", 5, "012"},
	{"bitlab", 6, "345"},
	{"gitlab", 7, "678"},
	{"gitlab", 8, "901"},
	{"gitlab", 9, "234"},
}

func TestGetEnrollmentsByCourse(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := MockClient(t, db, nil)
	ctx := context.Background()

	var users []*qf.User
	for _, u := range allUsers {
		user := qtest.CreateFakeUser(t, db, u.remoteID)
		// remote identities should not be loaded.
		user.RemoteIdentities = nil
		users = append(users, user)
	}
	admin := users[0]
	for _, course := range qtest.MockCourses {
		err := db.CreateCourse(admin.ID, course)
		if err != nil {
			t.Fatal(err)
		}
	}

	ctx = auth.WithUserContext(ctx, admin)

	// users to enroll in course DAT520 Distributed Systems
	// (excluding admin because admin is enrolled on creation)
	wantUsers := users[0 : len(allUsers)-3]
	for i, user := range wantUsers {
		if i == 0 {
			// skip enrolling admin as student
			continue
		}
		if err := db.CreateEnrollment(&qf.Enrollment{
			UserID:   user.ID,
			CourseID: qtest.MockCourses[0].ID,
		}); err != nil {
			t.Fatal(err)
		}
		if err := db.UpdateEnrollment(&qf.Enrollment{
			UserID:   user.ID,
			CourseID: qtest.MockCourses[0].ID,
			Status:   qf.Enrollment_STUDENT,
		}); err != nil {
			t.Fatal(err)
		}
	}

	// users to enroll in course DAT320 Operating Systems
	// (excluding admin because admin is enrolled on creation)
	osUsers := users[3:7]
	for _, user := range osUsers {
		if err := db.CreateEnrollment(&qf.Enrollment{
			UserID:   user.ID,
			CourseID: qtest.MockCourses[1].ID,
		}); err != nil {
			t.Fatal(err)
		}
		if err := db.UpdateEnrollment(&qf.Enrollment{
			UserID:   user.ID,
			CourseID: qtest.MockCourses[1].ID,
			Status:   qf.Enrollment_STUDENT,
		}); err != nil {
			t.Fatal(err)
		}
	}

	request := &connect.Request[qf.EnrollmentRequest]{
		Msg: &qf.EnrollmentRequest{CourseID: qtest.MockCourses[0].ID},
	}
	gotEnrollments, err := client.GetEnrollmentsByCourse(ctx, request)
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

func TestEnrollmentsWithoutGroupMembership(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := MockClient(t, db, nil)
	ctx := context.Background()

	var users []*qf.User
	for _, u := range allUsers {
		user := qtest.CreateFakeUser(t, db, u.remoteID)
		users = append(users, user)
	}
	admin := users[0]

	ctx = auth.WithUserContext(ctx, admin)

	course := qtest.MockCourses[1]
	err := db.CreateCourse(admin.ID, course)
	if err != nil {
		t.Fatal(err)
	}

	var wantEnrollments []*qf.Enrollment
	for i, user := range users {
		query := &qf.Enrollment{
			UserID:   user.ID,
			CourseID: course.ID,
			Status:   qf.Enrollment_STUDENT,
		}
		if i == 0 {
			// we want to skip enrolling admin, as he must have been enrolled when creating course
			enr, err := db.GetEnrollmentByCourseAndUser(course.ID, user.ID)
			if err != nil {
				t.Fatal(err)
			}
			enr.User = nil
			enr.Course = nil
			wantEnrollments = append(wantEnrollments, enr)
		} else if i%3 != 0 {
			// enroll every third student as a group member
			if err := db.CreateEnrollment(&qf.Enrollment{
				UserID: user.ID, CourseID: course.ID, GroupID: 1,
			}); err != nil {
				t.Fatal(err)
			}
			if err := db.UpdateEnrollment(query); err != nil {
				t.Fatal(err)
			}
		} else {
			// enroll rest of the students and add them to the list to check against
			if err := db.CreateEnrollment(&qf.Enrollment{
				UserID: user.ID, CourseID: course.ID,
			}); err != nil {
				t.Fatal(err)
			}
			if err := db.UpdateEnrollment(query); err != nil {
				t.Fatal(err)
			}
			enr, err := db.GetEnrollmentByCourseAndUser(course.ID, user.ID)
			if err != nil {
				t.Fatal(err)
			}
			enr.User = nil
			enr.Course = nil
			wantEnrollments = append(wantEnrollments, enr)
		}
	}

	request := connect.NewRequest(
		&qf.EnrollmentRequest{CourseID: course.ID, IgnoreGroupMembers: true},
	)
	enrollments, err := client.GetEnrollmentsByCourse(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	gotEnrollments := enrollments.Msg.GetEnrollments()
	// set user references to nil as db methods populating the first list will not have them
	for _, u := range gotEnrollments {
		u.User = nil
		u.Course = nil
	}
	if diff := cmp.Diff(wantEnrollments, gotEnrollments, protocmp.Transform()); diff != "" {
		t.Errorf("GetEnrollmentsByCourse() mismatch (-wantEnrollments +gotEnrollments):\n%s", diff)
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
	client := MockClient(t, db, connect.WithInterceptors(
		interceptor.NewUserInterceptor(logger, tm),
	))
	ctx := context.Background()

	firstAdminUser := qtest.CreateFakeUser(t, db, 1)
	nonAdminUser := qtest.CreateFakeUser(t, db, 11)

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
		ID:               gotUser.ID,
		Name:             "Scrooge McDuck",
		IsAdmin:          true,
		StudentID:        "99",
		Email:            "test@test.com",
		AvatarURL:        "www.hello.com",
		RemoteIdentities: nonAdminUser.RemoteIdentities,
	}
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateUser() mismatch (-wantUser +gotUser):\n%s", diff)
	}
}

func TestUpdateUserFailures(t *testing.T) {
	t.Skip("TODO: Needs to be rewritten as a client-server test to verify (with interceptors) that the server is actually enforcing the rules")
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := MockClient(t, db, nil)
	ctx := context.Background()

	wantAdminUser := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateFakeUser(t, db, 11)

	u := qtest.CreateFakeUser(t, db, 3)
	if u.IsAdmin {
		t.Fatalf("expected user %v to be non-admin", u)
	}
	// context with user u (non-admin user); can only change its own name etc
	ctx = auth.WithUserContext(ctx, u)
	// trying to demote current adminUser by setting IsAdmin to false
	nameChangeRequest := connect.NewRequest(&qf.User{
		ID:        wantAdminUser.ID,
		IsAdmin:   false,
		Name:      "Scrooge McDuck",
		StudentID: "99",
		Email:     "test@test.com",
		AvatarURL: "www.hello.com",
	})
	// current user u (non-admin) is in the ctx and tries to change adminUser
	us, err := client.UpdateUser(ctx, nameChangeRequest)
	if err == nil {
		fmt.Println(us)
		t.Fatal(err)
	}

	gotAdminUserWithoutChanges, err := db.GetUser(wantAdminUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantAdminUser, gotAdminUserWithoutChanges, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateUser() mismatch (-wantAdminUser +gotAdminUserWithoutChanges):\n%s", diff)
	}

	nameChangeRequest = connect.NewRequest(&qf.User{
		ID:        u.ID,
		IsAdmin:   true,
		Name:      "Scrooge McDuck",
		StudentID: "99",
		Email:     "test@test.com",
		AvatarURL: "www.hello.com",
	})
	_, err = client.UpdateUser(ctx, nameChangeRequest)
	if err != nil {
		t.Error(err)
	}
	gotUser, err := db.GetUser(u.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantUser := &qf.User{
		ID:               gotUser.ID,
		Name:             "Scrooge McDuck",
		IsAdmin:          false, // we want that the current user u cannot promote himself to admin
		StudentID:        "99",
		Email:            "test@test.com",
		AvatarURL:        "www.hello.com",
		RemoteIdentities: u.RemoteIdentities,
	}
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateUser() mismatch (-wantUser +gotUser):\n%s", diff)
	}
}
