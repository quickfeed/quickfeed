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
	"time"

	pb "github.com/autograde/quickfeed/ag"
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

// // PreAuth checks the current user session and executes the next handler if none
// // was found for the given provider.
// func PreAuth(logger *zap.SugaredLogger, db database.Database) echo.MiddlewareFunc {
// 	logger.Debug("PREAUTH STARTED")
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {
// 			logger.Debug("PRE AUTH")
// 			sess, err := session.Get(SessionKey, c)
// 			if err != nil {
// 				logger.Error(err.Error())
// 				if err := sess.Save(c.Request(), c.Response()); err != nil {
// 					logger.Error(err.Error())
// 					return err
// 				}
// 				return next(c)
// 			}
// 			logger.Debug(sessionData(sess))

// 			if i, ok := sess.Values[UserKey]; ok {
// 				// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
// 				us := i.(*UserSession)
// 				logger.Debug(us)
// 				user, err := db.GetUser(us.ID)
// 				if err != nil {
// 					logger.Error(err.Error())
// 					return OAuth2Logout(logger)(c)
// 				}
// 				logger.Debugf("User: %v", user)
// 			}
// 			logger.Debug("PRE AUTH next")
// 			return next(c)
// 		}
// 	}
// }

// func sessionData(session *sessions.Session) string {
// 	if session == nil {
// 		return "<nil>"
// 	}
// 	out := "Values: "
// 	for k, v := range session.Values {
// 		out += fmt.Sprintf("<%s: %v>, ", k, v)
// 	}
// 	out += "Options: "
// 	out += fmt.Sprintf("<%s: %v>, ", "MaxAge", session.Options.MaxAge)
// 	out += fmt.Sprintf("<%s: %v>, ", "Path", session.Options.Path)
// 	out += fmt.Sprintf("<%s: %v>, ", "Domain", session.Options.Domain)
// 	out += fmt.Sprintf("<%s: %v>, ", "Secure", session.Options.Secure)
// 	out += fmt.Sprintf("<%s: %v>, ", "HttpOnly", session.Options.HttpOnly)
// 	out += fmt.Sprintf("<%s: %v>, ", "SameSite", session.Options.SameSite)

// 	return fmt.Sprintf("Session: ID=%s, IsNew=%t, %s", session.ID, session.IsNew, out)
// }

// OAuth2Login tries to authenticate against an oauth2 provider.
func OAuth2Login(logger *zap.SugaredLogger, db database.Database, config oauth2.Config, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("LOGIN STARTED")
		if r.Method != "GET" {
			logger.Errorf("GitHub login failed: request method %s", r.Method)
			http.Redirect(w, r, "/", http.StatusUnauthorized)
		}
		// TODO(vera): adapt to use with other providers if needed
		provider := "github"

		// TODO(vera): make sure teacher suffix no longer necessary
		// var teacher int
		// if strings.HasSuffix(provider, TeacherSuffix) {
		// 	teacher = 1
		// }
		// logger.Debugf("Provider: %v ; Teacher: %v", provider, teacher)
		// qv := r.URL.Query()
		// logger.Debugf("qv: %v", qv)
		// // redirect := extractRedirectURL(r, Redirect)
		logger.Debugf("redirect: %v", config.RedirectURL)
		// TODO(vera): Add a random string to protect against CSRF.
		// qv.Set(State, strconv.Itoa(teacher)+config.RedirectURL)
		// logger.Debugf("State: %v", strconv.Itoa(teacher)+config.RedirectURL)
		// r.URL.RawQuery = qv.Encode()
		// logger.Debugf("RawQuery: %v", r.URL.RawQuery)
		redirectURL := config.AuthCodeURL(secret)
		logger.Debugf("Redirecting to %s to perform authentication; AuthURL: %v", provider, redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

// OAuth2Callback handles the callback from an oauth2 provider.
func OAuth2Callback(logger *zap.SugaredLogger, db database.Database, config oauth2.Config, app *scm.GithubApp, tokens *TokenManager, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("CALLBACK STARTED")
		if r.Method != "GET" {
			logger.Errorf("GitHub login failed: request method %s", r.Method)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		logger.Debug("OAuth2Callback: started")
		// qv := r.URL.Query()
		// logger.Debugf("qv: %v", qv)
		// redirect, teacher := extractState(r, State)
		// logger.Debugf("Redirect: %v ; Teacher: %t", redirect, teacher)

		provider := "github"
		// // TODO(vera): remove teacher suffix if not needed
		// // Add teacher suffix if upgrading scope.
		// if teacher {
		// 	qv.Set("provider", provider+TeacherSuffix)
		// 	logger.Debugf("Set('provider') = %v", provider+TeacherSuffix)
		// }
		// r.URL.RawQuery = qv.Encode()
		// logger.Debugf("RawQuery: %v", r.URL.RawQuery)

		// Complete authentication.
		// parse request for code and state
		if err := r.ParseForm(); err != nil {
			logger.Error("GitHub login failed: error parsing authentication code")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		logger.Debug("VALIDATING STATE") // tmp
		// validate state
		callbackSecret := r.FormValue("state")
		logger.Debug("Callback: got state in request: ", callbackSecret) // tmp
		if callbackSecret != secret {
			logger.Errorf("GitHub login failed: secrets don't match: expected %s, got %s", secret, callbackSecret)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		logger.Debug("EXCHANGING CODE FOR TOKEN") // tmp
		// exchange code for token
		code := r.FormValue("code")
		if code == "" {
			logger.Error("GitHub login failed: received empty code")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		logger.Debug("CODE RECEIVED, PROCEED") // tmp
		githubToken, err := config.Exchange(context.Background(), code)
		if err != nil {
			logger.Errorf("GitHub login failed: cannot exchange token: %s", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		logger.Debugf("Successfully fetched access token: %s", githubToken.AccessToken) // tmp

		// get user info with the token
		logger.Debugf("Making user request: want url https://api.github.com/user, have url %s", app.GetUserURL())
		req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
		if err != nil {
			logger.Errorf("GitHub login failed: failed to make user request: %s", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		logger.Debugf("REQUEST HEADER want (%s), have (%s)", "Bearer "+githubToken.AccessToken, fmt.Sprintf("Bearer %s", githubToken.AccessToken))
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
			logger.Errorf("GitHub login failed: failed to send user request: %s", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Errorf("GitHub login failed: API responded with status: %d: %s", resp.StatusCode, resp.Status)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		respBits, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error reading response bits from user API: ", err.Error())
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		externalUser := &externalUser{}
		if err := json.NewDecoder(bytes.NewReader(respBits)).Decode(&externalUser); err != nil {
			logger.Errorf("GitHub login failed: failed to decode user information: %s", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		logger.Debugf("externalUser: %v", lg.IndentJson(externalUser))
		// logger.Debugf("EXTRACTED set-cookie token: %s", extractToken(w)) // tmp

		// TODO(vera): this is only used when trying to log in with github explicitly. If a user already has a JWT
		// he must be logged in automatically when loading the page, no use to check it here
		// for _, cookie := range r.Cookies() {
		// 	logger.Debugf("AUTH: Checking cookie with name %s: %+v", tokens.cookieName, cookie) // tmp
		// 	if cookie.Name == tokens.cookieName {
		// 		userToken = cookie.Value
		// 	}
		// }
		// logger.Debugf("EXTRACTED auth cookie", userToken) // tmp

		// There is already a cookie with JWT, make sure the user exists in the database
		userToken := extractToken(r, tokens.cookieName)
		logger.Debugf("GitHub login: extracted token from request: %s", userToken)
		if userToken != "" {
			claims, err := tokens.GetClaims(userToken)
			if err != nil {
				logger.Errorf("GitHub login failed: failed to read user claims: %s", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// TODO(vera): check that user in the claims and in remote ID is the same user

			if err := db.AssociateUserWithRemoteIdentity(claims.UserID, provider, externalUser.ID, githubToken.AccessToken); err != nil {
				logger.Debugf("Associate failed: %d, %s, %d, %s", claims.UserID, provider, externalUser.ID, githubToken.AccessToken)
				logger.Errorf("GitHub login failed: failed to associate user with remote identity: %s", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
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
		authToken := tokens.NewToken(claims)
		logger.Debugf("Created new JWT for user %s: %+v", user.Login, authToken)
		cookie, err := tokens.NewTokenCookie(context.Background(), authToken)
		if err != nil {
			logger.Errorf("GitHub login failed: failed to make token cookie for user %v: %s", externalUser, err)
			// TODO(vera): this pattern for handling auth errors might be better
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		logger.Debugf("setting cookie: %+v", cookie)
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// // AccessControl returns an access control middleware. Given a valid context
// // with sufficient access the next handler is called. Missing or invalid
// // credentials results in a 401 unauthorized response.
// func AccessControl(logger *zap.SugaredLogger, db database.Database) echo.MiddlewareFunc {
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {
// 			logger.Debug("ACCESS CONTROL")
// 			sess, err := session.Get(SessionKey, c)
// 			if err != nil {
// 				logger.Error(err.Error())
// 				// Save fixes the session if it has been modified
// 				// or it is no longer valid due to newUserSess change of keys.
// 				if err := sess.Save(c.Request(), c.Response()); err != nil {
// 					logger.Error(err.Error())
// 					return err
// 				}
// 				return next(c)
// 			}
// 			logger.Debug(sessionData(sess))

// 			i, ok := sess.Values[UserKey]
// 			if !ok {
// 				return next(c)
// 			}

// 			// If type assertion fails, the recover middleware will catch the panic and log a stack trace.
// 			us := i.(*UserSession)
// 			logger.Debug(us)
// 			user, err := db.GetUser(us.ID)
// 			if err != nil {
// 				logger.Error(err.Error())
// 				// Invalidate session. This could happen if the user has been entirely remove
// 				// from the database, but a valid session still exists.
// 				if err == gorm.ErrRecordNotFound {
// 					logger.Error(err.Error())
// 					return OAuth2Logout(logger)(c)
// 				}
// 				logger.Error(echo.ErrUnauthorized.Error())
// 				return next(c)
// 			}
// 			c.Set(UserKey, user)

// 			// TODO: Add access control list.
// 			// - Extract endpoint.
// 			// - Verify whether the user has sufficient rights. This
// 			//   can be a simple hash map. A user should be able to
// 			//   access /users/:uid if the user's id is uid.
// 			//   - Not authorized: return c.NoContent(http.StatusUnauthorized)
// 			//   - Authorized: return next(c)
// 			return next(c)
// 		}
// 	}
// }

// func extractRedirectURL(r *http.Request, key string) string {
// 	// TODO: Validate redirect URL.

// 	url := r.URL.Query().Get(key)
// 	if url == "" {
// 		url = "/"
// 	}
// 	return url
// }

// func extractState(r *http.Request, key string) (redirect string, teacher bool) {
// 	// TODO: Validate redirect URL.
// 	url := r.URL.Query().Get(key)
// 	log.Printf("EXTRACT STATE: url for key (%s) is %s", key, url)
// 	log.Printf("URL [1:], [:1] is %s, %s", url[1:], url[:1])
// 	teacher = url != "" && url[:1] == "1"

// 	if url == "" || url[1:] == "" {
// 		return "/", teacher
// 	}
// 	return url[1:], teacher
// }

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

// var (
// 	ErrInvalidSessionCookie = status.Errorf(codes.Unauthenticated, "Request does not contain a valid session cookie.")
// 	ErrContextMetadata      = status.Errorf(codes.Unauthenticated, "Could not obtain metadata from context")
// )

// func UserVerifier() grpc.UnaryServerInterceptor {
// 	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
// 		meta, ok := metadata.FromIncomingContext(ctx)
// 		if !ok {
// 			return nil, ErrContextMetadata
// 		}
// 		newMeta, err := userValidation(meta)
// 		if err != nil {
// 			return nil, err
// 		}
// 		// create new context with user id instead of cookie for use internally
// 		newCtx := metadata.NewIncomingContext(ctx, newMeta)
// 		resp, err := handler(newCtx, req)
// 		return resp, err
// 	}
// }

// // userValidation returns modified metadata containing a valid user.
// // An error is returned if the user is not authenticated.
// func userValidation(meta metadata.MD) (metadata.MD, error) {
// 	for _, cookie := range meta.Get(Cookie) {
// 		if user := Get(cookie); user > 0 {
// 			meta.Set(UserKey, strconv.FormatUint(user, 10))
// 			return meta, nil
// 		}
// 	}
// 	return nil, ErrInvalidSessionCookie
// }
