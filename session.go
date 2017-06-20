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
		return err
	}

	ss.Values[accessTokenKey] = accessToken
	ss.Save(r, w)

	return nil
}

func (s *Session) Logout(w http.ResponseWriter, r *http.Request) error {
	ss, err := s.store.Get(r, s.authSessionName)

	if err != nil {
		return err
	}

	delete(ss.Values, accessTokenKey)
	ss.Save(r, w)

	return nil
}

func (s *Session) IsLogin(r *http.Request) (bool, error) {
	ss, err := s.store.Get(r, s.authSessionName)

	if err != nil {
		return false, err
	}

	if _, ok := ss.Values[accessTokenKey]; !ok {
		return false, nil
	}

	return true, nil
}
