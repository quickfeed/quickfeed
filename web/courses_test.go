package web_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/autograde/aguis/database"
	log "github.com/autograde/aguis/logger"
	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/web"
	"github.com/labstack/echo"
	_ "github.com/mattn/go-sqlite3"
)

const (
	listCoursesURL        = "/courses"
	listCoursesForUserURL = "/courses?user=1"
)

const (
	uID  = 1
	rID1 = 1

	secret1   = "123"
	provider1 = "github"
	remoteID1 = 10
)

const (
	cID1  = 100
	name1 = "Distributed Systems"
	code1 = "DAT520"
	y1    = 2018
	tag1  = "Spring"

	cID2  = 101
	name2 = "Operating Systems"
	code2 = "DAT320"
	y2    = 2017
	tag2  = "Fall"

	cID3  = 102
	name3 = "New Systems"
	code3 = "DATx20"
	y3    = 2019
	tag3  = "Fall"
)

var (
	allCourses = []*models.Course{
		&models.Course{
			ID:   cID1,
			Name: name1,
			Code: code1,
			Year: y1,
			Tag:  tag1,
		},
		&models.Course{
			ID:   cID2,
			Name: name2,
			Code: code2,
			Year: y2,
			Tag:  tag2,
		}, &models.Course{
			ID:   cID3,
			Name: name3,
			Code: code3,
			Year: y3,
			Tag:  tag3,
		},
	}
)

// Run with LOGDB=true go test -v to see database statements

func TestListCoursesInSystem(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	for _, course := range allCourses {
		err := db.CreateCourse(course)
		if err != nil {
			t.Fatal(err)
		}
	}
	courses, err := db.GetCourses()
	if err != nil {
		t.Fatal(err)
	}
	for i, c := range *courses {
		if !reflect.DeepEqual(c, *allCourses[i]) {
			t.Errorf("have course %+v want %+v", c, *allCourses[i])
		}
	}

	r := httptest.NewRequest(http.MethodGet, listCoursesURL, nil)
	w := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(r, w)

	coursesHandler := web.ListCourses(db)
	if err := coursesHandler(c); err != nil {
		t.Error(err)
	}
	var gotCourses []*models.Course
	json.Unmarshal(w.Body.Bytes(), &gotCourses)
	for i, c := range gotCourses {
		if !reflect.DeepEqual(c, allCourses[i]) {
			t.Errorf("have course %+v want %+v", c, allCourses[i])
		}
	}
	assertCode(t, w.Code, http.StatusOK)
}

func TestListCoursesForUserNoEnrolledCourses(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	_, err := db.NewUserFromRemoteIdentity(provider1, remoteID1, secret1)
	if err != nil {
		t.Fatal(err)
	}
	for _, course := range allCourses {
		err = db.CreateCourse(course)
		if err != nil {
			t.Fatal(err)
		}
	}
	courses, err := db.GetCourses()
	if err != nil {
		t.Fatal(err)
	}
	for i, c := range *courses {
		if !reflect.DeepEqual(c, *allCourses[i]) {
			t.Errorf("have course %+v want %+v", c, *allCourses[i])
		}
	}

	coursesForUser, err := db.GetCoursesForUser(uID)
	if err != nil {
		t.Fatal(err)
	}
	if len(*coursesForUser) > 0 {
		t.Errorf("got %d courses, want 0", len(*coursesForUser))
	}

	r := httptest.NewRequest(http.MethodGet, listCoursesForUserURL, nil)
	w := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(r, w)

	coursesHandler := web.ListCourses(db)
	if err := coursesHandler(c); err != nil {
		t.Error(err)
	}
	var gotCourses []*models.Course
	json.Unmarshal(w.Body.Bytes(), &gotCourses)
	if len(gotCourses) > 0 {
		t.Errorf("got %d courses, want 0", len(gotCourses))
	}
	assertCode(t, w.Code, http.StatusOK)
}

func TestListCoursesForUserWithTwoEnrolledCourses(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	_, err := db.NewUserFromRemoteIdentity(provider1, remoteID1, secret1)
	if err != nil {
		t.Fatal(err)
	}
	for _, course := range allCourses {
		err = db.CreateCourse(course)
		if err != nil {
			t.Fatal(err)
		}
	}

	// enroll user in two of the three courses
	wantCourses := []*models.Course{allCourses[0], allCourses[1]}
	for _, c := range wantCourses {
		err = db.EnrollUserInCourse(uID, c.ID)
		if err != nil {
			t.Fatal(err)
		}
	}
	// check that database returns exectly two of three courses for user
	coursesForUser, err := db.GetCoursesForUser(uID)
	if err != nil {
		t.Fatal(err)
	}
	if len(*coursesForUser) != len(wantCourses) {
		t.Errorf("got %d courses, want %d", len(*coursesForUser), len(wantCourses))
	}

	r := httptest.NewRequest(http.MethodGet, listCoursesForUserURL, nil)
	w := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(r, w)

	coursesHandler := web.ListCourses(db)
	if err := coursesHandler(c); err != nil {
		t.Error(err)
	}
	var gotCourses []*models.Course
	json.Unmarshal(w.Body.Bytes(), &gotCourses)
	for i, c := range gotCourses {
		if !reflect.DeepEqual(c, wantCourses[i]) {
			t.Errorf("have course %+v want %+v", c, wantCourses[i])
		}
	}
	assertCode(t, w.Code, http.StatusOK)
}

func setup(t *testing.T) (*database.GormDB, func()) {
	const (
		driver = "sqlite3"
		prefix = "testdb"
	)

	f, err := ioutil.TempFile(os.TempDir(), prefix)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	db, err := database.NewGormDB(driver, f.Name(), envSet("LOGDB"))
	if err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	return db, func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
		if err := os.Remove(f.Name()); err != nil {
			t.Error(err)
		}
	}
}

func assertCode(t *testing.T, haveCode, wantCode int) {
	if haveCode != wantCode {
		t.Errorf("have status code %d want %d", haveCode, wantCode)
	}
}

func envSet(env string) database.GormLogger {
	logger := logrus.New()
	logger.Formatter = log.NewDevFormatter(logger.Formatter)
	if os.Getenv(env) != "" {
		return database.Logger{Logger: logger}
	}
	return nil
}
