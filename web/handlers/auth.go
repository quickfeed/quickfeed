package handlers

import (
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/autograde/aguis"
	"github.com/autograde/aguis/web"
	"github.com/gorilla/mux"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

// AuthHandler tries to authenticate against a oauth2 provider.
func AuthHandler(db aguis.UserDatabase, s *aguis.Session) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := tryAuthenticate(w, r, db, s)
		if err != nil {
			gothic.BeginAuthHandler(w, r)
		}
		serveInfo(w, user)
	})
}

// AuthCallbackHandler handles the callback from a oauth2 provider.
func AuthCallbackHandler(db aguis.UserDatabase, s *aguis.Session) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := tryAuthenticate(w, r, db, s)
		if err != nil {
			web.HTTPError(w, http.StatusInternalServerError, err)
			return
		}
		serveInfo(w, user)
	})
}

// AuthenticatedHandler ensures that only authenticated sessions are allowed to
// pass through to the endpoints that require authentication.
func AuthenticatedHandler(m *mux.Router, s *aguis.Session) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggedIn, err := s.LoggedIn(w, r)

		if err != nil {
			web.HTTPError(w, http.StatusInternalServerError, err)
			return
		}

		if strings.HasPrefix(r.RequestURI, "/api") && !loggedIn {
			web.HTTPError(w, http.StatusForbidden, nil)
			return
		}
		m.ServeHTTP(w, r)
	})
}

// Try to get the user without re-authenticating.
func tryAuthenticate(
	w http.ResponseWriter, r *http.Request,
	db aguis.UserDatabase, s *aguis.Session,
) (*goth.User, error) {
	user, err := gothic.CompleteUserAuth(w, r)

	if err != nil {
		return nil, err
	}

	switch user.Provider {
	case "github":
		if err := loginGithub(db, user.UserID); err != nil {
			return nil, err
		}
		if err := s.Login(w, r); err != nil {
			return nil, err
		}
		return &user, nil
	default:
		return nil, errors.New(user.Provider + " provider not implemented")
	}
}

func loginGithub(db aguis.UserDatabase, userID string) error {
	githubID, err := strconv.Atoi(userID)
	if err != nil {
		return err
	}
	_, err = db.GetUserWithGithubID(githubID)
	if err != nil {
		return err
	}
	return nil
}

func serveInfo(w http.ResponseWriter, user *goth.User) {
	t, _ := template.New("").Parse(`
	<p><a href="/logout">logout</a></p>
	<p>Name: {{.Name}}</p>
	<p>NickName: {{.NickName}}</p>
	<p>UserID: {{.UserID}}</p>
	<p>AccessToken: {{.AccessToken}}</p>
	`)

	t.Execute(w, user)
}
