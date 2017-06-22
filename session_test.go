package aguis_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/autograde/aguis"
	"github.com/gorilla/sessions"
)

const loginSession = "login"

func TestLoginLogout(t *testing.T) {
	store := sessions.NewCookieStore([]byte{})
	s := aguis.NewSessionStore(store, loginSession)

	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest(http.MethodGet, "/auth/github", nil)

	if err := s.Login(w1, r1, 0); err != nil {
		t.Error(err)
	}

	var cookie *http.Cookie
	for _, c := range w1.Result().Cookies() {
		if c.Name == loginSession {
			cookie = c
		}
	}
	if cookie == nil {
		t.Error("have 'login cookie not set' want 'login cookie set")
	}

	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	r2.AddCookie(cookie)

	id, err := s.Whois(w2, r2)
	if err != nil {
		t.Errorf("have '%s' want 'no error'", err)
	}
	if id > 0 {
		t.Error("have 'user not logged in' want 'user logged in'")
	}

	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest(http.MethodGet, "/logout", nil)

	if err := s.Logout(w3, r3); err != nil {
		t.Error(err)
	}

	var cookie2 *http.Cookie
	for _, c := range w1.Result().Cookies() {
		if c.Name == loginSession {
			cookie2 = c
		}
	}
	if cookie2 == nil {
		t.Error("have 'login cookie not set' want 'login cookie set")
	}

	w4 := httptest.NewRecorder()
	r4 := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	r4.AddCookie(cookie2)

	id, err = s.Whois(w4, r4)
	if err != nil {
		t.Errorf("have '%s' want 'no error'", err)
	}
	if id > 0 {
		t.Error("have 'user logged in' want 'user not logged in'")
	}
}
