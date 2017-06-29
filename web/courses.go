package web

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/autograde/aguis/database"
	gh "github.com/google/go-github/github"
	"github.com/labstack/echo"
)

// MaxWait is the maximum time a request is allowed to stay open before
// aborting.
const MaxWait = 10 * time.Second

// NewCourseRequest represents a request for a new course.
type NewCourseRequest struct {
	Name         string `json:"name"`
	Organization string `json:"organization"`
}

func (cr *NewCourseRequest) valid() bool {
	return cr != nil && cr.Name != "" && cr.Organization != ""
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

		user := c.Get("user").(*database.User)
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: user.AccessToken},
		)

		ctx, cancel := context.WithTimeout(c.Request().Context(), MaxWait)
		defer cancel()

		// TODO: Should be created (retrieved from cache) in access
		// control and added to the context. We should then be able to
		// test handlers without hitting the github API.
		tc := oauth2.NewClient(ctx, ts)
		client := gh.NewClient(tc)

		_, _, err := client.Organizations.GetOrgMembership(ctx, "", cr.Organization)
		if err != nil {
			return err
		}

		// TODO: Does the user have sufficient rights?
		// TODO: Initialize organization?

		if err := db.CreateCourse(cr.Name, cr.Organization); err != nil {
			return err
		}

		return c.NoContent(http.StatusCreated)
	}
}
