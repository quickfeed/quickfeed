package web_test

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/scm"
	"github.com/autograde/aguis/web"
	"github.com/autograde/aguis/web/auth"
	"github.com/labstack/echo"
	_ "github.com/mattn/go-sqlite3"
)

var allCourses = []*models.Course{
	{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryID: 1,
	},
	{
		Name:        "Operating Systems",
		Code:        "DAT320",
		Year:        2017,
		Tag:         "Fall",
		Provider:    "fake",
		DirectoryID: 2,
	}, {
		Name:        "New Systems",
		Code:        "DATx20",
		Year:        2019,
		Tag:         "Fall",
		Provider:    "fake",
		DirectoryID: 3,
	}, {
		Name:        "Hyped Systems",
		Code:        "DATx20",
		Year:        2019,
		Tag:         "Fall",
		Provider:    "fake",
		DirectoryID: 4,
	},
}

func TestListCourses(t *testing.T) {
	const route = "/courses"

	db, cleanup := setup(t)
	defer cleanup()

	var testCourses []*models.Course
	for _, course := range allCourses {
		testCourse := *course
		err := db.CreateCourse(&testCourse)
		if err != nil {
			t.Fatal(err)
		}
		testCourses = append(testCourses, &testCourse)
	}

	r := httptest.NewRequest(http.MethodGet, route, nil)
	w := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(r, w)

	coursesHandler := web.ListCourses(db)
	if err := coursesHandler(c); err != nil {
		t.Fatal(err)
	}

	var foundCourses []*models.Course
	if err := json.Unmarshal(w.Body.Bytes(), &foundCourses); err != nil {
		t.Fatal(err)
	}

	for i, course := range foundCourses {
		if !reflect.DeepEqual(course, testCourses[i]) {
			t.Errorf("have course %+v want %+v", course, testCourses[i])
		}
	}

	assertCode(t, w.Code, http.StatusOK)
}

func TestNewCourse(t *testing.T) {
	const (
		route = "/courses"
		fake  = "fake"
	)

	db, cleanup := setup(t)
	defer cleanup()

	var user models.User
	if err := db.CreateUserFromRemoteIdentity(
		&user, &models.RemoteIdentity{
			Provider: fake,
		},
	); err != nil {
		t.Fatal(err)
	}

	testCourse := *allCourses[0]

	// Convert course to course request, this allows us to verify that the
	// course we get from the database is correct.
	cr := courseToRequest(t, &testCourse)

	b, err := json.Marshal(cr)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, route, bytes.NewReader(b))
	r.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(r, w)
	f := scm.NewFakeSCMClient()
	if _, err := f.CreateDirectory(context.Background(), &scm.CreateDirectoryOptions{
		Name: testCourse.Code,
		Path: testCourse.Code,
	}); err != nil {
		t.Fatal(err)
	}
	c.Set(fake, f)
	c.Set(auth.UserKey, &models.User{ID: user.ID})

	h := web.NewCourse(nullLogger(), db, &web.BaseHookOptions{})
	if err := h(c); err != nil {
		t.Fatal(err)
	}

	var respCourse models.Course
	if err := json.Unmarshal(w.Body.Bytes(), &respCourse); err != nil {
		t.Fatal(err)
	}

	course, err := db.GetCourse(respCourse.ID)
	if err != nil {
		t.Fatal(err)
	}

	testCourse.ID = respCourse.ID
	if !reflect.DeepEqual(course, &testCourse) {
		t.Errorf("have database course %+v want %+v", course, &testCourse)
	}

	if !reflect.DeepEqual(&respCourse, course) {
		t.Errorf("have response course %+v want %+v", &respCourse, course)
	}

	if len(f.Hooks) != 4 {
		t.Errorf("have %d hooks want %d", len(f.Hooks), 4)
	}

	assertCode(t, w.Code, http.StatusCreated)
}

