package web_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/autograde/aguis/models"
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
