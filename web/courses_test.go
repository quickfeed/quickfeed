package web_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/web"
	"github.com/labstack/echo"
	_ "github.com/mattn/go-sqlite3"
)

const (
	listCoursesURL = "/courses?user=1"
)

func TestListCourses(t *testing.T) {
	const (
		uID  = 1
		rID1 = 1
		rID2 = 2

		secret1   = "123"
		provider1 = "github"
		remoteID1 = 10

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
		wantUser1 = &models.User{
			ID: uID,
			RemoteIdentities: []models.RemoteIdentity{{
				ID:          rID1,
				Provider:    provider1,
				RemoteID:    remoteID1,
				AccessToken: secret1,
				UserID:      uID,
			}},
		}

		wantCourse1 = &models.Course{
			ID:   cID1,
			Name: name1,
			Code: code1,
			Year: y1,
			Tag:  tag1,
		}

		wantCourse2 = &models.Course{
			ID:   cID2,
			Name: name2,
			Code: code2,
			Year: y2,
			Tag:  tag2,
		}

		wantCourse3 = &models.Course{
			ID:   cID3,
			Name: name3,
			Code: code3,
			Year: y3,
			Tag:  tag3,
		}
	)

	db, cleanup := setup(t)
	defer cleanup()

	user1, err := db.NewUserFromRemoteIdentity(provider1, remoteID1, secret1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(user1, wantUser1) {
		t.Errorf("have user %+v want %+v", user1, wantUser1)
	}

	err = db.CreateCourse(wantCourse1)
	if err != nil {
		t.Fatal(err)
	}
	err = db.CreateCourse(wantCourse2)
	if err != nil {
		t.Fatal(err)
	}
	err = db.CreateCourse(wantCourse3)
	if err != nil {
		t.Fatal(err)
	}
	courses, err := db.GetCourses()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("courses: %v \n", courses)

	courseForUser, err := db.GetCoursesForUser(uID)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("courseForUser: %v \n", courseForUser)

	err = db.EnrollUserInCourse(uID, cID1)
	if err != nil {
		t.Fatal(err)
	}
	courseForUser, err = db.GetCoursesForUser(uID)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("courseForUser: %v \n", courseForUser)
	err = db.EnrollUserInCourse(uID, cID2)
	if err != nil {
		t.Fatal(err)
	}
	courseForUser, err = db.GetCoursesForUser(uID)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("courseForUser: %v \n", courseForUser)

	// if !reflect.DeepEqual(course1, wantCourse1) {
	// 	t.Errorf("have course %+v want %+v", course1, wantCourse1)
	// }

	r := httptest.NewRequest(http.MethodGet, listCoursesURL, nil)
	w := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(r, w)

	coursesHandler := web.ListCourses(db)
	if err := coursesHandler(c); err != nil {
		t.Error(err)
	}
	fmt.Printf("list courses: %v %v\n", w.Body, w.Header())

	// location := w.Header().Get("Location")
	// if location != loginRedirect {
	// 	t.Errorf("have Location '%v' want '%v'", location, loginRedirect)
	// }
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
	if os.Getenv(env) != "" {
		return database.Logger{Logger: logrus.New()}
	}
	return nil
}
