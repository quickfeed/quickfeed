package github

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/web"
	gh "github.com/google/go-github/github"
	"github.com/labstack/echo"
)

// ListOrganizations returns the result of
// GET api.github.com/user/memberships/orgs.
func ListOrganizations() echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*database.User)
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: user.AccessToken},
		)

		ctx, cancel := context.WithTimeout(c.Request().Context(), web.MaxWait)
		defer cancel()

		// TODO: Should be created (retrieved from cache) in access
		// control and added to the context. We should then be able to
		// test handlers without hitting the github API.
		tc := oauth2.NewClient(ctx, ts)
		client := gh.NewClient(tc)

		orgs, _, err := client.Organizations.ListOrgMemberships(ctx, nil)
		if err != nil {
			return err
		}

		return c.JSONPretty(http.StatusOK, orgs, "\t")
	}
}
