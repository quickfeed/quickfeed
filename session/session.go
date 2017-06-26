package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

const userIDKey = iota

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

// Login logs a user in by associating the provided user id and session.
func (s *Session) Login(w http.ResponseWriter, r *http.Request, userID int) error {
	ss, err := s.store.Get(r, s.authSessionName)
	if err != nil {
		return ss.Save(r, w)
	}
	ss.Values[userIDKey] = userID
	return ss.Save(r, w)
}

// Logout logs a user out by deleting the user's id from the session.
func (s *Session) Logout(w http.ResponseWriter, r *http.Request) error {
	ss, err := s.store.Get(r, s.authSessionName)
	if err != nil {
		return ss.Save(r, w)
	}
	delete(ss.Values, userIDKey)
	return ss.Save(r, w)
}

// Whois returns a user id if the session is logged in, otherwise -1 is
// returned.
func (s *Session) Whois(w http.ResponseWriter, r *http.Request) (int, error) {
	ss, err := s.store.Get(r, s.authSessionName)
	if err != nil {
		return -1, ss.Save(r, w)
	}
	if id, ok := ss.Values[userIDKey]; ok {
		return id.(int), nil
	}
	return -1, nil
}
