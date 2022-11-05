package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/qf"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

var httpClient = &http.Client{
	Timeout: time.Second * 30,
}

func authenticationError(logger *zap.SugaredLogger, w http.ResponseWriter, err error) {
	logger.Error(err)
	w.WriteHeader(http.StatusUnauthorized)
}

// OAuth2Logout invalidates the session for the logged in user.
func OAuth2Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newCookie := &http.Cookie{
			Name:     CookieName,
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
		redirectURL := authConfig.AuthCodeURL(secret)
		logger.Debugf("Redirecting to AuthURL: %v", redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

// OAuth2Callback handles the callback from an oauth2 provider.
func OAuth2Callback(logger *zap.SugaredLogger, db database.Database, tm *TokenManager, authConfig *oauth2.Config, secret string) http.HandlerFunc {
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
		externalUser, err := FetchExternalUser(token)
		if err != nil {
			authenticationError(logger, w, err)
			return
		}
		logger.Debugf("ExternalUser: %v", qlog.IndentJson(externalUser))
		// in case this is a new user we need a user object with full information,
		// otherwise frontend will get user object where only name, email and url are set.
		user, err := fetchUser(logger, db, token, externalUser)
		if err != nil {
			authenticationError(logger, w, fmt.Errorf("failed to fetch user %q for remote identity: %w", externalUser.Login, err))
			return
		}
		logger.Debugf("Fetched full user info for user: %v", user)

		cookie, err := tm.NewAuthCookie(user.ID)
		if err != nil {
			authenticationError(logger, w, fmt.Errorf("failed to create authentication cookie for user %q: %w", externalUser.Login, err))
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
func FetchExternalUser(token *oauth2.Token) (*externalUser, error) {
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
func fetchUser(logger *zap.SugaredLogger, db database.Database, token *oauth2.Token, externalUser *externalUser) (*qf.User, error) {
	logger.Debugf("Lookup user: %q in database with SCM remote ID: %d", externalUser.Login, externalUser.ID)
	user, err := db.GetUserByRemoteIdentity(externalUser.ID)
	switch {
	case err == nil:
		logger.Debugf("Found user: %v in database", user)
		user.RefreshToken = token.RefreshToken
		if err = db.UpdateUser(user); err != nil {
			return nil, fmt.Errorf("failed to update access token for user %q: %w", externalUser.Login, err)
		}
		logger.Debugf("Refresh token updated: %v", token.RefreshToken)

	case err == gorm.ErrRecordNotFound:
		logger.Debugf("User %q not found in database; creating new user", externalUser.Login)
		user = &qf.User{
			Name:         externalUser.Name,
			Email:        externalUser.Email,
			AvatarURL:    externalUser.AvatarURL,
			Login:        externalUser.Login,
			ScmRemoteID:  externalUser.ID,
			RefreshToken: token.RefreshToken,
		}
		if err = db.CreateUser(user); err != nil {
			return nil, fmt.Errorf("failed to create remote identity for user %q: %w", externalUser.Login, err)
		}
		logger.Debugf("New user created: %v", user)

	default:
		return nil, fmt.Errorf("failed to fetch user %q for remote identity: %w", externalUser.Login, err)
	}
	logger.Debugf("Retry database lookup for user %q", externalUser.Login)
	return db.GetUserByRemoteIdentity(externalUser.ID)
}
