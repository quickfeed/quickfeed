package web_test

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"
	"github.com/autograde/aguis/web"
	"github.com/autograde/aguis/web/grpc_service"
	"google.golang.org/grpc/metadata"
)

/*
func TestGetSelf(t *testing.T) {
	const (
		selfURL   = "/user"
		apiPrefix = "/api/v1"
	)

	db, cleanup := setup(t)
	defer cleanup()

	r := httptest.NewRequest(http.MethodGet, selfURL, nil)
	w := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(r, w)

	user := &models.User{ID: 1}
	c.Set(auth.UserKey, user)

	userHandler := web.GetSelf(db)
	if err := userHandler(c); err != nil {
		t.Error(err)
	}

	userURL := fmt.Sprintf("/users/%d", user.ID)
	location := w.Header().Get("Location")
	if location != apiPrefix+userURL {
		t.Errorf("have Location '%v' want '%v'", location, apiPrefix+userURL)
	}
	assertCode(t, w.Code, http.StatusFound)
}*/

func TestGetUser(t *testing.T) {

	const (
		provider    = "github"
		accessToken = "secret"
	)

	db, cleanup := setup(t)
	defer cleanup()

	// Create first user (the admin).
	if err := db.CreateUserFromRemoteIdentity(
		&pb.User{},
		&pb.RemoteIdentity{},
	); err != nil {
		t.Fatal(err)
	}

	var user pb.User
	if err := db.CreateUserFromRemoteIdentity(
		&user,
		&pb.RemoteIdentity{
			Provider:    provider,
			AccessToken: accessToken,
		},
	); err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(user.ID))

	foundUser, err := test_ag.GetUser(cont, &pb.RecordRequest{ID: user.ID})
	if err != nil {
		t.Error(err)
	}
	user.RemoteIdentities = make([]*pb.RemoteIdentity, 0)
	if !reflect.DeepEqual(foundUser, &user) {
		t.Errorf("have user %+v want %+v", foundUser, &user)
	}
}

/*
func TestGetUsers(t *testing.T) {
	const route = "/users"

	db, cleanup := setup(t)
	defer cleanup()

	var user1 pb.User
	if err := db.CreateUserFromRemoteIdentity(
		&user1,
		&pb.RemoteIdentity{
			Provider: "github",
		},
	); err != nil {
		t.Fatal(err)
	}
	var user2 pb.User
	if err := db.CreateUserFromRemoteIdentity(
		&user2,
		&pb.RemoteIdentity{
			Provider: "gitlab",
		},
	); err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(user1.ID))

	foundUsers, err := test_ag.GetUsers(cont, &pb.Void{})
	if err != nil {
		t.Fatal(err)
	}

	// Remote identities should not be loaded.
	user1.RemoteIdentities = nil
	user2.RemoteIdentities = nil
	// First user should be admin.
	admin := true
	user1.IsAdmin = admin
	wantUsers := make([]*pb.User, 0)
	wantUsers = append(wantUsers, &user1)
	wantUsers = append(wantUsers, &user2)
	gotUsers := foundUsers.Users

	if !cmp.Equal(gotUsers, wantUsers) {
		t.Errorf("have users %+v want %+v", foundUsers.Users, wantUsers)
	}

}*/

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

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(admin.ID))

	// users to enroll in course DAT520 Distributed Systems
	// (excluding admin because admin is enrolled on creation)
	wantUsers := users[0 : len(allUsers)-3]
	for i, user := range wantUsers {
		if i == 0 {
			// skip enrolling admin as student
			continue
		}
		if err := db.CreateEnrollment(&pb.Enrollment{
			User_ID:   user.ID,
			Course_ID: allCourses[0].ID,
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
			User_ID:   user.ID,
			Course_ID: allCourses[1].ID,
		}); err != nil {
			t.Fatal(err)
		}
		if err := db.EnrollStudent(user.ID, allCourses[1].ID); err != nil {
			t.Fatal(err)
		}
	}

	foundEnrollments, err := test_ag.GetEnrollmentsByCourse(cont, &pb.RecordRequest{ID: allCourses[0].ID})
	if err != nil {
		t.Error(err)
	}

	var foundUsers []*pb.User
	for _, e := range foundEnrollments.Enrollments {
		// remote identities should not be loaded.
		e.User.RemoteIdentities = nil
		foundUsers = append(foundUsers, e.User)
	}

	if !reflect.DeepEqual(foundUsers, wantUsers) {
		for _, u := range foundUsers {
			t.Logf("user %+v", u)
		}
		for _, u := range wantUsers {
			t.Logf("want %+v", u)
		}
		t.Errorf("have users %+v want %+v", foundUsers, wantUsers)
	}

}

func TestPatchUser(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()
	user := &pb.User{Name: "Test User", Student_ID: "11", Email: "test@email", Avatar_URL: "url.com"}
	adminUser := createFakeUser(t, db, 1)
	remoteIdentity := &pb.RemoteIdentity{Provider: "fake", AccessToken: "token"}
	if err := db.CreateUserFromRemoteIdentity(
		user, remoteIdentity,
	); err != nil {
		t.Fatal(err)
	}
	user, err := db.GetUserByRemoteIdentity(remoteIdentity)
	if err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	meta := metadata.New(map[string]string{"user": strconv.Itoa(int(adminUser.ID))})
	cont := metadata.NewIncomingContext(context.Background(), meta)

	respUser, err := web.PatchUser(adminUser, user, db)
	if err != nil {
		t.Fatal(err)
	}

	admin, err := db.GetUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !admin.IsAdmin {
		t.Error("expected user to have become admin")
	}

	namechangeRequest := &pb.User{
		ID:         respUser.ID,
		IsAdmin:    respUser.IsAdmin,
		Name:       "Scrooge McDuck",
		Student_ID: "99",
		Email:      "test@test.com",
		Avatar_URL: "www.hello.com",
	}

	_, err = test_ag.UpdateUser(cont, namechangeRequest)
	if err != nil {
		t.Error(err)
	}
	withName, err := db.GetUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantAdmin := true
	wantUser := &pb.User{
		ID:               withName.ID,
		Name:             "Scrooge McDuck",
		IsAdmin:          wantAdmin,
		Student_ID:       "99",
		Email:            "test@test.com",
		Avatar_URL:       "www.hello.com",
		RemoteIdentities: user.RemoteIdentities,
	}

	if !reflect.DeepEqual(withName, wantUser) {
		t.Errorf("have users %+v want %+v", withName, wantUser)
	}
}

// createFakeUser is a test helper to create a user in the database
// with the given remote id and the fake scm provider.
func createFakeUser(t *testing.T, db database.Database, remoteID uint64) *pb.User {
	var user pb.User
	err := db.CreateUserFromRemoteIdentity(&user,
		&pb.RemoteIdentity{
			Provider:    "fake",
			Remote_ID:   remoteID,
			AccessToken: "token",
		})
	if err != nil {
		t.Fatal(err)
	}
	return &user
}
