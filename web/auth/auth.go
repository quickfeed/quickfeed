package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/web/auth/tokens"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

const (
	UserKey        = "user"
	TeacherSuffix  = "teacher"
	Cookie         = "cookie"
	OutgoingCookie = "Set-Cookie"
	githubUserAPI  = "https://api.github.com/user"
)

var (
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
}

func authenticationError(logger *zap.SugaredLogger, w http.ResponseWriter, err error) {
	logger.Error(err)
	w.WriteHeader(http.StatusUnauthorized)
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
		newCookie := &http.Cookie{
			Name:     "auth",
			Value:    "",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   true,
		}
		http.SetCookie(w, newCookie)
		http.Redirect(w, r, "/", http.StatusFound)
	}
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
func OAuth2Callback(logger *zap.SugaredLogger, db database.Database, tm *tokens.TokenManager, authConfig *oauth2.Config, scms *Scms, secret string) http.HandlerFunc {
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

		cookie, err := tm.NewAuthCookie(user.ID)
		if err != nil {
			authenticationError(logger, w, fmt.Errorf("failed to create authentication cookie for user %q: %w", externalUser.Login, err))
			return
		}
		if ok := updateScm(logger, scms, user); !ok {
			authenticationError(logger, w, fmt.Errorf("failed to update SCM for user %q: %w", externalUser.Login, err))
			return
		}
		http.SetCookie(w, cookie)
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

var (
	ErrInvalidAuthCookie = status.Errorf(codes.Unauthenticated, "request does not contain a valid authentication cookie.")
	ErrContextMetadata   = status.Errorf(codes.Unauthenticated, "could not obtain metadata from context")
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
	return nil, ErrInvalidAuthCookie
}
