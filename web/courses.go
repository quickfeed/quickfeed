package web

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/scm"
	"github.com/labstack/echo"
)

// MaxWait is the maximum time a request is allowed to stay open before
// aborting.
const MaxWait = 60 * time.Second

// NewCourseRequest represents a request for a new course.
type NewCourseRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
	Year uint   `json:"year"`
	Tag  string `json:"tag"`

	Provider    string `json:"provider"`
	DirectoryID uint64 `json:"directoryid"`
}

func (cr *NewCourseRequest) valid() bool {
	return cr != nil &&
		cr.Name != "" &&
		cr.Code != "" &&
		(cr.Provider == "github" || cr.Provider == "gitlab") &&
		cr.DirectoryID != 0 &&
		cr.Year != 0 &&
		cr.Tag != ""
}

// EnrollUserRequest represent a request for enrolling a user to a course
type EnrollUserRequest struct {
	UserID   uint64 `json:"userid"`
	CourseID uint64 `json:"courseid"`
}

func (eur *EnrollUserRequest) valid() bool {
	return eur.CourseID != 0 && eur.UserID != 0
}

// ListCourses returns a JSON object containing all the courses in the database.
func ListCourses(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO check that the user requesting "/courses?user=x" has sufficent privileges to access this page.
		// The session user should either be the same as "x" or a teacher.
		id, err := ParseUintParam(c.QueryParam("user"))
		if err != nil {
			return err
		}

		var courses *[]models.Course
		if id > 0 {
			courses, err = db.GetCoursesForUser(id)
		} else {
			courses, err = db.GetCourses()
		}
		if err != nil {
			return err
		}
		return c.JSONPretty(http.StatusOK, courses, "\t")
	}
}

// ListAssignments lists all the assignment found in a place
func ListAssignments(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO check if the user has right to show the assignments.
		// same as courses above, should not return to unauthorised users
		id, err := strconv.ParseUint(c.QueryParam("course"), 10, 64)
		if err != nil {
			return err
		}
		var assignments *[]models.Assignment
		assignments, err = db.GetAssignments(id)
		if err != nil {
			return err
		}
		return c.JSONPretty(http.StatusOK, assignments, "\t")
	}
}

// Default repository names.
const (
	InfoRepo       = "course-info"
	AssignmentRepo = "assignments"
	TestsRepo      = "tests"
	SolutionsRepo  = "solutions"
)

// NewCourse creates a new course and associates it with an organization.
func NewCourse(logger *logrus.Logger, db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		var cr NewCourseRequest
		if err := c.Bind(&cr); err != nil {
			return err
		}
		if !cr.valid() {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		if c.Get(cr.Provider) == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "provider "+cr.Provider+" not registered")
		}
		s := c.Get(cr.Provider).(scm.SCM)

		ctx, cancel := context.WithTimeout(c.Request().Context(), MaxWait)
		defer cancel()

		directory, err := s.GetDirectory(ctx, cr.DirectoryID)
		if err != nil {
			return err
		}

		var paths = []string{InfoRepo, AssignmentRepo, TestsRepo, SolutionsRepo}
		for _, path := range paths {
			repo, err := s.CreateRepository(
				ctx,
				&scm.CreateRepositoryOptions{
					Path:      path,
					Directory: directory},
			)
			if err != nil {
				return err
			}
			logger.WithField("repo", repo).Println("Created new repository")
		}

		course := models.Course{
			Name:        cr.Name,
			Code:        cr.Code,
			Year:        cr.Year,
			Tag:         cr.Tag,
			Provider:    cr.Provider,
			DirectoryID: directory.ID,
		}

		if err := db.CreateCourse(&course); err != nil {
			return err
		}

		return c.JSONPretty(http.StatusCreated, &course, "\t")
	}
}

// EnrollUser enrolls a user to a course
func EnrollUser(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		var eur EnrollUserRequest
		if err := c.Bind(&eur); err != nil {
			return err
		}

		if !eur.valid() {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		if err := db.EnrollUserInCourse(eur.UserID, eur.CourseID); err != nil {
			return nil
		}
		return c.NoContent(http.StatusCreated)
	}
}
