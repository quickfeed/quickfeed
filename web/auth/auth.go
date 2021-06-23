package auth

import (
	"encoding/gob"
	"net/http"
	"strconv"
	"strings"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"
	"go.uber.org/zap"
)

func init() {
	gob.Register(&UserSession{})
	TokenStore = &UserToken{store: make(map[string]uint64)}
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

// TokenToUserMap
type UserToken struct {
	store map[string]uint64
}

func (ut *UserToken) add(token string, id uint64) {
	for token, userID := range TokenStore.store {
		if userID == id {
			delete(TokenStore.store, token)
		}
	}
	ut.store[token] = id
}

func (ut *UserToken) Get(token string) uint64 {
	return TokenStore.store[token]
}

var TokenStore *UserToken

func (us *UserSession) enableProvider(provider string) {
	us.Providers[provider] = struct{}{}
}

// OAuth2Logout invalidates the session for the logged in user.
func OAuth2Logout(logger *zap.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		r := c.Request()
		w := c.Response()

		sess, err := session.Get(SessionKey, c)
		if err != nil {
			logger.Error(err.Error())
			return sess.Save(r, w)
		}

		if i, ok := sess.Values[UserKey]; ok {
			// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
			us := i.(*UserSession)
			// Invalidate gothic user sessions.
			for provider := range us.Providers {
				sess, err := session.Get(provider+GothicSessionKey, c)
				if err != nil {
					logger.Error(err.Error())
					return err
				}
				sess.Options.MaxAge = -1
				sess.Values = make(map[interface{}]interface{})
				if err := sess.Save(r, w); err != nil {
					logger.Error(err.Error())
				}
			}
		}
		// Invalidate our user session.
		sess.Options.MaxAge = -1
		sess.Values = make(map[interface{}]interface{})
		if err := sess.Save(r, w); err != nil {
			logger.Error(err.Error())
		}

		return c.Redirect(http.StatusFound, extractRedirectURL(r, Redirect))
	}
}

// PreAuth checks the current user session and executes the next handler if none
// was found for the given provider.
func PreAuth(logger *zap.Logger, db database.Database) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, err := session.Get(SessionKey, c)
			if err != nil {
				logger.Error(err.Error())
				if err := sess.Save(c.Request(), c.Response()); err != nil {
					logger.Error(err.Error())
					return err
				}
				return next(c)
			}

			if i, ok := sess.Values[UserKey]; ok {
				// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
				us := i.(*UserSession)
				if _, err := db.GetUser(us.ID); err != nil {
					logger.Error(err.Error())
					return OAuth2Logout(logger)(c)
				}
			}
			return next(c)
		}
	}
}

