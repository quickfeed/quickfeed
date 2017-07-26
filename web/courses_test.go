package web_test

import (
	"bytes"
	"encoding/gob"
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
		DirectoryID: 1,
	}, {
		Name:        "New Systems",
		Code:        "DATx20",
		Year:        2019,
		Tag:         "Fall",
		Provider:    "fake",
		DirectoryID: 1,
	},
}

func TestListCourses(t *testing.T) {
	const listCoursesURL = "/courses"

	db, cleanup := setup(t)
	defer cleanup()

	for _, course := range allCourses {
		err := db.CreateCourse(course)
		if err != nil {
			t.Fatal(err)
		}
	}

	r := httptest.NewRequest(http.MethodGet, listCoursesURL, nil)
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
		if !reflect.DeepEqual(course, allCourses[i]) {
			t.Errorf("have course %+v want %+v", course, allCourses[i])
		}
	}

	assertCode(t, w.Code, http.StatusOK)
}

func TestNewCourse(t *testing.T) {
	const (
		newCoursesURL = "/courses"
		provider      = "fake"
	)

	db, cleanup := setup(t)
	defer cleanup()

	testCourse := *allCourses[0]

	// Convert course to course request, this allows us to verify that the
	// course we get from the database is correct.
	cr := courseToRequest(t, &testCourse)

	b, err := json.Marshal(cr)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, newCoursesURL, bytes.NewReader(b))
	r.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(r, w)
	c.Set(provider, &scm.FakeSCM{})

	h := web.NewCourse(nullLogger(), db)
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

	assertCode(t, w.Code, http.StatusCreated)
}

func TestListCoursesWithEnrollment(t *testing.T) {
	const (
		userCoursesRoute = "/users/:uid/courses"

		secret   = "123"
		provider = "github"
		remoteID = 11
	)
	var (
		pending  = int(models.Pending)
		accepted = int(models.Accepted)
		rejected = int(models.Rejected)
		none     = models.None
	)

	db, cleanup := setup(t)
	defer cleanup()

	var course1 models.Course
	if err := db.CreateCourse(&course1); err != nil {
		t.Fatal(err)
	}

	var course2 models.Course
	if err := db.CreateCourse(&course2); err != nil {
		t.Fatal(err)
	}

	var course3 models.Course
	if err := db.CreateCourse(&course3); err != nil {
		t.Fatal(err)
	}

	var course4 models.Course
	if err := db.CreateCourse(&course4); err != nil {
		t.Fatal(err)
	}

	user, err := db.CreateUserFromRemoteIdentity(provider, remoteID, secret)
	if err != nil {
		t.Fatal(err)
	}

	enrollment1 := models.Enrollment{
		UserID:   user.ID,
		CourseID: course1.ID,
	}
	enrollment2 := models.Enrollment{
		UserID:   user.ID,
		CourseID: course2.ID,
	}
	enrollment3 := models.Enrollment{
		UserID:   user.ID,
		CourseID: course3.ID,
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
	if err := db.AcceptEnrollment(enrollment3.ID); err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodGet, userCoursesRoute, web.ListCoursesWithEnrollment(db))

	userCoursesURL := "/users/" + strconv.FormatUint(user.ID, 10) + "/courses"
	r := httptest.NewRequest(http.MethodGet, userCoursesURL, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	// Prepare context with user request.
	router.Find(http.MethodGet, userCoursesURL, c)

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
		{ID: course1.ID, Enrolled: &pending},
		{ID: course2.ID, Enrolled: &rejected},
		{ID: course3.ID, Enrolled: &accepted},
		{ID: course4.ID, Enrolled: &none},
	}
	if !reflect.DeepEqual(courses, wantCourses) {
		t.Errorf("have course %+v want %+v", courses, wantCourses)
	}
}

func TestListCoursesWithEnrollmentStatuses(t *testing.T) {
	const (
		query            = "?status=accepted,rejected"
		userCoursesRoute = "/users/:uid/courses" + query
		secret           = "123"
		provider         = "github"
		remoteID         = 11
	)
	var (
		accepted = int(models.Accepted)
		rejected = int(models.Rejected)
	)

	db, cleanup := setup(t)
	defer cleanup()

	var course1 models.Course
	if err := db.CreateCourse(&course1); err != nil {
		t.Fatal(err)
	}

	var course2 models.Course
	if err := db.CreateCourse(&course2); err != nil {
		t.Fatal(err)
	}

	var course3 models.Course
	if err := db.CreateCourse(&course3); err != nil {
		t.Fatal(err)
	}

	var course4 models.Course
	if err := db.CreateCourse(&course4); err != nil {
		t.Fatal(err)
	}

	user, err := db.CreateUserFromRemoteIdentity(provider, remoteID, secret)
	if err != nil {
		t.Fatal(err)
	}

	enrollment1 := models.Enrollment{
		UserID:   user.ID,
		CourseID: course1.ID,
	}
	enrollment2 := models.Enrollment{
		UserID:   user.ID,
		CourseID: course2.ID,
	}
	enrollment3 := models.Enrollment{
		UserID:   user.ID,
		CourseID: course3.ID,
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
	if err := db.AcceptEnrollment(enrollment3.ID); err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodGet, userCoursesRoute, web.ListCoursesWithEnrollment(db))

	userCoursesURL := "/users/" + strconv.FormatUint(user.ID, 10) + "/courses" + query
	r := httptest.NewRequest(http.MethodGet, userCoursesURL, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	// Prepare context with user request.
	router.Find(http.MethodGet, userCoursesURL, c)

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
		{ID: course2.ID, Enrolled: &rejected},
		{ID: course3.ID, Enrolled: &accepted},
	}
	if !reflect.DeepEqual(courses, wantCourses) {
		t.Errorf("have course %+v want %+v", courses, wantCourses)
	}

}

func TestGetCourse(t *testing.T) {
	const getCourseRoute = "/courses/:cid"
	courseURL := "/courses/" + strconv.FormatUint(allCourses[0].ID, 10)

	db, cleanup := setup(t)
	defer cleanup()

	for _, course := range allCourses {
		err := db.CreateCourse(course)
		if err != nil {
			t.Fatal(err)
		}
	}

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodGet, getCourseRoute, web.GetCourse(db))

	r := httptest.NewRequest(http.MethodGet, courseURL, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	// Prepare context with course request.
	router.Find(http.MethodGet, courseURL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Fatal(err)
	}

	var foundCourse *models.Course
	if err := json.Unmarshal(w.Body.Bytes(), &foundCourse); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(foundCourse, allCourses[0]) {
		t.Errorf("have course %+v want %+v", foundCourse, allCourses[0])
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
