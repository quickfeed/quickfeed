package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	pb "github.com/autograde/quickfeed/ag/types"
	"github.com/autograde/quickfeed/database"
	lg "github.com/autograde/quickfeed/log"
	"github.com/autograde/quickfeed/web/auth/tokens"
	"github.com/autograde/quickfeed/web/config"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

var (
	githubUserAPI = "https://api.github.com/user"
	httpClient    *http.Client
)

func init() {
	httpClient = &http.Client{
		Timeout: time.Second * 10,
	}
}

// OAuth2Logout replace the token cookie with an empty cookie to log the user out.
func OAuth2Logout(logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// redirect back
		newCookie := &http.Cookie{
			Name:     "auth",
			Value:    "",
			MaxAge:   0,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   true,
		}
		http.SetCookie(w, newCookie)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

const (
	login    = "Login failed:"
	callback = "Callback failed:"
)

func unauthorized(logger *zap.SugaredLogger, w http.ResponseWriter, prefix, format string, a ...interface{}) {
	logger.Error(prefix + fmt.Sprintf(format, a...))
	w.WriteHeader(http.StatusUnauthorized)
}

// OAuth2Login tries to authenticate against an oauth2 provider.
func OAuth2Login(logger *zap.SugaredLogger, db database.Database, config oauth2.Config, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			unauthorized(logger, w, login, "request method: %s", r.Method)
			return
		}
		// This will work for any provider oauth2.Config is configured for.
		// This check here is just for logging.
		provider := strings.Split(r.URL.Path, "/")[3]
		redirectURL := config.AuthCodeURL(secret)
		logger.Debugf("Redirecting to %s to perform authentication; AuthURL: %v", provider, redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

// OAuth2Callback handles the callback from an oauth2 provider.
func OAuth2Callback(logger *zap.SugaredLogger, db database.Database, scmConfig oauth2.Config, tokens *tokens.TokenManager, config *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			unauthorized(logger, w, callback, "request method: %s", r.Method)
			return
		}
		provider := strings.Split(r.URL.Path, "/")[2]
		switch provider {
		case "github":
			if err := r.ParseForm(); err != nil {
				unauthorized(logger, w, callback, "error parsing callback response: %v", err)
				return
			}
			// Make sure the callback is coming from the provider by validating the state.
			callbackSecret := r.FormValue("state")
			if callbackSecret != config.Secrets.CallbackSecret {
				unauthorized(logger, w, callback, "mismatching secrets: expected %s, got %s", config.Secrets.CallbackSecret, callbackSecret)
				return
			}
			// Exchange code for access token.
			code := r.FormValue("code")
			if code == "" {
				unauthorized(logger, w, callback, "empty code")
				return
			}
			githubToken, err := scmConfig.Exchange(context.Background(), code)
			if err != nil {
				unauthorized(logger, w, callback, "could not exchange token: %v", err)
				return
			}
			// Use access token to fetch information about the GitHub user.
			req, err := http.NewRequest("GET", githubUserAPI, nil)
			if err != nil {
				unauthorized(logger, w, callback, "failed to create user request: %v", err)
				return
			}
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", githubToken.AccessToken))

			resp, err := httpClient.Do(req)
			if err != nil {
				unauthorized(logger, w, callback, "failed to send user request: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				unauthorized(logger, w, callback, "API response status: %d: %s", resp.StatusCode, resp.Status)
				return
			}
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				unauthorized(logger, w, callback, "error reading response from %s: %v", provider, err)
				return
			}
			externalUser := &externalUser{}
			if err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&externalUser); err != nil {
				unauthorized(logger, w, callback, "failed to decode user information: %v", err)
				return
			}
			logger.Debugf("externalUser: %v", lg.IndentJson(externalUser))
			accessToken := githubToken.AccessToken
			if config.WithEncryption() {
				accessToken, err = config.Cipher(accessToken)
				if err != nil {
					unauthorized(logger, w, callback, "failed to encrypt access token: %v", err)
					return
				}
			}
			remote := &pb.RemoteIdentity{
				Provider:    provider,
				RemoteID:    externalUser.ID,
				AccessToken: accessToken,
			}
			// Try to get user from database.
			user, err := db.GetUserByRemoteIdentity(remote)
			switch err {
			case nil:
				logger.Debugf("Updating access token for user in database: %v", user)
				if err = db.UpdateAccessToken(remote); err != nil {
					unauthorized(logger, w, callback, "failed to update access token for user %v: %v", externalUser, err)
					return
				}
				logger.Debugf("Access token updated: %v", remote)

			case gorm.ErrRecordNotFound:
				logger.Debug("User not found in database; Creating new user")
				user = &pb.User{
					Name:      externalUser.Name,
					Email:     externalUser.Email,
					AvatarURL: externalUser.AvatarURL,
					Login:     externalUser.Login,
				}
				if err = db.CreateUserFromRemoteIdentity(user, remote); err != nil {
					unauthorized(logger, w, callback, "failed to create remote identity for user %v: %v", externalUser, err)
					return
				}
				logger.Debugf("New user created: %v, remote: %v", user, remote)

			default:
				unauthorized(logger, w, callback, "failed to fetch user by remote identity: %v", err)
				return
			}
			// in case this is a new user we need a user object with full information,
			// otherwise frontend will get user object where only name, email and url are set.
			user, err = db.GetUserByRemoteIdentity(remote)
			if err != nil {
				unauthorized(logger, w, callback, "failed to fetch user %v from database: %v", externalUser, err)
				return
			}
			logger.Debugf("Fetching full user info for %v, user: %v", remote, user)

			authToken, err := tokens.NewTokenCookie(user.ID)
			if err != nil {
				unauthorized(logger, w, callback, "failed to make token cookie for user %v: %v", externalUser, err)
				return
			}
			logger.Debugf("setting cookie: %+v", authToken)
			http.SetCookie(w, authToken)
			http.Redirect(w, r, "/", http.StatusFound)
		default:
			unauthorized(logger, w, callback, "unsupported provider: %s", provider)
		}
	}
}
