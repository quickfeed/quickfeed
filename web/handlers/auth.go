package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/autograde/aguis"
	"github.com/autograde/aguis/web"
	"github.com/gorilla/mux"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

// AuthHandler tries to authenticate against an oauth2 provider.
func AuthHandler(db aguis.UserDatabase, s *aguis.Session) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, _ := s.Whois(w, r); id >= 0 {
			http.Redirect(w, r, "/", http.StatusFound)
		}

		externalUser, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			gothic.BeginAuthHandler(w, r)
			return
		}
		if err := login(w, r, db, s, &externalUser); err != nil {
			web.HTTPError(w, http.StatusInternalServerError, err)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	})
}

// AuthCallbackHandler handles the callback from an oauth2 provider.
func AuthCallbackHandler(db aguis.UserDatabase, s *aguis.Session) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, _ := s.Whois(w, r); id >= 0 {
			http.Redirect(w, r, "/", http.StatusFound)
		}

		externalUser, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusUnauthorized)
			return
		}
		if err := login(w, r, db, s, &externalUser); err != nil {
			web.HTTPError(w, http.StatusInternalServerError, err)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	})
}

// AuthenticatedHandler ensures that only authenticated sessions are allowed to
// pass through to the endpoints that require authentication.
func AuthenticatedHandler(m *mux.Router, s *aguis.Session) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := s.Whois(w, r)

		if err != nil {
			web.HTTPError(w, http.StatusInternalServerError, err)
			return
		}

		if strings.HasPrefix(r.RequestURI, "/api") && id == -1 {
			web.HTTPError(w, http.StatusForbidden, nil)
			return
		}
		m.ServeHTTP(w, r)
	})
}

func getInteralUser(db aguis.UserDatabase, user *goth.User) (*aguis.User, error) {
	switch user.Provider {
	case "github":
		githubID, err := strconv.Atoi(user.UserID)
		if err != nil {
			return nil, err
		}
		user, err := db.GetUserWithGithubID(githubID)
		if err != nil {
			return nil, err
		}
		return user, nil
	default:
		return nil, errors.New("provider not implemented")
	}
}

func login(
	w http.ResponseWriter, r *http.Request,
	db aguis.UserDatabase, s *aguis.Session, externalUser *goth.User,
) error {
	user, err := getInteralUser(db, externalUser)
	if err != nil {
		return err
	}
	if err := s.Login(w, r, user.ID); err != nil {
		return err
	}
	return nil
}
