package auth

import (
	"encoding/gob"
	"net/http"
	"strconv"
	"strings"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/scm"
	"github.com/autograde/aguis/web"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/markbates/goth/gothic"
)

func init() {
	gob.Register(&UserSession{})
}

// GetCallbackURL returns the callback URL for a given base URL and a provider.
func GetCallbackURL(baseURL, provider string) string {
	return web.GetProviderURL(baseURL, "auth", provider, "callback")
}

// Session keys.
const (
	SessionKey       = "session"
	GothicSessionKey = "_gothic_session"
	UserKey          = "user"
)

// Query keys.
const (
	State    = "state" // As defined by the OAuth2 RFC.
	Redirect = "redirect"
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
			// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
			us := i.(*UserSession)
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

		return c.Redirect(http.StatusFound, extractRedirectURL(r, Redirect))
	}
}

// PreAuth checks the current user session and executes the next handler if none
// was found for the given provider.
func PreAuth(db database.Database) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, err := session.Get(SessionKey, c)
			if err != nil {
				if err := sess.Save(c.Request(), c.Response()); err != nil {
					return err
				}
				return next(c)
			}

			if i, ok := sess.Values[UserKey]; ok {
				// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
				us := i.(*UserSession)
				if _, err := db.GetUser(us.ID); err != nil {
					return OAuth2Logout()(c)
				}
			}

			return next(c)
		}
	}
}

// OAuth2Login tries to authenticate against an oauth2 provider.
func OAuth2Login(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		w := c.Response()
		r := c.Request()

		provider, err := gothic.GetProviderName(r)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		var teacher int
		if strings.HasSuffix(provider, TeacherSuffix) {
			teacher = 1
		}

		qv := r.URL.Query()
		redirect := extractRedirectURL(r, Redirect)
		// TODO: Add a random string to protect against CSRF.
		qv.Set(State, strconv.Itoa(teacher)+redirect)
		r.URL.RawQuery = qv.Encode()

		url, err := gothic.GetAuthURL(w, r)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		// Redirect to provider to perform authentication.
		return c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

// OAuth2Callback handles the callback from an oauth2 provider.
func OAuth2Callback(db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		w := c.Response()
		r := c.Request()

		qv := r.URL.Query()
		redirect, teacher := extractState(r, State)
		provider, err := gothic.GetProviderName(r)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		// Add teacher suffix if upgrading scope.
		if teacher {
			qv.Set("provider", provider+TeacherSuffix)
		}
		r.URL.RawQuery = qv.Encode()

		// Complete authentication.
		externalUser, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		remoteID, err := strconv.ParseUint(externalUser.UserID, 10, 64)
		if err != nil {
			return err
		}

		sess, err := session.Get(SessionKey, c)
		if err != nil {
			return err
		}

		// Try to get already logged in user.
		if sess.Values[UserKey] != nil {
			i, ok := sess.Values[UserKey]
			if !ok {
				return OAuth2Logout()(c)
			}
			// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
			us := i.(*UserSession)

			// Associate user with remote identity.
			if err := db.AssociateUserWithRemoteIdentity(
				us.ID, provider, remoteID, externalUser.AccessToken,
			); err != nil {
				return err
			}

			// Enable provider in session.
			us.enableProvider(provider)
			if err := sess.Save(r, w); err != nil {
				return err
			}
			return c.Redirect(http.StatusFound, redirect)
		}

		// Try to get user from database.
		var user *models.User
		user, err = db.GetUserByRemoteIdentity(provider, remoteID, externalUser.AccessToken)
		if err == gorm.ErrRecordNotFound {
			// Create new user.
			user, err = db.CreateUserFromRemoteIdentity(
				externalUser.FirstName, externalUser.LastName,
				externalUser.Email, externalUser.AvatarURL,
				provider, remoteID, externalUser.AccessToken,
			)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		// Register user session.
		us := newUserSession(user.ID)
		us.enableProvider(provider)
		sess.Values[UserKey] = us
		if err := sess.Save(r, w); err != nil {
			return err
		}

		return c.Redirect(http.StatusFound, redirect)
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

			// If type assertion fails, the recover middleware will catch the panic and log a stack trace.
			us := i.(*UserSession)
			user, err := db.GetUser(us.ID)
			if err != nil {
				// Invalidate session. This could happen if the user has been entirely remove
				// from the database, but a valid session still exists.
				if err == gorm.ErrRecordNotFound {
					OAuth2Logout()
				}
				return echo.ErrUnauthorized
			}

			// TODO: Check if the user is allowed to access the endpoint.
			c.Set(UserKey, user)
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

func extractRedirectURL(r *http.Request, key string) string {
	// TODO: Validate redirect URL.
	url := r.URL.Query().Get(key)
	if url == "" {
		url = "/"
	}
	return url
}

func extractState(r *http.Request, key string) (redirect string, teacher bool) {
	// TODO: Validate redirect URL.
	url := r.URL.Query().Get(key)
	teacher = url != "" && url[:1] == "1"

	if url == "" || url[1:] == "" {
		return "/", teacher
	}
	return url[1:], teacher
}