func TestEnrollmentProcess(t *testing.T) {
	const (
		route = "/courses/:cid/users/:uid"

		github = "github"
		gitlab = "gitlab"
	)

	db, cleanup := setup(t)
	defer cleanup()

	// Create course.
	testCourse := *allCourses[0]
	if err := db.CreateCourse(&testCourse); err != nil {
		t.Fatal(err)
	}
	// Create admin.
	var admin models.User
	if err := db.CreateUserFromRemoteIdentity(
		&admin, &models.RemoteIdentity{
			Provider: github,
		},
	); err != nil {
		t.Fatal(err)
	}
	// Create user.
	var user models.User
	if err := db.CreateUserFromRemoteIdentity(
		&user, &models.RemoteIdentity{
			Provider: gitlab,
		},
	); err != nil {
		t.Fatal(err)
	}

	// Prepare request payload.
	b, err := json.Marshal(&web.EnrollUserRequest{})
	if err != nil {
		t.Fatal(err)
	}
	requestBody := bytes.NewReader(b)

	e := echo.New()
	router := echo.NewRouter(e)
	requestURL := fmt.Sprintf("/courses/%d/users/%d", testCourse.ID, user.ID)

	// Add the route to handler.
	router.Add(http.MethodPut, route, web.CreateEnrollment(db))
	r := httptest.NewRequest(http.MethodPut, requestURL, requestBody)
	r.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	// Prepare context with user request.
	c.Set(auth.UserKey, &user)
	router.Find(http.MethodPut, requestURL, c)

	// Invoke the prepared handler. This will attempt to create an
	// enrollment for the user in the chosen course.
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}
	assertCode(t, w.Code, http.StatusCreated)

	// Verify that an appropriate enrollment was indeed created.
	pendingEnrollment, err := db.GetEnrollmentByCourseAndUser(testCourse.ID, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment := &models.Enrollment{
		ID:       pendingEnrollment.ID,
		CourseID: testCourse.ID,
		UserID:   user.ID,
	}
	if !reflect.DeepEqual(pendingEnrollment, wantEnrollment) {
		t.Errorf("have enrollment\n %+v\n want\n %+v", pendingEnrollment, wantEnrollment)
	}

	// Prepare request payload.
	b, err = json.Marshal(&web.EnrollUserRequest{
		Status: models.Student,
	})
	if err != nil {
		t.Fatal(err)
	}
	requestBody.Reset(b)

	e = echo.New()
	router = echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodPatch, route, web.UpdateEnrollment(db))
	r = httptest.NewRequest(http.MethodPatch, requestURL, requestBody)
	r.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	qv := r.URL.Query()
	qv.Set("status", "student")
	r.URL.RawQuery = qv.Encode()
	w = httptest.NewRecorder()
	c.Reset(r, w)
	// Prepare context with user request.
	c.Set(auth.UserKey, &user)
	router.Find(http.MethodPatch, requestURL, c)

	// Invoke the prepared handler. This will attempt to accept the
	// previously created enrollment. This should fail with a 401
	// Unauthorized as the user is not an administrator.
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}
	assertCode(t, w.Code, http.StatusUnauthorized)

	requestBody.Reset(b)
	w = httptest.NewRecorder()
	c.Reset(r, w)
	c.Set(auth.UserKey, &admin)
	router.Find(http.MethodPatch, requestURL, c)

	// Invoke the prepared handler. This will attempt to accept the
	// previously created enrollment. This should succeed with a 200 OK as
	// the current user is an administrator.
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}
	assertCode(t, w.Code, http.StatusOK)

	// Verify that the enrollment have been accepted.
	acceptedEnrollment, err := db.GetEnrollmentByCourseAndUser(testCourse.ID, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment.Status = models.Student
	if !reflect.DeepEqual(acceptedEnrollment, wantEnrollment) {
		t.Errorf("have enrollment %+v want %+v", acceptedEnrollment, wantEnrollment)
	}
}

func TestListCoursesWithEnrollment(t *testing.T) {
	const route = "/users/:uid/courses"

	db, cleanup := setup(t)
	defer cleanup()

	var testCourses []*models.Course
	for _, course := range allCourses {
		testCourse := *course
		err := db.CreateCourse(&testCourse)
		if err != nil {
			t.Fatal(err)
		}
		testCourses = append(testCourses, &testCourse)
	}

	var user models.User
	if err := db.CreateUserFromRemoteIdentity(
		&user, &models.RemoteIdentity{},
	); err != nil {
		t.Fatal(err)
	}

	enrollment1 := models.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[0].ID,
	}
	enrollment2 := models.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[1].ID,
	}
	enrollment3 := models.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[2].ID,
	}
	if err := db.CreateEnrollment(&enrollment1); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&enrollment2); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&enrollment3); err != nil {
		t.Fatal(err)
	}
	if err := db.RejectEnrollment(enrollment2.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(enrollment3.ID); err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodGet, route, web.ListCoursesWithEnrollment(db))

	requestURL := "/users/" + strconv.FormatUint(user.ID, 10) + "/courses"
	r := httptest.NewRequest(http.MethodGet, requestURL, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	// Prepare context with user request.
	router.Find(http.MethodGet, requestURL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}

	var courses []*models.Course
	if err := json.Unmarshal(w.Body.Bytes(), &courses); err != nil {
		t.Fatal(err)
	}

	assertCode(t, w.Code, http.StatusOK)
	wantCourses := []*models.Course{
		{ID: testCourses[0].ID, Enrolled: int(models.Pending)},
		{ID: testCourses[1].ID, Enrolled: int(models.Rejected)},
		{ID: testCourses[2].ID, Enrolled: int(models.Student)},
		{ID: testCourses[3].ID, Enrolled: models.None},
	}
	for i := range courses {
		if courses[i].ID != wantCourses[i].ID {
			t.Errorf("have course %+v want %+v", courses[i].ID, wantCourses[i].ID)
		}
		if courses[i].Enrolled != wantCourses[i].Enrolled {
			t.Errorf("have course %+v want %+v", courses[i].Enrolled, wantCourses[i].Enrolled)
		}
	}
}

