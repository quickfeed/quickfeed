package web_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/web"
	"github.com/autograde/aguis/web/auth"
	"github.com/labstack/echo"
)

func TestGetSelf(t *testing.T) {
	const (
		selfURL   = "/user"
		apiPrefix = "/api/v1"
	)

	r := httptest.NewRequest(http.MethodGet, selfURL, nil)
	w := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(r, w)

	user := &models.User{ID: 1}
	c.Set(auth.UserKey, user)

	userHandler := web.GetSelf()
	if err := userHandler(c); err != nil {
		t.Error(err)
	}

	userURL := fmt.Sprintf("/users/%d", user.ID)
	location := w.Header().Get("Location")
	if location != apiPrefix+userURL {
		t.Errorf("have Location '%v' want '%v'", location, apiPrefix+userURL)
	}
	assertCode(t, w.Code, http.StatusFound)
}

func TestGetUser(t *testing.T) {
	const route = "/users/:uid"

	db, cleanup := setup(t)
	defer cleanup()

	// Create first user (the admin).
	createFakeUser(t, db, 1)
	user := createFakeUser(t, db, 2)

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodGet, route, web.GetUser(db))

	requestURL := fmt.Sprintf("/users/%d", user.ID)
	r := httptest.NewRequest(http.MethodGet, requestURL, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	// Prepare context with user request.
	router.Find(http.MethodGet, requestURL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}

	var foundUser *models.User
	if err := json.Unmarshal(w.Body.Bytes(), &foundUser); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(foundUser, user) {
		t.Errorf("have user %+v want %+v", foundUser, user)
	}
	assertCode(t, w.Code, http.StatusFound)
}

func TestGetUsers(t *testing.T) {
	const route = "/users"

	db, cleanup := setup(t)
	defer cleanup()

	user1 := createFakeUser(t, db, 1)
	user2 := createFakeUser(t, db, 2)

	e := echo.New()
	r := httptest.NewRequest(http.MethodGet, route, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)

	h := web.GetUsers(db)
	if err := h(c); err != nil {
		t.Error(err)
	}

	var foundUsers []*models.User
	if err := json.Unmarshal(w.Body.Bytes(), &foundUsers); err != nil {
		t.Fatal(err)
	}

	// Remote identities should not be loaded.
	user1.RemoteIdentities = nil
	user2.RemoteIdentities = nil
	// First user should be admin.
	admin := true
	user1.IsAdmin = &admin
	wantUsers := []*models.User{user1, user2}
	if !reflect.DeepEqual(foundUsers, wantUsers) {
		t.Errorf("have users %+v want %+v", foundUsers, wantUsers)
	}

	assertCode(t, w.Code, http.StatusFound)
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
	const route = "/courses/:cid/users"

	db, cleanup := setup(t)
	defer cleanup()

	var users []*models.User
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

	// users to enroll in course DAT520 Distributed Systems
	// (excluding admin because admin is enrolled on creation)
	wantUsers := users[0 : len(allUsers)-3]
	for i, user := range wantUsers {
		if i == 0 {
			// skip enrolling admin as student
			continue
		}
		if err := db.CreateEnrollment(&models.Enrollment{
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
		if err := db.CreateEnrollment(&models.Enrollment{
			UserID:   user.ID,
			CourseID: allCourses[1].ID,
		}); err != nil {
			t.Fatal(err)
		}
		if err := db.EnrollStudent(user.ID, allCourses[1].ID); err != nil {
			t.Fatal(err)
		}
	}

	e := echo.New()
	router := echo.NewRouter(e)

	// add the route to handler
	router.Add(http.MethodGet, route, web.GetEnrollmentsByCourse(db))
	requestURL := fmt.Sprintf("/courses/%d/users", allCourses[0].ID)
	r := httptest.NewRequest(http.MethodGet, requestURL, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	router.Find(http.MethodGet, requestURL, c)
	// invoke the prepared handler
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}

	var foundEnrollments []*models.Enrollment
	if err := json.Unmarshal(w.Body.Bytes(), &foundEnrollments); err != nil {
		t.Fatal(err)
	}
	var foundUsers []*models.User
	for _, e := range foundEnrollments {
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

	assertCode(t, w.Code, http.StatusOK)
}

func TestPatchUser(t *testing.T) {
	const route = "/users/:uid"

	db, cleanup := setup(t)
	defer cleanup()

	adminUser := createFakeUser(t, db, 1)
	user := createFakeUser(t, db, 2)

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodPatch, route, web.PatchUser(db))

	// Send empty request, the user should not be modified.
	emptyJSON, err := json.Marshal(&web.UpdateUserRequest{})
	if err != nil {
		t.Fatal(err)
	}
	requestBody := bytes.NewReader(emptyJSON)

	requestURL := fmt.Sprintf("/users/%d", user.ID)
	r := httptest.NewRequest(http.MethodPatch, requestURL, requestBody)
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	// Prepare context with user request.
	c.Set("user", adminUser)
	router.Find(http.MethodPatch, requestURL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}
	assertCode(t, w.Code, http.StatusNotModified)

	tmp := true
	// Send request with IsAdmin set to true, the user should become admin.
	trueJSON, err := json.Marshal(&web.UpdateUserRequest{
		IsAdmin: &tmp,
	})
	if err != nil {
		t.Fatal(err)
	}
	requestBody.Reset(trueJSON)

	r = httptest.NewRequest(http.MethodPatch, requestURL, requestBody)
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w = httptest.NewRecorder()
	c.Reset(r, w)
	// Prepare context with user request.
	c.Set("user", adminUser)
	router.Find(http.MethodPatch, requestURL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}
	assertCode(t, w.Code, http.StatusOK)

	admin, err := db.GetUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !admin.IAdmin() {
		t.Error("expected user to have become admin")
	}

	// Send request with Name.
	nameChangeJSON, err := json.Marshal(&web.UpdateUserRequest{
		Name:      "Scrooge McDuck",
		StudentID: "99",
		Email:     "test@test.com",
		AvatarURL: "www.hello.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	requestBody.Reset(nameChangeJSON)

	r = httptest.NewRequest(http.MethodPatch, requestURL, requestBody)
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w = httptest.NewRecorder()
	c.Reset(r, w)
	// Prepare context with user request.
	c.Set("user", adminUser)
	router.Find(http.MethodPatch, requestURL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}
	assertCode(t, w.Code, http.StatusOK)

	withName, err := db.GetUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantAdmin := true
	wantUser := &models.User{
		ID:               withName.ID,
		Name:             "Scrooge McDuck",
		IsAdmin:          &wantAdmin,
		StudentID:        "99",
		Email:            "test@test.com",
		AvatarURL:        "www.hello.com",
		RemoteIdentities: user.RemoteIdentities,
	}
	if !reflect.DeepEqual(withName, wantUser) {
		t.Errorf("have users %+v want %+v", withName, wantUser)
	}
}

// createFakeUser is a test helper to create a user in the database
// with the given remote id and the fake scm provider.
func createFakeUser(t *testing.T, db database.Database, remoteID uint64) *models.User {
	var user models.User
	err := db.CreateUserFromRemoteIdentity(&user,
		&models.RemoteIdentity{
			Provider: "fake",
			RemoteID: remoteID,
		})
	if err != nil {
		t.Fatal(err)
	}
	return &user
}
