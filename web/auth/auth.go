package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/env"
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
			Domain:   env.Domain(),
			Path:     "/",
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

		// If no next URL is provided, next will be the root path ("/").
		// This is used to redirect the user back to the page they were on after logging in.
		// The next URL is sanitized to ensure it is a valid path.
		rawNext := r.URL.Query().Get("next")
		nextURL := SanitizeNext(rawNext)

		// Store the next URL in a (short-lived) cookie so we can redirect the user back to it in the callback handler.
		http.SetCookie(w, &http.Cookie{
			Name:     nextCookieName,
			Value:    url.QueryEscape(nextURL),
			Domain:   env.Domain(),
			Path:     "/",
			MaxAge:   300,
			Expires:  time.Now().Add(5 * time.Minute),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})

		redirectURL := authConfig.AuthCodeURL(secret)
		logger.Debugf("Redirecting to AuthURL: %v (nextURL=%q)", redirectURL, nextURL)
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

		cookie, err := tm.NewAuthCookie(user.GetID())
		if err != nil {
			authenticationError(logger, w, fmt.Errorf("failed to create authentication cookie for user %q: %w", externalUser.Login, err))
			return
		}
		http.SetCookie(w, cookie)

		// pull the redirect URL from the cookie (if any) and delete it
		redirectURL := "/"
		if c, err := r.Cookie(nextCookieName); err == nil {
			if v, e := url.QueryUnescape(c.Value); e == nil {
				redirectURL = SanitizeNext(v)
			}
			// delete cookie
			http.SetCookie(w, &http.Cookie{
				Name:     nextCookieName,
				Value:    "",
				Domain:   env.Domain(),
				Path:     "/",
				MaxAge:   -1,
				Expires:  time.Unix(0, 0),
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			})
		}
		http.Redirect(w, r, redirectURL, http.StatusFound)
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
func FetchExternalUser(token *oauth2.Token) (user *ExternalUser, err error) {
	req, err := http.NewRequest("GET", githubUserAPI, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create user request: %w", err)
	}
	token.SetAuthHeader(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send user request: %w", err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			// only overwrite error if there was no error before
			err = closeErr
		}
	}()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected OAuth provider response: %d (%s)", resp.StatusCode, resp.Status)
		return
	}
	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		err = fmt.Errorf("failed to read authentication response: %w", readErr)
		return
	}
	if jsonErr := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&user); jsonErr != nil {
		err = fmt.Errorf("failed to decode user information: %w", jsonErr)
	}
	return
}

// CheckExternalUser validates that the external user has required fields populated.
// It checks that Login, Name (with at least first and last name), and Email are not empty.
// The Email field must also be a valid email address.
func CheckExternalUser(externalUser *ExternalUser) error {
	if externalUser.Login == "" {
		return errors.New("missing login")
	}
	if externalUser.Name == "" {
		return errors.New("missing name")
	}
	// Check that name has at least two components (first and last name)
	nameParts := strings.Fields(externalUser.Name)
	if len(nameParts) < 2 {
		return errors.New("name must contain at least first and last name")
	}
	if externalUser.Email == "" {
		return errors.New("missing email")
	}
	// Validate that email is a proper email address
	if _, err := mail.ParseAddress(externalUser.Email); err != nil {
		return fmt.Errorf("invalid email address: %w", err)
	}
	return nil
}

// fetchUser saves or updates user information fetched from the OAuth provider in the database.
func fetchUser(logger *zap.SugaredLogger, db database.Database, token *oauth2.Token, externalUser *ExternalUser) (*qf.User, error) {
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
		// Validate external user information before creating account
		if err := CheckExternalUser(externalUser); err != nil {
			return nil, fmt.Errorf("cannot create account for user %q: %w", externalUser.Login, err)
		}
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

// SanitizeNext sanitizes the next URL to ensure it is a valid path.
// It removes leading/trailing whitespace, checks for absolute URLs, and cleans the path.
// If the next URL is invalid or otherwise unsafe, it returns the root path "/".
func SanitizeNext(next string) string {
	next = strings.TrimSpace(next)
	if next == "" {
		return "/"
	}

	u, err := url.Parse(next)
	if err != nil {
		return "/"
	}
	// Check if the URL is absolute and has a scheme (e.g., http, https)
	// If it starts with // or contains a backslash, treat it as invalid.
	if u.IsAbs() || strings.HasPrefix(next, "//") || strings.Contains(next, `\`) {
		return "/"
	}

	// If the path is empty or does not start with a slash, return the root path.
	if u.Path == "" || !strings.HasPrefix(u.Path, "/") {
		return "/"
	}

	// clean removes .., duplicate slashes, etc.
	cleaned := path.Clean(u.Path)
	if cleaned == "." { // path.Clean("/") == "/"; path.Clean("") == "."
		return "/"
	}
	u.Path = cleaned
	return u.String()
}
