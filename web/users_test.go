package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
	"github.com/autograde/aguis/web"
	"github.com/autograde/aguis/web/auth"
	"github.com/autograde/aguis/web/grpc_service"
	"github.com/labstack/echo"
	"google.golang.org/grpc/metadata"
)

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

	user := &pb.User{Id: 1}
	c.Set(auth.UserKey, user)

	userHandler := web.GetSelf(db)
	if err := userHandler(c); err != nil {
		t.Error(err)
	}

	userURL := "/users/" + strconv.FormatUint(user.Id, 10)
	location := w.Header().Get("Location")
	if location != apiPrefix+userURL {
		t.Errorf("have Location '%v' want '%v'", location, apiPrefix+userURL)
	}
	assertCode(t, w.Code, http.StatusFound)
}

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
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(user.Id))

	foundUser, err := test_ag.GetUser(cont, &pb.RecordRequest{Id: user.Id})
	if err != nil {
		t.Error(err)
	}
	user.RemoteIdentities = make([]*pb.RemoteIdentity, 0)
	if !reflect.DeepEqual(foundUser, &user) {
		t.Errorf("have user %+v want %+v", foundUser, &user)
	}
}

func TestGetUsers(t *testing.T) {
	const (
		route = "/users"

		github = "github"
		gitlab = "gitlab"
	)

	db, cleanup := setup(t)
	defer cleanup()

	var user1 pb.User
	if err := db.CreateUserFromRemoteIdentity(
		&user1,
		&pb.RemoteIdentity{
			Provider: github,
		},
	); err != nil {
		t.Fatal(err)
	}
	var user2 pb.User
	if err := db.CreateUserFromRemoteIdentity(
		&user2,
		&pb.RemoteIdentity{
			Provider: gitlab,
		},
	); err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(user1.Id))

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
	if !reflect.DeepEqual(foundUsers.Users, wantUsers) {
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
		err := db.CreateCourse(admin.Id, course)
		if err != nil {
			t.Fatal(err)
		}
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(admin.Id))

	// users to enroll in course DAT520 Distributed Systems
	// (excluding admin because admin is enrolled on creation)
	wantUsers := users[0 : len(allUsers)-3]
	for i, user := range wantUsers {
		if i == 0 {
			// skip enrolling admin as student
			continue
		}
		if err := db.CreateEnrollment(&pb.Enrollment{
			UserId:   user.Id,
			CourseId: allCourses[0].Id,
		}); err != nil {
			t.Fatal(err)
		}
		if err := db.EnrollStudent(user.Id, allCourses[0].Id); err != nil {
			t.Fatal(err)
		}
	}

	// users to enroll in course DAT320 Operating Systems
	// (excluding admin because admin is enrolled on creation)
	osUsers := users[3:7]
	for _, user := range osUsers {
		if err := db.CreateEnrollment(&pb.Enrollment{
			UserId:   user.Id,
			CourseId: allCourses[1].Id,
		}); err != nil {
			t.Fatal(err)
		}
		if err := db.EnrollStudent(user.Id, allCourses[1].Id); err != nil {
			t.Fatal(err)
		}
	}

	foundEnrollments, err := test_ag.GetEnrollmentsByCourse(cont, &pb.RecordRequest{Id: allCourses[0].Id})
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

	var user pb.User
	var adminUser pb.User
	isAdmin := true
	adminUser.IsAdmin = isAdmin
	var remoteIdentity pb.RemoteIdentity
	if err := db.CreateUserFromRemoteIdentity(
		&user, &remoteIdentity,
	); err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(adminUser.Id))

	respUser, err := test_ag.UpdateUser(cont, &user)
	if err != nil {
		t.Fatal(err)
	}

	admin, err := db.GetUser(user.Id)
	if err != nil {
		t.Fatal(err)
	}

	if !admin.IsAdmin {
		t.Error("expected user to have become admin")
	}

	namechangeRequest := &pb.User{
		Id:        respUser.Id,
		Name:      "Scrooge McDuck",
		StudentId: "99",
		Email:     "test@test.com",
		AvatarUrl: "www.hello.com",
	}

	_, err = test_ag.UpdateUser(cont, namechangeRequest)
	if err != nil {
		t.Error(err)
	}
	withName, err := db.GetUser(user.Id)
	if err != nil {
		t.Fatal(err)
	}
	wantAdmin := true
	wantUser := &pb.User{
		Id:               withName.Id,
		Name:             "Scrooge McDuck",
		IsAdmin:          wantAdmin,
		StudentId:        "99",
		Email:            "test@test.com",
		AvatarUrl:        "www.hello.com",
		RemoteIdentities: []*pb.RemoteIdentity{&remoteIdentity},
	}
	if !reflect.DeepEqual(withName, wantUser) {
		t.Errorf("have users %+v want %+v", withName, wantUser)
	}
}
