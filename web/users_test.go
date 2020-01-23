package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/web"
	"github.com/autograde/aguis/web/auth"
	"github.com/google/go-cmp/cmp"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

func TestGetSelf(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	_ = createFakeUser(t, db, 1)

	const (
		selfURL   = "/user"
		apiPrefix = "/api/v1"
	)

	r := httptest.NewRequest(http.MethodGet, selfURL, nil)
	w := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(r, w)

	user := &pb.User{ID: 1}
	c.Set(auth.UserKey, user)

	userHandler := web.GetSelf(db)
	if err := userHandler(c); err != nil {
		t.Error(err)
	}

	if w.Code == http.StatusNotFound {
		t.Fatal("not found")
	}
	assertCode(t, w.Code, http.StatusFound)
}

func TestGetUsers(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	_, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	unexpectedUsers, err := ags.GetUsers(context.Background(), &pb.Void{})
	if err == nil && unexpectedUsers != nil && len(unexpectedUsers.GetUsers()) > 0 {
		t.Fatalf("found unexpected users %+v", unexpectedUsers)
	}

	admin := createFakeUser(t, db, 1)
	user2 := createFakeUser(t, db, 2)
	ctx := withUserContext(context.Background(), user2)
	_, err = ags.GetUsers(ctx, &pb.Void{})
	if err == nil {
		t.Fatal("expected 'rpc error: code = PermissionDenied desc = only admin can access other users'")
	}
	// now switch to use admin as the user; this should pass
	ctx = withUserContext(context.Background(), admin)
	foundUsers, err := ags.GetUsers(ctx, &pb.Void{})
	if err != nil {
		t.Fatal(err)
	}

	wantUsers := make([]*pb.User, 0)
	wantUsers = append(wantUsers, admin, user2)

	if !cmp.Equal(foundUsers.Users, wantUsers) {
		t.Errorf("have users %+v want %+v", foundUsers.Users, wantUsers)
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
	db, cleanup := setup(t)
	defer cleanup()

	var users []*pb.User
	for _, u := range allUsers {
		user := createFakeUser(t, db, u.remoteID)
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

	_, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	ctx := withUserContext(context.Background(), admin)

	// users to enroll in course DAT520 Distributed Systems
	// (excluding admin because admin is enrolled on creation)
	wantUsers := users[0 : len(allUsers)-3]
	for i, user := range wantUsers {
		if i == 0 {
			// skip enrolling admin as student
			continue
		}
		if err := db.CreateEnrollment(&pb.Enrollment{
			UserID:   user.ID,
			CourseID: allCourses[0].ID,
		}); err != nil {
			t.Fatal(err)
		}
		if err := db.EnrollStudent(user.ID, allCourses[0].ID); err != nil {
			t.Fatal(err)
		}
	}

	// users to enroll in course DAT320 Operating Systems
	// (excluding admin because admin is enrolled on creation)
	osUsers := users[3:7]
	for _, user := range osUsers {
		if err := db.CreateEnrollment(&pb.Enrollment{
			UserID:   user.ID,
			CourseID: allCourses[1].ID,
		}); err != nil {
			t.Fatal(err)
		}
		if err := db.EnrollStudent(user.ID, allCourses[1].ID); err != nil {
			t.Fatal(err)
		}
	}

	foundEnrollments, err := ags.GetEnrollmentsByCourse(ctx, &pb.EnrollmentRequest{CourseID: allCourses[0].ID})
	if err != nil {
		t.Error(err)
	}

	var foundUsers []*pb.User
	for _, e := range foundEnrollments.Enrollments {
		// remote identities should not be loaded.
		e.User.RemoteIdentities = nil
		foundUsers = append(foundUsers, e.User)
	}

	if !cmp.Equal(foundUsers, wantUsers) {
		for _, u := range foundUsers {
			t.Logf("user %+v", u)
		}
		for _, u := range wantUsers {
			t.Logf("want %+v", u)
		}
		t.Errorf("have users %+v want %+v", foundUsers, wantUsers)
	}
}

func TestEnrollmentsWithoutGroupMembership(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	var users []*pb.User
	for _, u := range allUsers {
		user := createFakeUser(t, db, u.remoteID)
		users = append(users, user)
	}
	admin := users[0]

	_, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	ctx := withUserContext(context.Background(), admin)

	course := allCourses[1]
	err := db.CreateCourse(admin.ID, course)
	if err != nil {
		t.Fatal(err)
	}

	var wantEnrollments []*pb.Enrollment
	for i, user := range users {
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
			if err := db.CreateEnrollment(&pb.Enrollment{
				UserID: user.ID, CourseID: course.ID, GroupID: 1}); err != nil {
				t.Fatal(err)
			}
			if err := db.EnrollStudent(user.ID, course.ID); err != nil {
				t.Fatal(err)
			}
		} else {
			// enroll rest of the students and add them to the list to check against
			if err := db.CreateEnrollment(&pb.Enrollment{
				UserID: user.ID, CourseID: course.ID}); err != nil {
				t.Fatal(err)
			}
			if err := db.EnrollStudent(user.ID, course.ID); err != nil {
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

	gotEnrollments, err := ags.GetEnrollmentsByCourse(ctx, &pb.EnrollmentRequest{CourseID: course.ID, FilterOutGroupMembers: true})
	if err != nil {
		t.Fatal(err)
	}
	// set user references to nil as db methods populating the first list will not have them
	for _, u := range gotEnrollments.Enrollments {
		u.User = nil
		u.Course = nil
	}

	if !cmp.Equal(gotEnrollments.Enrollments, wantEnrollments) {
		for _, u := range gotEnrollments.Enrollments {
			t.Logf("user %+v", u)
		}
		for _, u := range wantEnrollments {
			t.Logf("want %+v", u)
		}
		t.Errorf("have users %+v want %+v", gotEnrollments, wantEnrollments)
	}

}

func TestUpdateUser(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()
	firstAdminUser := createFakeUser(t, db, 1)
	nonAdminUser := createFakeUser(t, db, 11)

	_, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	ctx := withUserContext(context.Background(), firstAdminUser)

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

	nameChangeRequest := &pb.User{
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
	withName, err := db.GetUser(nonAdminUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	withName.Enrollments = nil
	// withName.Enrollments = make([]*pb.Enrollment, 0)
	wantUser := &pb.User{
		ID:               withName.ID,
		Name:             "Scrooge McDuck",
		IsAdmin:          true,
		StudentID:        "99",
		Email:            "test@test.com",
		AvatarURL:        "www.hello.com",
		RemoteIdentities: nonAdminUser.RemoteIdentities,
	}

	if !cmp.Equal(withName, wantUser) {
		t.Errorf("have users\n%+v want\n%+v\n", withName, wantUser)
	}
}

func TestUpdateUserFailures(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()
	//user := &pb.User{Name: "Test User", StudentID: "11", Email: "test@email", AvatarURL: "url.com"}
	adminUser := createFakeUser(t, db, 1)
	createFakeUser(t, db, 11)

	_, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})

	u := createFakeUser(t, db, 3)
	if u.IsAdmin {
		t.Fatalf("expected user %v to be non-admin", u)
	}
	// context with user u (non-admin user); can only change its own name etc
	ctx := withUserContext(context.Background(), u)

	// trying to demote current adminUser by setting IsAdmin to false
	nameChangeRequest := &pb.User{
		ID:        adminUser.ID,
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

	noChangeAdmin, err := db.GetUser(adminUser.ID)
	noChangeAdmin.Enrollments = nil
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(noChangeAdmin, adminUser) {
		t.Errorf("\nhave: %+v\nwant: %+v\n", noChangeAdmin, adminUser)
	}

	nameChangeRequest = &pb.User{
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
	withName, err := db.GetUser(u.ID)
	if err != nil {
		t.Fatal(err)
	}
	withName.Enrollments = nil
	wantUser := &pb.User{
		ID:               withName.ID,
		Name:             "Scrooge McDuck",
		IsAdmin:          false, // we want that the current user u cannot promote himself to admin
		StudentID:        "99",
		Email:            "test@test.com",
		AvatarURL:        "www.hello.com",
		RemoteIdentities: u.RemoteIdentities,
	}

	if !cmp.Equal(withName, wantUser) {
		t.Errorf("\nhave: %+v\nwant: %+v\n", withName, wantUser)
	}
}

// createFakeUser is a test helper to create a user in the database
// with the given remote id and the fake scm provider.
func createFakeUser(t *testing.T, db database.Database, remoteID uint64) *pb.User {
	t.Helper()
	var user pb.User
	err := db.CreateUserFromRemoteIdentity(&user,
		&pb.RemoteIdentity{
			Provider:    "fake",
			RemoteID:    remoteID,
			AccessToken: "token",
		})
	if err != nil {
		t.Fatal(err)
	}
	return &user
}