func TestListCoursesWithEnrollmentStatuses(t *testing.T) {
	const (
		query = "?status=student,rejected"
		route = "/users/:uid/courses" + query
	)

	db, cleanup := setup(t)
	defer cleanup()

	var testCourses []*models.Course
	for _, course := range allCourses {
		testCourse := *course
		err := db.CreateCourse(&testCourse)
		if err != nil {
			t.Fatal(err)
		}
		testCourses = append(testCourses, &testCourse)
	}

	var user models.User
	if err := db.CreateUserFromRemoteIdentity(
		&user, &models.RemoteIdentity{},
	); err != nil {
		t.Fatal(err)
	}

	enrollment1 := models.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[0].ID,
	}
	enrollment2 := models.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[1].ID,
	}
	enrollment3 := models.Enrollment{
		UserID:   user.ID,
		CourseID: testCourses[2].ID,
	}
	if err := db.CreateEnrollment(&enrollment1); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&enrollment2); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&enrollment3); err != nil {
		t.Fatal(err)
	}
	if err := db.RejectEnrollment(enrollment2.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(enrollment3.ID); err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodGet, route, web.ListCoursesWithEnrollment(db))

	requestURL := "/users/" + strconv.FormatUint(user.ID, 10) + "/courses" + query
	r := httptest.NewRequest(http.MethodGet, requestURL, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	// Prepare context with user request.
	router.Find(http.MethodGet, requestURL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}

	var courses []*models.Course
	if err := json.Unmarshal(w.Body.Bytes(), &courses); err != nil {
		t.Fatal(err)
	}

	assertCode(t, w.Code, http.StatusOK)
	wantCourses, err := db.GetCoursesByUser(user.ID, models.Rejected, models.Student)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(courses, wantCourses) {
		t.Errorf("have course %+v want %+v", courses, wantCourses)
	}

}

func TestGetCourse(t *testing.T) {
	const route = "/courses/:cid"

	db, cleanup := setup(t)
	defer cleanup()

	var course models.Course
	err := db.CreateCourse(&course)
	if err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodGet, route, web.GetCourse(db))

	requestURL := "/courses/" + strconv.FormatUint(course.ID, 10)
	r := httptest.NewRequest(http.MethodGet, requestURL, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	// Prepare context with course request.
	router.Find(http.MethodGet, requestURL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Fatal(err)
	}

	var foundCourse models.Course
	if err := json.Unmarshal(w.Body.Bytes(), &foundCourse); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(&foundCourse, &course) {
		t.Errorf("have course %+v want %+v", &foundCourse, &course)
	}

	assertCode(t, w.Code, http.StatusOK)
}

func TestDeleteGroup(t *testing.T) {
	const (
		route  = "/groups/:gid"
		gitlab = "gitlab"
	)
	db, cleanup := setup(t)
	defer cleanup()

	// Create course.
	testCourse := *allCourses[0]
	if err := db.CreateCourse(&testCourse); err != nil {
		t.Fatal(err)
	}

	// Create user.
	var user models.User
	if err := db.CreateUserFromRemoteIdentity(
		&user, &models.RemoteIdentity{
			Provider: gitlab,
		},
	); err != nil {
		t.Fatal(err)
	}

	// create enrollment
	enrollment := models.Enrollment{
		UserID:   user.ID,
		CourseID: testCourse.ID,
	}

	if err := db.CreateEnrollment(&enrollment); err != nil {
		t.Fatal(err)
	}

	if err := db.AcceptEnrollment(enrollment.ID); err != nil {
		t.Fatal(err)
	}

	// create group
	group := models.Group{
		Name:     "group1",
		Status:   models.Pending,
		CourseID: testCourse.ID,
		// Users:    []*models.User{&user},
	}
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

func courseToRequest(t *testing.T, course *models.Course) (cr web.NewCourseRequest) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(course); err != nil {
		t.Fatal(err)
	}
	dec := gob.NewDecoder(&b)
	if err := dec.Decode(&cr); err != nil {
		t.Fatal(err)
	}
	return
}
