package handlers

import (
	"context"
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

type contextKey string

const userContextKey contextKey = "user"

// Auth tries to authenticate against an oauth2 provider.
func Auth(db aguis.UserDatabase, s *aguis.Session) http.Handler {
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

// AuthCallback handles the callback from an oauth2 provider.
func AuthCallback(db aguis.UserDatabase, s *aguis.Session) http.Handler {
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

// Authenticated ensures that only authenticated sessions are allowed to
// pass through to the endpoints that require authentication.
func Authenticated(m *mux.Router, db aguis.UserDatabase, s *aguis.Session) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := s.Whois(w, r)
		if err != nil {
			web.HTTPError(w, http.StatusInternalServerError, err)
			return
		}
		if strings.HasPrefix(r.URL.RequestURI(), "/api") && id == -1 {
			web.HTTPError(w, http.StatusUnauthorized, nil)
			return
		}

		if id >= 0 {
			user, err := db.GetUser(id)
			if err != nil {
				web.HTTPError(w, http.StatusInternalServerError, err)
				return
			}

			ctx := newUserContext(context.Background(), user)
			r = r.WithContext(ctx)
		}

		m.ServeHTTP(w, r)
	})
}

func newUserContext(ctx context.Context, u *aguis.User) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

func userFromContext(ctx context.Context) (*aguis.User, bool) {
	u, ok := ctx.Value(userContextKey).(*aguis.User)
	return u, ok
}

func getInteralUser(db aguis.UserDatabase, externalUser *goth.User) (*aguis.User, error) {
	switch externalUser.Provider {
	case "github":
		githubID, err := strconv.Atoi(externalUser.UserID)
		if err != nil {
			return nil, err
		}
		user, err := db.GetUserWithGithubID(githubID, externalUser.AccessToken)
		if err != nil {
			return nil, err
		}
		return user, nil
	case "faux": // Provider is only registered and reachable from tests.
		return &aguis.User{}, nil
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
