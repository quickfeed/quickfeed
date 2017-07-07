package web

import (
	"context"
	"net/http"
	"time"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/scm"
	"github.com/labstack/echo"
)

// MaxWait is the maximum time a request is allowed to stay open before
// aborting.
const MaxWait = 10 * time.Second

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

// ListCourses returns a JSON object containing all the courses in the database.
func ListCourses(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		courses, err := db.GetCourses()
		if err != nil {
			return err
		}
		return c.JSONPretty(http.StatusOK, courses, "\t")
	}
}

// NewCourse creates a new course and associates it with an organization.
func NewCourse(db database.Database) echo.HandlerFunc {
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

		// Check that the directory exists.
		directory, err := s.GetDirectory(ctx, cr.DirectoryID)
		if err != nil {
			return err
		}

		// TODO: Does the user have sufficient rights?
		// TODO: Initialize directory?

		if err := db.CreateCourse(&models.Course{
			Name:        cr.Name,
			Code:        cr.Code,
			Year:        cr.Year,
			Tag:         cr.Tag,
			Provider:    cr.Provider,
			DirectoryID: directory.ID,
		}); err != nil {
			return err
		}

		return c.NoContent(http.StatusCreated)
	}
}
