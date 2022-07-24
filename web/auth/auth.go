package auth

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/rand"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
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
	TeacherSuffix  = "teacher"
	Cookie         = "cookie"
	OutgoingCookie = "Set-Cookie"
	githubUserAPI  = "https://api.github.com/user"
)

// Query keys.
const (
	State    = "state" // As defined by the OAuth2 RFC.
	Redirect = "redirect"
)

// Temporary solution. Will be removed when sessions replaced by JWT.
var (
	// Gothic http cookie sessionStore
	sessionStore  *sessions.CookieStore
	teacherScopes = []string{"repo:invite", "user", "repo", "delete_repo", "admin:org", "admin:org_hook"}
	studentScopes = []string{"repo:invite"}
	// map from session cookies to user IDs.
	cookieStore = make(map[string]uint64)
	httpClient  *http.Client
)

func init() {
	httpClient = &http.Client{
		Timeout: time.Second * 30,
	}
	sessionStore = sessions.NewCookieStore([]byte(rand.String()))
	sessionStore.Options.HttpOnly = true
	sessionStore.Options.Secure = true
}

// UserSession holds user session information.
type UserSession struct {
	ID        uint64
	Providers map[string]struct{}
}

func authenticationError(logger *zap.SugaredLogger, w http.ResponseWriter, err error) {
	logger.Error(err)
	w.WriteHeader(http.StatusUnauthorized)
}

func newUserSession(id uint64) *UserSession {
	return &UserSession{
		ID:        id,
		Providers: make(map[string]struct{}),
	}
}

