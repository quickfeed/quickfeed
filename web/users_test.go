package web_test

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetSelf(t *testing.T) {
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()

	const (
		bufSize = 1024 * 1024
	)

	adminUser := qtest.CreateFakeUser(t, db, 1)
	student := qtest.CreateFakeUser(t, db, 56)

	store := sessions.NewCookieStore([]byte("secret"))
	store.Options.HttpOnly = true
	store.Options.Secure = true
	gothic.Store = store

	lis := bufconn.Listen(bufSize)
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	opt := grpc.ChainUnaryInterceptor(auth.UnaryUserVerifier())
	s := grpc.NewServer(opt)
	qf.RegisterQuickFeedServiceServer(s, ags)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := qf.NewQuickFeedServiceClient(conn)

	userTest := []struct {
		id       uint64
		code     codes.Code
		metadata bool
		token    string
		wantUser *qf.User
	}{
		{id: 1, code: codes.Unauthenticated, metadata: false, token: "", wantUser: nil},
		{id: 6, code: codes.PermissionDenied, metadata: true, token: "", wantUser: nil},
		{id: 1, code: codes.Unauthenticated, metadata: true, token: "should fail", wantUser: nil},
		{id: 1, code: codes.OK, metadata: true, token: "", wantUser: adminUser},
		{id: 2, code: codes.OK, metadata: true, token: "", wantUser: student},
	}

	for _, user := range userTest {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		sess := sessions.NewSession(store, auth.SessionKey)

		sess.Values[auth.UserKey] = &auth.UserSession{
			ID:        user.id,
			Providers: map[string]struct{}{"github": {}},
		}
		if err := sess.Save(r, w); err != nil {
			t.Errorf("sess.Save(): %v", err)
		}

		token := w.Result().Header.Get(auth.OutgoingCookie)
		auth.Add(token, user.id)

		if user.metadata {
			meta := metadata.MD{}
			if len(user.token) > 0 {
				token = user.token
			}
			meta.Set(auth.Cookie, token)
			ctx = metadata.NewOutgoingContext(ctx, meta)
		}
		gotUser, err := client.GetUser(ctx, &qf.Void{})
		if s, ok := status.FromError(err); ok {
			if s.Code() != user.code {
				t.Errorf("GetUser().Code(): %v, want: %v", s.Code(), user.code)
			}
		}
		if user.wantUser != nil {
			// ignore comparing remote identity
			user.wantUser.RemoteIdentities = nil
		}
		wantUser := user.wantUser
		if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
			t.Errorf("GetSelf() mismatch (-wantUser +gotUser):\n%s", diff)
		}
	}
}

func TestGetUsers(t *testing.T) {
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()

	unexpectedUsers, err := ags.GetUsers(context.Background(), &qf.Void{})
	if err == nil && unexpectedUsers != nil && len(unexpectedUsers.GetUsers()) > 0 {
		t.Fatalf("found unexpected users %+v", unexpectedUsers)
	}

	admin := qtest.CreateFakeUser(t, db, 1)
	user2 := qtest.CreateFakeUser(t, db, 2)
	ctx := qtest.WithUserContext(context.Background(), user2)
	_, err = ags.GetUsers(ctx, &qf.Void{})
	if err == nil {
		t.Fatal("expected 'rpc error: code = PermissionDenied desc = only admin can access other users'")
	}
	// now switch to use admin as the user; this should pass
	ctx = qtest.WithUserContext(context.Background(), admin)
	foundUsers, err := ags.GetUsers(ctx, &qf.Void{})
	if err != nil {
		t.Fatal(err)
	}

	wantUsers := make([]*qf.User, 0)
	wantUsers = append(wantUsers, admin, user2)
	gotUsers := foundUsers.GetUsers()
	if diff := cmp.Diff(wantUsers, gotUsers, protocmp.Transform()); diff != "" {
		t.Errorf("ags.GetUsers() mismatch (-wantUsers +gotUsers):\n%s", diff)
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
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()

	var users []*qf.User
	for _, u := range allUsers {
		user := qtest.CreateFakeUser(t, db, u.remoteID)
		// remote identities should not be loaded.
		user.RemoteIdentities = nil
		users = append(users, user)
	}
	admin := users[0]
	for _, course := range allCourses {
		err := db.CreateCourse(admin.ID, course)
		if err != nil {
			t.Fatal(err)
		}
	}

	ctx := qtest.WithUserContext(context.Background(), admin)

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
			CourseID: allCourses[0].ID,
		}); err != nil {
			t.Fatal(err)
		}
		if err := db.UpdateEnrollment(&qf.Enrollment{
			UserID:   user.ID,
			CourseID: allCourses[0].ID,
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
			CourseID: allCourses[1].ID,
		}); err != nil {
			t.Fatal(err)
		}
		if err := db.UpdateEnrollment(&qf.Enrollment{
			UserID:   user.ID,
			CourseID: allCourses[1].ID,
			Status:   qf.Enrollment_STUDENT,
		}); err != nil {
			t.Fatal(err)
		}
	}

	gotEnrollments, err := ags.GetEnrollmentsByCourse(ctx, &qf.EnrollmentRequest{CourseID: allCourses[0].ID})
	if err != nil {
		t.Error(err)
	}
	var gotUsers []*qf.User
	for _, e := range gotEnrollments.Enrollments {
		gotUsers = append(gotUsers, e.User)
	}
	if diff := cmp.Diff(wantUsers, gotUsers, protocmp.Transform()); diff != "" {
		t.Errorf("ags.GetEnrollmentsByCourse() mismatch (-wantUsers +gotUsers):\n%s", diff)
	}
}

