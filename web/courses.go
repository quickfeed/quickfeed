package web

import (
	"context"
	"net/http"
	"time"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"
	"github.com/labstack/echo"
)

// MaxWait is the maximum time a request is allowed to stay open before
// aborting.
const MaxWait = 10 * time.Second

// NewCourseRequest represents a request for a new course.
type NewCourseRequest struct {
	Name         string `json:"name"`
	Organization int    `json:"organization"`
}

func (cr *NewCourseRequest) valid() bool {
	return cr != nil && cr.Name != "" && cr.Organization != 0
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

		// TODO: Should get provider from request.
		s := c.Get("github").(scm.SCM)
		ctx, cancel := context.WithTimeout(c.Request().Context(), MaxWait)
		defer cancel()
		directory, err := s.GetDirectory(ctx, cr.Organization)
		if err != nil {
			return err
		}

		// TODO: Does the user have sufficient rights?
		// TODO: Initialize organization?

		// TODO: Should take directory.ID not name.
		if err := db.CreateCourse(cr.Name, directory.Name); err != nil {
			return err
		}

		return c.NoContent(http.StatusCreated)
	}
}
