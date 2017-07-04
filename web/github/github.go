package github

import (
	"context"
	"net/http"

	"github.com/autograde/aguis/scm"
	"github.com/autograde/aguis/web"
	"github.com/labstack/echo"
)

// ListOrganizations returns the result of
// GET api.github.com/user/memberships/orgs.
func ListOrganizations() echo.HandlerFunc {
	return func(c echo.Context) error {
		s := c.Get("github").(scm.SCM)
		ctx, cancel := context.WithTimeout(c.Request().Context(), web.MaxWait)
		defer cancel()

		directories, err := s.ListDirectories(ctx)
		if err != nil {
			return err
		}

		return c.JSONPretty(http.StatusOK, directories, "\t")
	}
}
