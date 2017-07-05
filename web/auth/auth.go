package auth

import (
	"encoding/gob"
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

func init() {
	gob.Register(&UserSession{})
}

// Frontend URLs.
const (
	logout = "/app/logout"
	login  = "/app/newlogin"
)

// Session keys.
const (
	SessionKey       = "session"
	GothicSessionKey = "_gothic_session"
	UserKey          = "user"
)

// UserSession holds user session information.
type UserSession struct {
	ID        uint64
	Providers map[string]struct{}
}

func newUserSession(id uint64) *UserSession {
	return &UserSession{
		ID:        id,
		Providers: make(map[string]struct{}),
	}
}

func (us *UserSession) enableProvider(provider string) {
	us.Providers[provider] = struct{}{}
}

func (us *UserSession) disableProvider(provider string) {
	delete(us.Providers, provider)
}

// OAuth2Logout invalidates the session for the logged in user.
func OAuth2Logout() echo.HandlerFunc {
	return func(c echo.Context) error {
		r := c.Request()
		w := c.Response()

		sess, err := session.Get(SessionKey, c)
		if err != nil {
			return sess.Save(r, w)
		}

		if i, ok := sess.Values[UserKey]; ok {
			us, ok := i.(*UserSession)
			if !ok {
				return c.Redirect(http.StatusFound, logout)
			}
			// Invalidate gothic user sessions.
			for provider := range us.Providers {
				sess, err := session.Get(provider+GothicSessionKey, c)
				if err != nil {
					return err
				}
				sess.Options.MaxAge = -1
				sess.Values = make(map[interface{}]interface{})
				sess.Save(r, w)
			}
		}

		// Invalidate our user session.
		sess.Options.MaxAge = -1
		sess.Values = make(map[interface{}]interface{})
		sess.Save(r, w)

		return c.Redirect(http.StatusFound, logout)
	}
}

// OAuth2Login tries to authenticate against an oauth2 provider.
func OAuth2Login(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		w := c.Response()
		r := c.Request()

		sess, err := session.Get(SessionKey, c)
		if err != nil {
			return sess.Save(r, w)
		}

		if i, ok := sess.Values[UserKey]; ok {
			us, ok := i.(*UserSession)
			if !ok {
				return OAuth2Logout()(c)
			}
			if _, err := db.GetUser(us.ID); err != nil {
				return OAuth2Logout()(c)
			}
			if _, ok := us.Providers[c.Param("provider")]; ok {
				// Provider has already been registered.
				return c.Redirect(http.StatusFound, login)
			}
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

		if sess.Values[UserKey] == nil {
			sess.Values[UserKey] = newUserSession(user.ID)
		}
		us := sess.Values[UserKey].(*UserSession)
		us.enableProvider(c.Param("provider"))
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

		sess, err := session.Get(SessionKey, c)
		if err != nil {
			return sess.Save(r, w)
		}

		if i, ok := sess.Values[UserKey]; ok {
			us, ok := i.(*UserSession)
			if !ok {
				return OAuth2Logout()(c)
			}
			if _, err := db.GetUser(us.ID); err != nil {
				return OAuth2Logout()(c)
			}
			if _, ok := us.Providers[c.Param("provider")]; ok {
				// Provider has already been registered.
				return c.Redirect(http.StatusFound, login)
			}
		}

		externalUser, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			return echo.ErrUnauthorized
		}

		user, err := getInteralUser(db, &externalUser)
		if err != nil {
			return err
		}

		if sess.Values[UserKey] == nil {
			sess.Values[UserKey] = newUserSession(user.ID)
		}
		us := sess.Values[UserKey].(*UserSession)
		us.enableProvider(c.Param("provider"))
		if err := sess.Save(r, w); err != nil {
			return err
		}

		return c.Redirect(http.StatusFound, login)
	}
}

// AccessControl returns an access control middleware. Given a valid context
// with sufficient access the next handler is called. Missing or invalid
// credentials results in a 401 unauthorized response.
func AccessControl(db database.Database, scms map[string]scm.SCM) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, err := session.Get(SessionKey, c)
			if err != nil {
				// Save fixes the session if it has been modified
				// or it is no longer valid due to change of keys.
				sess.Save(c.Request(), c.Response())
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}

			i, ok := sess.Values[UserKey]
			if !ok {
				return echo.ErrUnauthorized
			}

			us, ok := i.(*UserSession)
			if !ok {
				return echo.ErrUnauthorized
			}

			user, err := db.GetUser(us.ID)
			if err != nil {
				return err
			}

			// TODO: Check if the user is allowed to access the endpoint.
			c.Set("user", user)
			for _, remoteIdentity := range user.RemoteIdentities {
				if _, ok := scms[remoteIdentity.AccessToken]; !ok {
					client, err := scm.NewSCMClient(remoteIdentity.Provider, remoteIdentity.AccessToken)
					if err != nil {
						return err
					}
					scms[remoteIdentity.AccessToken] = client
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
	case "github", "gitlab":
		remoteID, err := strconv.ParseUint(externalUser.UserID, 10, 64)
		if err != nil {
			return nil, err
		}

		user, err := db.GetUserByRemoteIdentity(provider.Name(), remoteID, externalUser.AccessToken)
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
