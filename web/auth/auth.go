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
	"strings"
	"time"

	pb "github.com/autograde/quickfeed/ag/types"
	"github.com/autograde/quickfeed/database"
	lg "github.com/autograde/quickfeed/log"
	"github.com/autograde/quickfeed/scm"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
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
	log.Println("NEW USER SESSION")
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
func OAuth2Logout(logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO(vera): get token from cookie, set a new empty expired token,
		// redirect back

		// if i, ok := sess.Values[UserKey]; ok {
		// 	// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
		// 	us := i.(*UserSession)
		// 	logger.Debug(us)
		// 	// Invalidate gothic user sessions.
		// 	for provider := range us.Providers {
		// 		sess, err := session.Get(provider+gothic.SessionName, c)
		// 		if err != nil {
		// 			logger.Error(err.Error())
		// 			return err
		// 		}
		// 		logger.Debug(sessionData(sess))

		// 		sess.Options.MaxAge = -1
		// 		sess.Values = make(map[interface{}]interface{})
		// 		if err := sess.Save(r, w); err != nil {
		// 			logger.Error(err.Error())
		// 		}
		// 	}
		// }
		// // Invalidate our user session.
		// sess.Options.MaxAge = -1
		// sess.Values = make(map[interface{}]interface{})
		// if err := sess.Save(r, w); err != nil {
		// 	logger.Error(err.Error())
		// }
		// return c.Redirect(http.StatusFound, extractRedirectURL(r, Redirect))
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
		provider := strings.Split(r.URL.Path, "/")[2]
		// TODO(vera): Add a random string to protect against CSRF.
		redirectURL := config.AuthCodeURL(secret)
		logger.Debugf("Redirecting to %s to perform authentication; AuthURL: %v", provider, redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

// OAuth2Callback handles the callback from an oauth2 provider.
func OAuth2Callback(logger *zap.SugaredLogger, db database.Database, config oauth2.Config, app *scm.GithubApp, tokens *TokenManager, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			unauthorized(logger, w, callback, "request method: %s", r.Method)
			return
		}
		provider := strings.Split(r.URL.Path, "/")[2]
		if err := r.ParseForm(); err != nil {
			unauthorized(logger, w, callback, "error parsing authentication code: %v", err)
			return
		}
		// Make sure the callback is coming from the provider by validating the state.
		callbackSecret := r.FormValue("state")
		if callbackSecret != secret {
			unauthorized(logger, w, callback, "mismatching secrets: expected %s, got %s", secret, callbackSecret)
			return
		}
		// Exchange code for access token.
		code := r.FormValue("code")
		if code == "" {
			unauthorized(logger, w, callback, "empty code")
			return
		}
		githubToken, err := config.Exchange(context.Background(), code)
		if err != nil {
			unauthorized(logger, w, callback, "could not exchange token: %v", err)
			return
		}
		// Use access token to fetch information about the GitHub user.
		req, err := http.NewRequest("GET", app.GetUserURL(), nil)
		if err != nil {
			unauthorized(logger, w, callback, "failed to create user request: %v", err)
			return
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", githubToken.AccessToken))
		// TODO(vera): this http client has only  one purpose: to fetch user data from github on auth. Somehow, the http client that
		// alredy exists for the github app fails to make this request. However, this client will be used every time a user logs into
		// the system without a cookie, which means it is dumb to make a new client every time -> we have to create one when the server starts
		// and reuse it (for example as a part of the github app struct) or find out what can be done to use the app client for this request
		httpClient := &http.Client{
			Timeout: time.Second * 10,
		}
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
		remote := &pb.RemoteIdentity{
			Provider:    provider,
			RemoteID:    externalUser.ID,
			AccessToken: githubToken.AccessToken,
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
	}
}

// extractToken returns a request cookie with given name, or an empty string
// is cookie does not exist
func extractToken(r *http.Request, cookieName string) string {
	for _, cookie := range r.Cookies() {
		if cookie.Name == cookieName {
			return cookie.Value
		}
	}
	return ""
}
