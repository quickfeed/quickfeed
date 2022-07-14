package auth

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/rand"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
	"go.uber.org/zap"
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

func authenticationError(logger *zap.SugaredLogger, w http.ResponseWriter, err string) {
	logger.Error(err)
	w.WriteHeader(http.StatusUnauthorized)
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
			logger.Error(err.Error())
		}
		logger.Debug(sessionData(sess))
		sess.Options.MaxAge = -1
		sess.Values = make(map[interface{}]interface{})
		if err := sess.Save(r, w); err != nil {
			logger.Error(err.Error())
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
func OAuth2Login(logger *zap.SugaredLogger, db database.Database, authConfig *AuthConfig, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("AUTH2LOGIN started with: ", r.Method) // tmp
		if r.Method != "GET" {
			authenticationError(logger, w, fmt.Sprintf("illegal request method: %s", r.Method))
			return
		}

		// Start or refresh the session.
		// Issue: server generates a new random key to encode sessions on each restart.
		// A session from before the restart will be detected, but cannot be decoded with
		// the new key, resulting in an error. Instead of returning, we attempt to refresh the session first.
		s, err := sessionStore.Get(r, SessionKey)
		if err != nil {
			if err := sessionStore.Save(r, w, s); err != nil {
				authenticationError(logger, w, fmt.Sprintf("failed to create new session (%s): %s", SessionKey, err))
				return
			}
		}

		// Get provider name.
		provider := GetProviderName(r.URL.Path, 2)
		if provider == "" {
			authenticationError(logger, w, fmt.Sprintf("incorrect request URL: %s", r.URL.Path))
			return
		}
		providerConfig, ok := authConfig.providers[provider]
		if !ok {
			authenticationError(logger, w, fmt.Sprintf("provider %s is not enabled", provider))
			return
		}

		// Check teacher suffix, update scopes if teacher.
		// Won't be necessary with GitHub App.
		if strings.Contains(r.URL.Path, TeacherSuffix) {
			logger.Debug("found teacher suffix on login")
			providerConfig.Scopes = teacherScopes
		} else {
			providerConfig.Scopes = studentScopes
		}
		logger.Debugf("Provider callback URL: %s", providerConfig.RedirectURL)
		redirectURL := providerConfig.AuthCodeURL(secret)
		logger.Debugf("redirecting to %s to perform authentication; AuthURL: %v", provider, redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

// OAuth2Callback handles the callback from an oauth2 provider.
func OAuth2Callback(logger *zap.SugaredLogger, db database.Database, authConfig *AuthConfig, scms *Scms, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("OAuth2Callback: started")

		if r.Method != "GET" {
			authenticationError(logger, w, fmt.Sprintf("illegal request method: %s", r.Method))
			return
		}

		// Get provider.
		provider := GetProviderName(r.URL.Path, 3)
		if provider == "" {
			authenticationError(logger, w, fmt.Sprintf("incorrect request URL: %s", r.URL.Path))
			return
		}

		providerConfig, ok := authConfig.providers[provider]
		if !ok {
			authenticationError(logger, w, fmt.Sprintf("provider %s is not enabled", provider))
			return
		}

		if err := r.ParseForm(); err != nil {
			authenticationError(logger, w, fmt.Sprintf("failed to parse request form: %s", err))
			return
		}

		// Validate redirect state to make sure the request is in fact coming from the oauth provider.
		callbackSecret := r.FormValue("state")
		if callbackSecret != secret {
			authenticationError(logger, w, fmt.Sprintf("mismatching secrets: expected %s, got %s", secret, callbackSecret))
			return
		}

		// Exchange code for access token.
		code := r.FormValue("code")
		if code == "" {
			authenticationError(logger, w, fmt.Sprintf("got empty code on callback from %s", provider))
			return
		}

		authToken, err := providerConfig.Exchange(context.Background(), code)
		if err != nil {
			authenticationError(logger, w, fmt.Sprintf("failed to exchange access token: %s", err))
			return
		}

		// Use access token to fetch information about the GitHub user.
		req, err := http.NewRequest("GET", githubUserAPI, nil)
		if err != nil {
			authenticationError(logger, w, fmt.Sprintf("failed to create user request: %s", err))
			return
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken.AccessToken))

		resp, err := httpClient.Do(req)
		if err != nil {
			authenticationError(logger, w, fmt.Sprintf("failed to send user request: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			authenticationError(logger, w, fmt.Sprintf("unexpected OAuth provider response: %d (%s)", resp.StatusCode, resp.Status))
			return
		}
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			authenticationError(logger, w, fmt.Sprintf("failed to read response from %s: %v", provider, err))
			return
		}

		// Complete user authentication.
		externalUser := &externalUser{}
		if err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&externalUser); err != nil {
			authenticationError(logger, w, fmt.Sprintf("failed to decode user information: %v", err))
			return
		}
		logger.Debugf("externalUser: %v", qlog.IndentJson(externalUser))

		accessToken := authToken.AccessToken

		// There is no need to check for existing session here because a user with a valid session
		// will be logged in authomatically and will never see the login button. Only a user without
		// a valid session can be redirected to this endpoint. (Same will hold for JWT-based authentication).

		remote := &qf.RemoteIdentity{
			Provider:    provider,
			RemoteID:    externalUser.ID,
			AccessToken: accessToken,
		}

		// Try to get user from database.
		user, err := db.GetUserByRemoteIdentity(remote)
		switch {
		case err == nil:
			logger.Debugf("found user: %v", user)
			// found user in database; update access token
			err = db.UpdateAccessToken(remote)
			if err != nil {
				authenticationError(logger, w, fmt.Sprintf("Failed to update access token for user %s: %s", externalUser.Login, err))
				return
			}
			logger.Debugf("access token updated: %v", remote)

		case err == gorm.ErrRecordNotFound:
			logger.Debug("user not found in database; creating new user")
			// user not in database; create new user
			user = &qf.User{
				Name:      externalUser.Name,
				Email:     externalUser.Email,
				AvatarURL: externalUser.AvatarURL,
				Login:     externalUser.Login,
			}
			err = db.CreateUserFromRemoteIdentity(user, remote)
			if err != nil {
				authenticationError(logger, w, fmt.Sprintf("failed to create remote identity for user %s with %s: %s", externalUser.Login, provider, err))
				return
			}
			logger.Debugf("New user created: %v, remote: %v", user, remote)

		default:
			authenticationError(logger, w, fmt.Sprintf("failed to fetch user %s for remote identity from %s: %s", externalUser.Login, provider, err))
			return
		}

		// in case this is a new user we need a user object with full information,
		// otherwise frontend will get user object where only name, email and url are set.
		user, err = db.GetUserByRemoteIdentity(remote)
		if err != nil {
			authenticationError(logger, w, fmt.Sprintf("failed to fetch user %s for remote identity from %s: %s", externalUser.Login, provider, err))
			return
		}
		logger.Debugf("Fetching full user info for %v, user: %v", remote, user)

		// Register user session.
		// Temporary. Will be removed when sessions are replaced with JWT.
		s, err := sessionStore.Get(r, SessionKey)

		us := newUserSession(user.ID)
		us.enableProvider(provider)
		s.Values[UserKey] = us
		logger.Debugf("New Session: %s", us)
		// save the session to a store in addition to adding an outgoing ('set-cookie') session cookie to the response.
		if err := sessionStore.Save(r, w, s); err != nil {
			authenticationError(logger, w, fmt.Sprintf("failed to save session: %s", err))
			return
		}
		logger.Debugf("Session.Save: %v", s)

		if ok := updateScm(logger, scms, user); !ok {
			logger.Debugf("Failed to update SCM for User: %v", user)
		}

		// Register session and associated user ID to enable gRPC requests for this session.
		if token := extractSessionCookie(w); len(token) > 0 {
			logger.Debugf("extractSessionCookie: %v", token)
			Add(token, us.ID)
		} else {
			authenticationError(logger, w, fmt.Sprintf("no session cookie found in %v", w.Header()))
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
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
		// ctx.Set(remoteID.Provider, scm)
	}
	if !foundSCMProvider {
		logger.Debugf("No SCM provider found for user %v", user)
	}
	return foundSCMProvider
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