// OAuth2Login tries to authenticate against an oauth2 provider.
func OAuth2Login(logger *zap.Logger, db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		w := c.Response()
		r := c.Request()

		provider, err := gothic.GetProviderName(r)
		if err != nil {
			logger.Error(err.Error())
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
			logger.Error(err.Error())
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		// Redirect to provider to perform authentication.
		return c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

// OAuth2Callback handles the callback from an oauth2 provider.
func OAuth2Callback(logger *zap.Logger, db database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		logger.Debug("OAuth2Callback: started")
		w := c.Response()
		r := c.Request()

		qv := r.URL.Query()
		redirect, teacher := extractState(r, State)
		provider, err := gothic.GetProviderName(r)
		if err != nil {
			logger.Error("failed to get gothic provider", zap.Error(err))
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
			logger.Error("failed to complete user authentication", zap.Error(err))
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		remoteID, err := strconv.ParseUint(externalUser.UserID, 10, 64)
		if err != nil {
			logger.Error(err.Error())
			return err
		}

		sess, err := session.Get(SessionKey, c)
		if err != nil {
			logger.Error(err.Error())
			return err
		}

		// Try to get already logged in user.
		if sess.Values[UserKey] != nil {
			i, ok := sess.Values[UserKey]
			if !ok {
				logger.Debug("failed to get logged in user from session; logout")
				return OAuth2Logout(logger)(c)
			}

			// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
			us := i.(*UserSession)
			// Associate user with remote identity.
			if err := db.AssociateUserWithRemoteIdentity(
				us.ID, provider, remoteID, externalUser.AccessToken,
			); err != nil {
				logger.Error("failed to associate user with remote identity", zap.Error(err))
				return err
			}

			// Enable provider in session.
			us.enableProvider(provider)
			if err := sess.Save(r, w); err != nil {
				logger.Error(err.Error())
				return err
			}
			return c.Redirect(http.StatusFound, redirect)
		}

		remote := &pb.RemoteIdentity{
			Provider:    provider,
			RemoteID:    remoteID,
			AccessToken: externalUser.AccessToken,
		}
		// Try to get user from database.
		user, err := db.GetUserByRemoteIdentity(remote)
		switch {
		case err == nil:
			// found user in database; update access token
			err = db.UpdateAccessToken(remote)
			if err != nil {
				logger.Error("failed to update access token for user", zap.Error(err), zap.String("user", user.String()))
				return err
			}
		case err == gorm.ErrRecordNotFound:
			// user not in database; create new user
			user = &pb.User{
				Name:      externalUser.Name,
				Email:     externalUser.Email,
				AvatarURL: externalUser.AvatarURL,
				Login:     externalUser.NickName,
			}
			err = db.CreateUserFromRemoteIdentity(user, remote)
			if err != nil {
				logger.Error("failed to create remote identify for user", zap.Error(err), zap.String("user", user.String()))
				return err
			}
		default:
			logger.Error("failed to fetch user for remote identity", zap.Error(err))
		}

		// in case this is a new user we need a user object with full information,
		// otherwise frontend will get user object where only name, email and url are set.
		user, err = db.GetUserByRemoteIdentity(remote)
		if err != nil {
			logger.Error(err.Error())
			return err
		}

		// Register user session.
		us := newUserSession(user.ID)
		us.enableProvider(provider)
		sess.Values[UserKey] = us
		if err := sess.Save(r, w); err != nil {
			logger.Error(err.Error())
			return err
		}
		return c.Redirect(http.StatusFound, redirect)
	}
}

// AccessControl returns an access control middleware. Given a valid context
// with sufficient access the next handler is called. Missing or invalid
// credentials results in a 401 unauthorized response.
func AccessControl(logger *zap.Logger, db database.Database, scms *Scms) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, err := session.Get(SessionKey, c)
			if err != nil {
				logger.Error(err.Error())
				// Save fixes the session if it has been modified
				// or it is no longer valid due to newUserSess change of keys.
				if err := sess.Save(c.Request(), c.Response()); err != nil {
					logger.Error(err.Error())
					return err
				}
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}

			i, ok := sess.Values[UserKey]
			if !ok {
				logger.Error(echo.ErrUnauthorized.Error())
				return echo.ErrUnauthorized
			}

			// If type assertion fails, the recover middleware will catch the panic and log a stack trace.
			us := i.(*UserSession)
			user, err := db.GetUser(us.ID)
			if err != nil {
				logger.Error(err.Error())
				// Invalidate session. This could happen if the user has been entirely remove
				// from the database, but a valid session still exists.
				if err == gorm.ErrRecordNotFound {
					logger.Error(err.Error())
					return OAuth2Logout(logger)(c)
				}
				logger.Error(echo.ErrUnauthorized.Error())
				return echo.ErrUnauthorized
			}
			c.Set(UserKey, user)

			foundSCMProvider := false
			for _, remoteID := range user.RemoteIdentities {
				scm, err := scms.GetOrCreateSCMEntry(logger, remoteID.GetProvider(), remoteID.GetAccessToken())
				if err != nil {
					logger.Error("unknown SCM provider", zap.Error(err))
					continue
				}
				foundSCMProvider = true
				c.Set(remoteID.Provider, scm)
			}
			if !foundSCMProvider {
				logger.Info("no SCM providers found for", zap.String("user", user.String()))
				return echo.NewHTTPError(http.StatusBadRequest, err)
			}

			token, err := c.Cookie(SessionKey)
			if err != nil {
				return err
			}
			if id := TokenStore.Get(token.String()); id != user.GetID() {
				TokenStore.add(token.String(), user.GetID())
			}

			// TODO: Add access control list.
			// - Extract endpoint.
			// - Verify whether the user has sufficient rights. This
			//   can be a simple hash map. A user should be able to
			//   access /users/:uid if the user's id is uid.
			//   - Not authorized: return c.NoContent(http.StatusUnauthorized)
			//   - Authorized: return next(c)
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
