package web_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/scm"
	"github.com/autograde/aguis/web"
	"github.com/labstack/echo"
	_ "github.com/mattn/go-sqlite3"
)

func TestDeleteGroup(t *testing.T) {
	const (
		route = "/groups/:gid"
	)
	db, cleanup := setup(t)
	defer cleanup()

	// Create course.
	testCourse := models.Course{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryID: 1,
	}
	if err := db.CreateCourse(&testCourse); err != nil {
		t.Fatal(err)
	}

	// Create user.
	var user models.User
	if err := db.CreateUserFromRemoteIdentity(
		&user, &models.RemoteIdentity{},
	); err != nil {
		t.Fatal(err)
	}

	// Create enrollment.
	if err := db.CreateEnrollment(&models.Enrollment{
		UserID:   user.ID,
		CourseID: testCourse.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, testCourse.ID); err != nil {
		t.Fatal(err)
	}

	// Create group.
	group := models.Group{CourseID: testCourse.ID}
	if err := db.CreateGroup(&group); err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodDelete, route, web.DeleteGroup(db))

	requestURL := "/groups/" + strconv.FormatUint(group.ID, 10)
	r := httptest.NewRequest(http.MethodDelete, requestURL, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	// Prepare context with course request.
	router.Find(http.MethodDelete, requestURL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Fatal(err)
	}

	assertCode(t, w.Code, http.StatusOK)

}

func TestGetGroup(t *testing.T) {
	const (
		route = "/groups/:gid"
	)
	db, cleanup := setup(t)
	defer cleanup()

	// Create course.
	testCourse := models.Course{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryID: 1,
	}
	if err := db.CreateCourse(&testCourse); err != nil {
		t.Fatal(err)
	}

	// Create user.
	var user models.User
	if err := db.CreateUserFromRemoteIdentity(
		&user, &models.RemoteIdentity{},
	); err != nil {
		t.Fatal(err)
	}

	// Create enrollment.
	if err := db.CreateEnrollment(&models.Enrollment{
		UserID:   user.ID,
		CourseID: testCourse.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, testCourse.ID); err != nil {
		t.Fatal(err)
	}

	// Create group.
	group := models.Group{CourseID: testCourse.ID}
	if err := db.CreateGroup(&group); err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodDelete, route, web.GetGroup(db))

	requestURL := "/groups/" + strconv.FormatUint(group.ID, 10)
	r := httptest.NewRequest(http.MethodGet, requestURL, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	// Prepare context with course request.
	router.Find(http.MethodDelete, requestURL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Fatal(err)
	}
	assertCode(t, w.Code, http.StatusOK)

	var respGroup models.Group
	if err := json.Unmarshal(w.Body.Bytes(), &respGroup); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(respGroup, group) {
		t.Errorf("have response group %+v, while database has %+v", &respGroup, group)
	}
}

func TestPatchGroupStatus(t *testing.T) {
	const (
		route = "/groups/:gid"
		fake  = "fake"
	)
	db, cleanup := setup(t)
	defer cleanup()

	// Prepare course
	testCourse := models.Course{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryID: 1,
		ID:          1,
	}

	err := db.CreateCourse(&testCourse)
	if err != nil {
		t.Fatal(err)
	}

	var adminUser models.User
	adminUser.ID = 1
	adminRemote := &models.RemoteIdentity{ID: 1, UserID: 1, Provider: "fake", AccessToken: "", RemoteID: 2}

	var user models.User
	user.ID = 2
	userRemote := &models.RemoteIdentity{ID: 2, UserID: 2, Provider: "fake", AccessToken: "", RemoteID: 2}

	if err := db.CreateUserFromRemoteIdentity(
		&adminUser, adminRemote,
	); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateUserFromRemoteIdentity(
		&user, userRemote,
	); err != nil {
		t.Fatal(err)
	}

	// Create Enrollments
	userEnroll := &models.Enrollment{
		UserID:   user.ID,
		CourseID: testCourse.ID,
		GroupID:  1,
	}

	adminEnroll := &models.Enrollment{
		UserID:   adminUser.ID,
		CourseID: testCourse.ID,
		GroupID:  1,
	}

	if err := db.CreateEnrollment(userEnroll); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(adminEnroll); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(userEnroll.ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(adminEnroll.ID, 1); err != nil {
		t.Fatal(err)
	}

	createGroup := &models.Group{
		CourseID: testCourse.ID,
		ID:       1,
		Users:    []*models.User{&user, &adminUser},
	}
	err = db.CreateGroup(createGroup)
	if err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodPatch, route, web.PatchGroup(nullLogger(), db))

	// Send empty request, the user should not be modified.
	emptyJSON, err := json.Marshal(&web.UpdateGroupRequest{})
	if err != nil {
		t.Fatal(err)
	}
	requestBody := bytes.NewReader(emptyJSON)

	requestURL := "/groups/" + strconv.FormatUint(createGroup.ID, 10)
	r := httptest.NewRequest(http.MethodPatch, requestURL, requestBody)
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	f := scm.NewFakeSCMClient()
	if _, err := f.CreateDirectory(context.Background(), &scm.CreateDirectoryOptions{
		Name: testCourse.Code,
		Path: testCourse.Code,
	}); err != nil {
		t.Fatal(err)
	}
	c.Set(fake, f)
	// Prepare context with user request.
	c.Set("user", &adminUser)
	router.Find(http.MethodPatch, requestURL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}
	assertCode(t, w.Code, http.StatusOK)

	// Send request with name changed, the group name should change.
	trueJSON, err := json.Marshal(&web.UpdateGroupRequest{
		Status: 3,
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
	c.Set("user", &adminUser)
	fakeProvider, err := scm.NewSCMClient("fake", "token")
	if err != nil {
		t.Fatal(err)
	}
	fakeProvider.CreateDirectory(c.Request().Context(),
		&scm.CreateDirectoryOptions{Path: "path", Name: "name"},
	)
	c.Set("fake", fakeProvider)

	router.Find(http.MethodPatch, requestURL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}
	assertCode(t, w.Code, http.StatusOK)

	haveGroup, err := db.GetGroup(createGroup.ID)
	if err != nil {
		t.Fatal(err)
	}

	wantGroup := &models.Group{
		ID:     1,
		Status: 3,
		Users:  []*models.User{&user, &adminUser},
	}

	if !reflect.DeepEqual(wantGroup.Status, haveGroup.Status) {
		t.Errorf("have group %+v want %+v", haveGroup.Status, wantGroup.Status)
	}
}