func TestEnrollmentsWithoutGroupMembership(t *testing.T) {
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()

	var users []*qf.User
	for _, u := range allUsers {
		user := qtest.CreateFakeUser(t, db, u.remoteID)
		users = append(users, user)
	}
	admin := users[0]

	ctx := qtest.WithUserContext(context.Background(), admin)

	course := allCourses[1]
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

	enrollments, err := ags.GetEnrollmentsByCourse(ctx, &qf.EnrollmentRequest{CourseID: course.ID, IgnoreGroupMembers: true})
	if err != nil {
		t.Fatal(err)
	}
	gotEnrollments := enrollments.GetEnrollments()
	// set user references to nil as db methods populating the first list will not have them
	for _, u := range gotEnrollments {
		u.User = nil
		u.Course = nil
	}
	if diff := cmp.Diff(wantEnrollments, gotEnrollments, protocmp.Transform()); diff != "" {
		t.Errorf("ags.GetEnrollmentsByCourse() mismatch (-wantEnrollments +gotEnrollments):\n%s", diff)
	}
}

func TestUpdateUser(t *testing.T) {
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()
	firstAdminUser := qtest.CreateFakeUser(t, db, 1)
	nonAdminUser := qtest.CreateFakeUser(t, db, 11)

	ctx := qtest.WithUserContext(context.Background(), firstAdminUser)

	// we want to update nonAdminUser to become admin
	nonAdminUser.IsAdmin = true
	err := db.UpdateUser(nonAdminUser)
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

	nameChangeRequest := &qf.User{
		ID:        nonAdminUser.ID,
		IsAdmin:   nonAdminUser.IsAdmin,
		Name:      "Scrooge McDuck",
		StudentID: "99",
		Email:     "test@test.com",
		AvatarURL: "www.hello.com",
	}

	_, err = ags.UpdateUser(ctx, nameChangeRequest)
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
		t.Errorf("ags.UpdateUser() mismatch (-wantUser +gotUser):\n%s", diff)
	}
}

func TestUpdateUserFailures(t *testing.T) {
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()
	wantAdminUser := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateFakeUser(t, db, 11)

	u := qtest.CreateFakeUser(t, db, 3)
	if u.IsAdmin {
		t.Fatalf("expected user %v to be non-admin", u)
	}
	// context with user u (non-admin user); can only change its own name etc
	ctx := qtest.WithUserContext(context.Background(), u)

	// trying to demote current adminUser by setting IsAdmin to false
	nameChangeRequest := &qf.User{
		ID:        wantAdminUser.ID,
		IsAdmin:   false,
		Name:      "Scrooge McDuck",
		StudentID: "99",
		Email:     "test@test.com",
		AvatarURL: "www.hello.com",
	}
	// current user u (non-admin) is in the ctx and tries to change adminUser
	_, err := ags.UpdateUser(ctx, nameChangeRequest)
	if err == nil {
		t.Fatal(err)
	}

	gotAdminUserWithoutChanges, err := db.GetUser(wantAdminUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantAdminUser, gotAdminUserWithoutChanges, protocmp.Transform()); diff != "" {
		t.Errorf("ags.UpdateUser() mismatch (-wantAdminUser +gotAdminUserWithoutChanges):\n%s", diff)
	}

	nameChangeRequest = &qf.User{
		ID:        u.ID,
		IsAdmin:   true,
		Name:      "Scrooge McDuck",
		StudentID: "99",
		Email:     "test@test.com",
		AvatarURL: "www.hello.com",
	}
	_, err = ags.UpdateUser(ctx, nameChangeRequest)
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
		t.Errorf("ags.UpdateUser() mismatch (-wantUser +gotUser):\n%s", diff)
	}
}
