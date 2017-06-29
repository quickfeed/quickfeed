package auth

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/autograde/aguis/database"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

// User session keys.
const (
	UserSession = "session"
	UserID      = "userid"
)

// OAuth2Logout invalidates the session for the logged in user.
func OAuth2Logout() echo.HandlerFunc {
	return func(c echo.Context) error {
		r := c.Request()
		w := c.Response()

		// Invalidate our user session.
		sess, _ := session.Get(UserSession, c)
		sess.Options.MaxAge = -1
		sess.Values = make(map[interface{}]interface{})
		sess.Save(r, w)

		// TODO: Get correct provider from user session and move
		// /:proivder/logout to /logout.

		// Invalidate gothic user session.
		gothic.Logout(w, r)

		return c.Redirect(http.StatusFound, "/")
	}
}

// OAuth2Login tries to authenticate against an oauth2 provider.
func OAuth2Login(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		w := c.Response()
		r := c.Request()

		sess, err := session.Get(UserSession, c)
		if err != nil {
			return sess.Save(r, w)
		}
		if _, ok := sess.Values[UserID]; ok {
			return c.Redirect(http.StatusFound, "/")
		}

		externalUser, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			url, err := gothic.GetAuthURL(w, r)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			return c.Redirect(http.StatusTemporaryRedirect, url)
		}

		user, err := getInteralUser(db, &externalUser)
		if err != nil {
			return err
		}

		sess.Values[UserID] = user.ID
		if err := sess.Save(r, w); err != nil {
			return err
		}

		return c.Redirect(http.StatusFound, "/")
	}
}

// OAuth2Callback handles the callback from an oauth2 provider.
func OAuth2Callback(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		w := c.Response()
		r := c.Request()

		sess, err := session.Get(UserSession, c)
		if err != nil {
			return sess.Save(r, w)
		}
		if _, ok := sess.Values[UserID]; ok {
			return c.Redirect(http.StatusFound, "/")
		}

		externalUser, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			return echo.ErrUnauthorized
		}

		user, err := getInteralUser(db, &externalUser)
		if err != nil {
			return err
		}

		sess.Values[UserID] = user.ID
		if err := sess.Save(r, w); err != nil {
			return err
		}

		return c.Redirect(http.StatusFound, "/")
	}
}

// AccessControl returns an AccessControl middleware. Given a valid context with
// sufficient access the next handler is called. Missing or invalid credentials
// results in a 401 unauthorized response.
func AccessControl(db database.Database) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, err := session.Get(UserSession, c)
			if err != nil {
				// Save fixes the session if it has been modified
				// or it is no longer valid due to change of keys.
				sess.Save(c.Request(), c.Response())
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}

			userID, ok := sess.Values[UserID]
			if !ok {
				return echo.ErrUnauthorized
			}

			user, err := db.GetUser(userID.(int))
			if err != nil {
				return err
			}

			// TODO: Check if the user is allowed to access the endpoint.
			c.Set("user", user)

			return next(c)
		}
	}
}

func getInteralUser(db database.Database, externalUser *goth.User) (*database.User, error) {
	provider, err := goth.GetProvider(externalUser.Provider)

	if err != nil {
		return nil, err
	}

	// TODO: Extract each case into a function so that they can be tested.
	switch provider.Name() {
	case "github":
		githubID, err := strconv.Atoi(externalUser.UserID)
		if err != nil {
			return nil, err
		}
		user, err := db.GetUserWithGithubID(githubID, externalUser.AccessToken)
		if err != nil {
			return nil, err
		}
		return user, nil
	case "faux": // Provider is only registered and reachable from tests.
		return &database.User{}, nil
	default:
		return nil, errors.New("provider not implemented")
	}
}