func (us UserSession) String() string {
	providers := ""
	for provider := range us.Providers {
		providers += provider + " "
	}
	return fmt.Sprintf("UserSession{ID: %d, Providers: %v}", us.ID, providers)
}

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
func OAuth2Logout(logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := sessionStore.Get(r, SessionKey)
		if err != nil {
			logger.Error(err)
		}
		logger.Debug(sessionData(sess))
		sess.Options.MaxAge = -1
		sess.Values = make(map[interface{}]interface{})
		if err := sess.Save(r, w); err != nil {
			logger.Error(err)
		}
		http.Redirect(w, r, "/", http.StatusFound)
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

// OAuth2Login redirects user to the provider's sign in page or, if user is already signed in with provider,
// authenticates the user in the background.
func OAuth2Login(logger *zap.SugaredLogger, authConfig *oauth2.Config, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("OAuth2Login: started")
		if r.Method != "GET" {
			authenticationError(logger, w, fmt.Errorf("illegal request method: %s", r.Method))
			return
		}
		// Start or refresh the session.
		// Issue: server generates a new random key to encode sessions on each restart.
		// A session from before the restart will be detected, but cannot be decoded with
		// the new key, resulting in an error. Instead of returning, we attempt to refresh the session first.
		s, err := sessionStore.New(r, SessionKey)
		if err != nil {
			if err := sessionStore.Save(r, w, s); err != nil {
				authenticationError(logger, w, fmt.Errorf("failed to create new session: %w", err))
				return
			}
		}
		// Check teacher suffix, update scopes.
		// Won't be necessary with GitHub App.
		setScopes(authConfig, r.URL.Path)
		logger.Debugf("Provider callback URL: %s", authConfig.RedirectURL)
		redirectURL := authConfig.AuthCodeURL(secret)
		logger.Debugf("Redirecting to AuthURL: %v", redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

// OAuth2Callback handles the callback from an oauth2 provider.
func OAuth2Callback(logger *zap.SugaredLogger, db database.Database, authConfig *oauth2.Config, scms *Scms, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("OAuth2Callback: started")
		if r.Method != "GET" {
			authenticationError(logger, w, fmt.Errorf("illegal request method: %s", r.Method))
			return
		}
		token, err := extractAccessToken(r, authConfig, secret)
		if err != nil {
			authenticationError(logger, w, err)
			return
		}
		externalUser, err := fetchExternalUser(token)
		if err != nil {
			authenticationError(logger, w, err)
			return
		}
		logger.Debugf("ExternalUser: %v", qlog.IndentJson(externalUser))

		// There is no need to check for existing session here because a user with a valid session
		// will be logged in automatically and will never see the login button. Only a user without
		// a valid session can be redirected to this endpoint. (Same will hold for JWT-based authentication).
		remote := &qf.RemoteIdentity{
			Provider:    env.ScmProvider(),
			RemoteID:    externalUser.ID,
			AccessToken: token.AccessToken,
		}
		// in case this is a new user we need a user object with full information,
		// otherwise frontend will get user object where only name, email and url are set.
		user, err := fetchUser(logger, db, remote, externalUser)
		if err != nil {
			authenticationError(logger, w, fmt.Errorf("failed to fetch user %q for remote identity: %w", externalUser.Login, err))
			return
		}
		logger.Debugf("Fetching full user info for %v, user: %v", remote, user)

		// Register user session.
		// Temporary. Will be removed when sessions are replaced with JWT.
		s, err := sessionStore.Get(r, SessionKey)
		if err != nil {
			authenticationError(logger, w, fmt.Errorf("failed to get user session for %q: %w", externalUser.Login, err))
			return
		}
		us := newUserSession(user.ID)
		s.Values[UserKey] = us
		logger.Debugf("New Session: %s", us)
		// save the session to a store in addition to adding an outgoing ('set-cookie') session cookie to the response.
		if err := sessionStore.Save(r, w, s); err != nil {
			authenticationError(logger, w, fmt.Errorf("failed to save session: %w", err))
			return
		}
		logger.Debugf("Session.Save: %v", s)

		if ok := updateScm(logger, scms, user); !ok {
			logger.Debugf("Failed to update SCM for User: %v", user)
		}

		// Register session and associated user ID to enable gRPC requests for this session.
		if token := extractSessionCookie(w); len(token) > 0 {
			logger.Debugf("SessionCookie: %v", token)
			Add(token, us.ID)
		} else {
			authenticationError(logger, w, fmt.Errorf("no session cookie found in %v", w.Header()))
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// extractAccessToken exchanges code received from OAuth provider for the user's access token.
func extractAccessToken(r *http.Request, authConfig *oauth2.Config, secret string) (*oauth2.Token, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	callbackSecret := r.FormValue("state")
	if callbackSecret != secret {
		return nil, errors.New("incorrect callback secret")
	}
	code := r.FormValue("code")
	if code == "" {
		return nil, errors.New("got empty code on callback")
	}
	token, err := authConfig.Exchange(r.Context(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange access token: %w", err)
	}
	return token, nil
}

// fetchExternalUser fetches information about the user from the provider.
func fetchExternalUser(token *oauth2.Token) (*externalUser, error) {
	req, err := http.NewRequest("GET", githubUserAPI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user request: %w", err)
	}
	token.SetAuthHeader(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send user request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected OAuth provider response: %d (%s)", resp.StatusCode, resp.Status)
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read authentication response: %w", err)
	}
	externalUser := &externalUser{}
	if err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&externalUser); err != nil {
		return nil, fmt.Errorf("failed to decode user information: %w", err)
	}
	return externalUser, nil
}

// fetchUser saves or updates user information fetched from the OAuth provider in the database.
func fetchUser(logger *zap.SugaredLogger, db database.Database, remote *qf.RemoteIdentity, externalUser *externalUser) (*qf.User, error) {
	logger.Debugf("Lookup user: %q in database with: %v", externalUser.Login, remote)
	user, err := db.GetUserByRemoteIdentity(remote)
	switch {
	case err == nil:
		logger.Debugf("Found user: %v in database", user)
		if err = db.UpdateAccessToken(remote); err != nil {
			return nil, fmt.Errorf("failed to update access token for user %q: %w", externalUser.Login, err)
		}
		logger.Debugf("Access token updated: %v", remote)

	case err == gorm.ErrRecordNotFound:
		logger.Debugf("User %q not found in database; creating new user", externalUser.Login)
		user = &qf.User{
			Name:      externalUser.Name,
			Email:     externalUser.Email,
			AvatarURL: externalUser.AvatarURL,
			Login:     externalUser.Login,
		}
		if err = db.CreateUserFromRemoteIdentity(user, remote); err != nil {
			return nil, fmt.Errorf("failed to create remote identity for user %q: %w", externalUser.Login, err)
		}
		logger.Debugf("New user created: %v, remote: %v", user, remote)

	default:
		return nil, fmt.Errorf("failed to fetch user %q for remote identity: %w", externalUser.Login, err)
	}
	logger.Debugf("Retry database lookup for user %q", externalUser.Login)
	return db.GetUserByRemoteIdentity(remote)
}

// setScopes sets student or teacher scopes for user authentication.
func setScopes(authConfig *oauth2.Config, url string) {
	if strings.Contains(url, TeacherSuffix) {
		authConfig.Scopes = teacherScopes
	} else {
		authConfig.Scopes = studentScopes
	}
}

func updateScm(logger *zap.SugaredLogger, scms *Scms, user *qf.User) bool {
	foundSCMProvider := false
	for _, remoteID := range user.RemoteIdentities {
		if _, err := scms.GetOrCreateSCMEntry(logger.Desugar(), remoteID.GetAccessToken()); err != nil {
			logger.Errorf("Unknown SCM provider: %v", err)
			continue
		}
		foundSCMProvider = true
	}
	if !foundSCMProvider {
		logger.Debugf("No SCM provider found for user %v", user)
	}
	return foundSCMProvider
}

func extractSessionCookie(w http.ResponseWriter) string {
	// Helper function that extracts an outgoing session cookie.
	outgoingCookies := w.Header()[OutgoingCookie]
	for _, cookie := range outgoingCookies {
		if c := strings.Split(cookie, "="); c[0] == SessionKey {
			token := strings.Split(cookie, ";")[0]
			return token
		}
	}
	return ""
}

var (
	ErrInvalidSessionCookie = status.Errorf(codes.Unauthenticated, "request does not contain a valid session cookie.")
	ErrContextMetadata      = status.Errorf(codes.Unauthenticated, "could not obtain metadata from context")
)

func UserVerifier() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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
		return handler(newCtx, req)
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
