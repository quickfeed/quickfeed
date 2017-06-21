package aguis

import (
	"net/http"

	"github.com/gorilla/sessions"
)

const (
	accessTokenKey = iota
)

type Session struct {
	store sessions.Store

	authSessionName string
}

func NewSessionStore(store sessions.Store, authSessionName string) *Session {
	return &Session{
		store:           store,
		authSessionName: authSessionName,
	}
}

func (s *Session) Login(w http.ResponseWriter, r *http.Request, accessToken string) error {
	ss, err := s.store.Get(r, s.authSessionName)
	if err != nil {
		return ss.Save(r, w)
	}
	ss.Values[accessTokenKey] = accessToken
	return ss.Save(r, w)
}

func (s *Session) Logout(w http.ResponseWriter, r *http.Request) error {
	ss, err := s.store.Get(r, s.authSessionName)
	if err != nil {
		return ss.Save(r, w)
	}
	delete(ss.Values, accessTokenKey)
	return ss.Save(r, w)
}

func (s *Session) IsLogin(w http.ResponseWriter, r *http.Request) (ok bool, err error) {
	ss, err := s.store.Get(r, s.authSessionName)
	if err != nil {
		return false, ss.Save(r, w)
	}

	if _, ok := ss.Values[accessTokenKey]; !ok {
		return false, nil
	}

	return true, nil
}
