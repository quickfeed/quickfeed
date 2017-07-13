package web_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/web"
	"github.com/labstack/echo"
	_ "github.com/mattn/go-sqlite3"
)

var allCourses = []*models.Course{
	{
		ID:   100,
		Name: "Distributed Systems",
		Code: "DAT520",
		Year: 2018,
		Tag:  "Spring",
	},
	{
		ID:   101,
		Name: "Operating Systems",
		Code: "DAT320",
		Year: 2017,
		Tag:  "Fall",
	}, {
		ID:   102,
		Name: "New Systems",
		Code: "DATx20",
		Year: 2019,
		Tag:  "Fall",
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
