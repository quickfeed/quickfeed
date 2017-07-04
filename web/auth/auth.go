package auth

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/scm"
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

// Frontend URLs.
const (
	logout = "/app/logout"
	login  = "/app/newlogin"
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

		return c.Redirect(http.StatusFound, logout)
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

		if userID, ok := sess.Values[UserID]; ok {
			if _, err := db.GetUser(userID.(uint64)); err != nil {
				return OAuth2Logout()(c)
			}
			return c.Redirect(http.StatusFound, login)
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

		return c.Redirect(http.StatusFound, login)
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

		if userID, ok := sess.Values[UserID]; ok {
			if _, err := db.GetUser(userID.(uint64)); err != nil {
				return OAuth2Logout()(c)
			}
			return c.Redirect(http.StatusFound, login)
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

		return c.Redirect(http.StatusFound, login)
	}
}

// AccessControl returns an AccessControl middleware. Given a valid context with
// sufficient access the next handler is called. Missing or invalid credentials
// results in a 401 unauthorized response.
func AccessControl(db database.Database, scms map[string]scm.SCM) echo.MiddlewareFunc {
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

			user, err := db.GetUser(userID.(uint64))
			if err != nil {
				return err
			}

			// TODO: Check if the user is allowed to access the endpoint.
			c.Set("user", user)
			for _, remoteIdentity := range user.RemoteIdentities {
				if _, ok := scms[remoteIdentity.AccessToken]; !ok {
					scms[remoteIdentity.AccessToken] = scm.NewGithubSCMClient(remoteIdentity.AccessToken)
				}
				c.Set(remoteIdentity.Provider, scms[remoteIdentity.AccessToken])
			}

			return next(c)
		}
	}
}

func getInteralUser(db database.Database, externalUser *goth.User) (*models.User, error) {
	provider, err := goth.GetProvider(externalUser.Provider)

	if err != nil {
		return nil, err
	}

	// TODO: Extract each case into a function so that they can be tested.
	switch provider.Name() {
	case "github":
		githubID, err := strconv.ParseUint(externalUser.UserID, 10, 64)
		if err != nil {
			return nil, err
		}

		user, err := db.GetUserByRemoteIdentity(provider.Name(), githubID, externalUser.AccessToken)
		if err != nil {
			return nil, err
		}
		return user, nil
	case "faux": // Provider is only registered and reachable from tests.
		return &models.User{}, nil
	default:
		return nil, errors.New("provider not implemented")
	}
}
