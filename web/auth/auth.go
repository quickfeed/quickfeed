package auth

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	lg "github.com/autograde/quickfeed/log"
	"github.com/autograde/quickfeed/scm"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func init() {
	gob.Register(&UserSession{})
}

// Session keys.
const (
	SessionKey     = "session"
	UserKey        = "user"
	Cookie         = "cookie"
	OutgoingCookie = "Set-Cookie"
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

func (us UserSession) String() string {
	providers := ""
	for provider := range us.Providers {
		providers += provider + " "
	}
	return fmt.Sprintf("UserSession{ID: %d, Providers: %v}", us.ID, providers)
}

// map from session cookies to user IDs.
var cookieStore = make(map[string]uint64)

// Add adds cookie for userID, replacing userID's current cookie, if any.
func Add(cookie string, userID uint64) {
	for currentCookie, id := range cookieStore {
		if id == userID && currentCookie != cookie {
			delete(cookieStore, currentCookie)
		}
	}
	cookieStore[cookie] = userID
}

func Get(cookie string) uint64 {
	return cookieStore[cookie]
}

// OAuth2Logout invalidates the session for the logged in user.
func OAuth2Logout(logger *zap.SugaredLogger) echo.HandlerFunc {
	return func(c echo.Context) error {
		r := c.Request()
		w := c.Response()

		sess, err := session.Get(SessionKey, c)
		if err != nil {
			logger.Error(err.Error())
			return sess.Save(r, w)
		}
		logger.Debug(sessionData(sess))

		if i, ok := sess.Values[UserKey]; ok {
			// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
			us := i.(*UserSession)
			logger.Debug(us)
			// Invalidate gothic user sessions.
			for provider := range us.Providers {
				sess, err := session.Get(provider+gothic.SessionName, c)
				if err != nil {
					logger.Error(err.Error())
					return err
				}
				logger.Debug(sessionData(sess))

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
func PreAuth(logger *zap.SugaredLogger, db database.Database) echo.MiddlewareFunc {
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
			logger.Debug(sessionData(sess))

			if i, ok := sess.Values[UserKey]; ok {
				// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
				us := i.(*UserSession)
				logger.Debug(us)
				user, err := db.GetUser(us.ID)
				if err != nil {
					logger.Error(err.Error())
					return OAuth2Logout(logger)(c)
				}
				logger.Debugf("User: %v", user)
			}
			return next(c)
		}
	}
}

func sessionData(session *sessions.Session) string {
	if session == nil {
		return "<nil>"
	}
	out := "Values: "
	for k, v := range session.Values {
		out += fmt.Sprintf("<%s: %v>, ", k, v)
	}
	out += "Options: "
	out += fmt.Sprintf("<%s: %v>, ", "MaxAge", session.Options.MaxAge)
	out += fmt.Sprintf("<%s: %v>, ", "Path", session.Options.Path)
	out += fmt.Sprintf("<%s: %v>, ", "Domain", session.Options.Domain)
	out += fmt.Sprintf("<%s: %v>, ", "Secure", session.Options.Secure)
	out += fmt.Sprintf("<%s: %v>, ", "HttpOnly", session.Options.HttpOnly)
	out += fmt.Sprintf("<%s: %v>, ", "SameSite", session.Options.SameSite)

	return fmt.Sprintf("Session: ID=%s, IsNew=%t, %s", session.ID, session.IsNew, out)
}

// OAuth2Login tries to authenticate against an oauth2 provider.
func OAuth2Login(logger *zap.SugaredLogger, db database.Database, config oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			logger.Errorf("GitHub login failed: request method %s", r.Method)
			http.Redirect(w, r, "/", http.StatusUnauthorized)
		}
		// TODO(vera): adapt to use with other providers if needed
		provider := "github"

		// TODO(vera): make sure teacher suffix no longer necessary
		var teacher int
		if strings.HasSuffix(provider, TeacherSuffix) {
			teacher = 1
		}
		logger.Debugf("Provider: %v ; Teacher: %v", provider, teacher)
		qv := r.URL.Query()
		logger.Debugf("qv: %v", qv)
		// redirect := extractRedirectURL(r, Redirect)
		logger.Debugf("redirect: %v", config.RedirectURL)
		// TODO: Add a random string to protect against CSRF.
		qv.Set(State, strconv.Itoa(teacher)+config.RedirectURL)
		logger.Debugf("State: %v", strconv.Itoa(teacher)+config.RedirectURL)
		r.URL.RawQuery = qv.Encode()
		logger.Debugf("RawQuery: %v", r.URL.RawQuery)
		logger.Debugf("Redirecting to %s to perform authentication; AuthURL: %v", provider, config.Endpoint.AuthURL)
		http.Redirect(w, r, config.Endpoint.AuthURL, http.StatusTemporaryRedirect)
	}
}

// OAuth2Callback handles the callback from an oauth2 provider.
func OAuth2Callback(logger *zap.SugaredLogger, db database.Database, config oauth2.Config, app *scm.GithubApp, tokens *TokenManager, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			logger.Errorf("GitHub login failed: request method %s", r.Method)
			http.Redirect(w, r, "/", http.StatusUnauthorized)
		}
		logger.Debug("OAuth2Callback: started")
		qv := r.URL.Query()
		logger.Debugf("qv: %v", qv)
		redirect, teacher := extractState(r, State)
		logger.Debugf("Redirect: %v ; Teacher: %t", redirect, teacher)

		provider := "github"
		// Add teacher suffix if upgrading scope.
		if teacher {
			qv.Set("provider", provider+TeacherSuffix)
			logger.Debugf("Set('provider') = %v", provider+TeacherSuffix)
		}
		r.URL.RawQuery = qv.Encode()
		logger.Debugf("RawQuery: %v", r.URL.RawQuery)

		// Complete authentication.
		// parse request for code and state
		if err := r.ParseForm(); err != nil {
			logger.Error("GitHub login failed: error parsing authentication code")
			http.Redirect(w, r, "/", http.StatusUnauthorized)
		}

		logger.Debug("VALIDATING STATE") // tmp
		// validate state
		callbackSecret := r.FormValue("state")
		logger.Debug("Callback: got state in request: ", callbackSecret) // tmp
		if callbackSecret != secret {
			logger.Errorf("Warning: secrets don't match: expected %s, got %s", secret, callbackSecret)
			http.Redirect(w, r, "/", http.StatusUnauthorized)
		}

		logger.Debug("EXCHANGING CODE FOR TOKEN") // tmp
		// exchange code for token
		code := r.FormValue("code")
		if code == "" {
			logger.Error("GitHub login failed: received empty code")
			http.Redirect(w, r, "/", http.StatusUnauthorized)
		}
		logger.Debug("CODE RECEIVED, PROCEED") // tmp
		githubToken, err := config.Exchange(context.Background(), code)
		if err != nil {
			logger.Errorf("GitHub login failed: cannot exchange token: %s", err)
			http.Redirect(w, r, "/", http.StatusUnauthorized)
		}
		logger.Debugf("Successfully fetched access token: %s", githubToken.AccessToken) // tmp

		// get user info with the token
		req, err := http.NewRequest("GET", app.GetUserURL(), nil)
		if err != nil {
			logger.Errorf("GitHub login failed: failed to make user request: %s", err)
			http.Redirect(w, r, "/", http.StatusUnauthorized)
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", githubToken.AccessToken))
		resp, err := app.App.Client().Do(req)
		if err != nil {
			logger.Errorf("GitHub login failed: failed to send user request: %s", err)
			http.Redirect(w, r, "/", http.StatusUnauthorized)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Errorf("GitHub login failed: API responded with status: %d: %s", resp.StatusCode, resp.Status)
			http.Redirect(w, r, "/", http.StatusUnauthorized)
		}
		respBits, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error reading response bits from user API: ", err.Error())
		}
		externalUser := &externalUser{}
		if err := json.NewDecoder(bytes.NewReader(respBits)).Decode(&externalUser); err != nil {
			logger.Errorf("GitHub login failed: failed to decode user information: %s", err)
			http.Redirect(w, r, "/", http.StatusUnauthorized)
		}
		logger.Debugf("externalUser: %v", lg.IndentJson(externalUser))
		logger.Debugf("EXTRACTED set-cookie token: %s", extractToken(w)) // tmp

		var userToken string
		for _, cookie := range r.Cookies() {
			if cookie.Name == "auth" {
				userToken = cookie.Value
			}
		}
		logger.Debugf("EXTRACTED auth cookie", userToken) // tmp

		// There is already a cookie with JWT, make sure the user exists in the database
		if userToken != "" {
			// TODO(vera): here is a good place to verify that the user claims and the current user that
			// is set in the frontend aren't different
			claims, err := tokens.GetClaims(userToken)
			if err != nil {
				logger.Errorf("GitHub login failed: failed to read user claims: %s", err)
				http.Redirect(w, r, "/", http.StatusUnauthorized)
			}

			if err := db.AssociateUserWithRemoteIdentity(claims.UserID, provider, externalUser.ID, githubToken.AccessToken); err != nil {
				logger.Debugf("Associate failed: %d, %s, %d, %s", claims.UserID, provider, externalUser.ID, githubToken.AccessToken)
				logger.Errorf("GitHub login failed: failed to associate user with remote identity: %s", err)
				http.Redirect(w, r, "/", http.StatusUnauthorized)
			}
			logger.Debugf("Associate: %d, %s, %d, %s", claims.UserID, provider, externalUser.ID, githubToken.AccessToken)
		}

		// If no user cookie in context
		remote := &pb.RemoteIdentity{
			Provider:    provider,
			RemoteID:    externalUser.ID,
			AccessToken: githubToken.AccessToken,
		}
		// Try to get user from database.
		user, err := db.GetUserByRemoteIdentity(remote)
		switch {
		case err == nil:
			logger.Debugf("found user: %v", user)
			// found user in database; update access token
			err = db.UpdateAccessToken(remote)
			if err != nil {
				logger.Errorf("GitHub login failed: failed to update access token for user %v: %s", externalUser, err)
				http.Redirect(w, r, "/", http.StatusUnauthorized)
			}
			logger.Debugf("access token updated: %v", remote)

		case err == gorm.ErrRecordNotFound:
			logger.Debug("user not found in database; creating new user")
			// user not in database; create new user
			user = &pb.User{
				Name:      externalUser.Name,
				Email:     externalUser.Email,
				AvatarURL: externalUser.AvatarURL,
				Login:     externalUser.Login,
			}
			err = db.CreateUserFromRemoteIdentity(user, remote)
			if err != nil {
				logger.Errorf("GitHub login failed: failed to create remote identity for user %v: %s", externalUser, err)
				http.Redirect(w, r, "/", http.StatusUnauthorized)
			}
			logger.Debugf("New user created: %v, remote: %v", user, remote)

		default:
			logger.Error("failed to fetch user for remote identity", zap.Error(err))
		}

		// in case this is a new user we need a user object with full information,
		// otherwise frontend will get user object where only name, email and url are set.
		user, err = db.GetUserByRemoteIdentity(remote)
		if err != nil {
			logger.Errorf("GitHub login failed: failed to fetch user %v	 from database: %s", externalUser, err)
			http.Redirect(w, r, "/", http.StatusUnauthorized)
		}
		logger.Debugf("Fetching full user info for %v, user: %v", remote, user)

		claims, err := tokens.NewClaims(user.ID)
		if err != nil {
			logger.Errorf("GitHub login failed: failed to make claims for user %v: %s", externalUser, err)
			http.Redirect(w, r, "/", http.StatusUnauthorized)
		}
		cookie, err := tokens.NewTokenCookie(context.Background(), tokens.NewToken(claims))
		if err != nil {
			logger.Errorf("GitHub login failed: failed to make token cookie for user %v: %s", externalUser, err)
			// TODO(vera): this pattern for handling auth errors might be better
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

// AccessControl returns an access control middleware. Given a valid context
// with sufficient access the next handler is called. Missing or invalid
// credentials results in a 401 unauthorized response.
func AccessControl(logger *zap.SugaredLogger, db database.Database) echo.MiddlewareFunc {
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
				return next(c)
			}
			logger.Debug(sessionData(sess))

			i, ok := sess.Values[UserKey]
			if !ok {
				return next(c)
			}

			// If type assertion fails, the recover middleware will catch the panic and log a stack trace.
			us := i.(*UserSession)
			logger.Debug(us)
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
				return next(c)
			}
			c.Set(UserKey, user)

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

func extractToken(w http.ResponseWriter) string {
	// Helper function that extracts an outgoing session cookie.
	outgoingCookies := w.Header().Values(OutgoingCookie)
	for _, cookie := range outgoingCookies {
		if c := strings.Split(cookie, "="); c[0] == "auth" {
			token := strings.Split(cookie, ";")[0]
			return token
		}
	}
	return ""
}

var (
	ErrInvalidSessionCookie = status.Errorf(codes.Unauthenticated, "Request does not contain a valid session cookie.")
	ErrContextMetadata      = status.Errorf(codes.Unauthenticated, "Could not obtain metadata from context")
)

func UserVerifier() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, ErrContextMetadata
		}
		newMeta, err := userValidation(meta)
		if err != nil {
			return nil, err
		}
		// create new context with user id instead of cookie for use internally
		newCtx := metadata.NewIncomingContext(ctx, newMeta)
		resp, err := handler(newCtx, req)
		return resp, err
	}
}

// userValidation returns modified metadata containing a valid user.
// An error is returned if the user is not authenticated.
func userValidation(meta metadata.MD) (metadata.MD, error) {
	for _, cookie := range meta.Get(Cookie) {
		if user := Get(cookie); user > 0 {
			meta.Set(UserKey, strconv.FormatUint(user, 10))
			return meta, nil
		}
	}
	return nil, ErrInvalidSessionCookie
}
