package aguis

import (
	"net/http"

	"github.com/gorilla/sessions"
)

const (
	accessTokenKey = iota
)

// Session holds a session store which stores user sessions.
type Session struct {
	store sessions.Store

	authSessionName string
}

// NewSessionStore initializes a new session store for saving sessions.
func NewSessionStore(store sessions.Store, authSessionName string) *Session {
	return &Session{
		store:           store,
		authSessionName: authSessionName,
	}
}

// Login logs a user in by storing the users access token in the session.
func (s *Session) Login(w http.ResponseWriter, r *http.Request, accessToken string) error {
	ss, err := s.store.Get(r, s.authSessionName)
	if err != nil {
		return ss.Save(r, w)
	}
	ss.Values[accessTokenKey] = accessToken
	return ss.Save(r, w)
}

// Logout logs a user out by deleting the users access token from the session.
func (s *Session) Logout(w http.ResponseWriter, r *http.Request) error {
	ss, err := s.store.Get(r, s.authSessionName)
	if err != nil {
		return ss.Save(r, w)
	}
	delete(ss.Values, accessTokenKey)
	return ss.Save(r, w)
}

// LoggedIn returns true if a user is logged in.
func (s *Session) LoggedIn(w http.ResponseWriter, r *http.Request) (ok bool, err error) {
	ss, err := s.store.Get(r, s.authSessionName)
	if err != nil {
		return false, ss.Save(r, w)
	}

	if _, ok := ss.Values[accessTokenKey]; !ok {
		return false, nil
	}

	return true, nil
}
