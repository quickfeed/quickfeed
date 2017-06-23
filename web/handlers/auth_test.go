package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/autograde/aguis"
	"github.com/autograde/aguis/web/handlers"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/faux"
)

const (
	loginSession    = "login"
	authURL         = "/auth?provider=faux"
	apiURL          = "/api/v1/test"
	fauxSessionName = "faux" + gothic.SessionName
	fauxSessionKey  = "faux"
)

func init() {
	goth.UseProviders(&faux.Provider{})
}

func TestAuthHandlerRedirect(t *testing.T) {
	w := httptest.NewRecorder()

	store := newStore()
	gothic.Store = store
	s := aguis.NewSessionStore(store, loginSession)

	authHandler := handlers.Auth(newDB(t), s)
	authHandler.ServeHTTP(w, newAuthRequest(t))
	checkResponse(t, w.Code, http.StatusTemporaryRedirect, w.Body.String())
}

func TestAuthCallbackHandlerUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()

	store := newStore()
	gothic.Store = store
	s := aguis.NewSessionStore(store, loginSession)

	authHandler := handlers.AuthCallback(newDB(t), s)
	authHandler.ServeHTTP(w, newAuthRequest(t))
	checkResponse(t, w.Code, http.StatusUnauthorized, w.Body.String())
}

func TestAuthHandlerLoggedIn(t *testing.T) {
	testAuthHandlerLoggedIn(t, handlers.Auth)
}

func TestAuthCallbackHandlerLoggedIn(t *testing.T) {
	testAuthHandlerLoggedIn(t, handlers.AuthCallback)
}

func testAuthHandlerLoggedIn(t *testing.T, newHandler func(db aguis.UserDatabase, s *aguis.Session) http.Handler) {
	w := httptest.NewRecorder()
	r := newAuthRequest(t)

	store := newStore()
	gothic.Store = store
	s := aguis.NewSessionStore(store, loginSession)

	if err := s.Login(w, r, 0); err != nil {
		t.Error(err)
	}

	authHandler := newHandler(newDB(t), s)
	authHandler.ServeHTTP(w, r)
	checkResponse(t, w.Code, http.StatusFound, w.Body.String())
}

func TestAuthHandlerAuthenticated(t *testing.T) {
	testAuthHandlerAuthenticated(t, handlers.Auth)
}

func TestAuthCallbackHandlerAuthenticated(t *testing.T) {
	testAuthHandlerAuthenticated(t, handlers.AuthCallback)
}

func testAuthHandlerAuthenticated(t *testing.T, newHandler func(db aguis.UserDatabase, s *aguis.Session) http.Handler) {
	w := httptest.NewRecorder()
	r := newAuthRequest(t)

	store := newStore()
	gothic.Store = store

	fauxSession := faux.Session{}
	authSession, _ := store.Get(r, fauxSessionName)
	authSession.Values[fauxSessionKey] = fauxSession.Marshal()
	if err := authSession.Save(r, w); err != nil {
		t.Error(err)
	}

	s := aguis.NewSessionStore(store, loginSession)

	authHandler := newHandler(newDB(t), s)
	authHandler.ServeHTTP(w, r)
	checkResponse(t, w.Code, http.StatusFound, w.Body.String())
}

func TestAuthenticatedHandlerAllowed(t *testing.T) {
	r := newAuthRequest(t)
	testAuthenticatedHandler(t, r, true, false)
}

func TestAuthenticatedHandlerUnauthorized(t *testing.T) {
	r := newAPIRequest(t)
	testAuthenticatedHandler(t, r, false, false)
}

func TestAuthenticatedHandlerLoggedIn(t *testing.T) {
	r := newAPIRequest(t)
	testAuthenticatedHandler(t, r, true, true)
}

func testAuthenticatedHandler(t *testing.T, r *http.Request, allowed, loggedIn bool) {
	w := httptest.NewRecorder()

	store := newStore()
	gothic.Store = store
	s := aguis.NewSessionStore(store, loginSession)

	if loggedIn {
		if err := s.Login(w, r, 0); err != nil {
			t.Fatal(err)
		}
	}

	m := mux.NewRouter()
	m.PathPrefix("/").HandlerFunc(func(http.ResponseWriter, *http.Request) {})

	authHandler := handlers.Authenticated(m, s)
	authHandler.ServeHTTP(w, r)

	wantCode := http.StatusUnauthorized
	if allowed {
		wantCode = http.StatusOK
	}
	checkResponse(t, w.Code, wantCode, w.Body.String())
}

func newAuthRequest(t *testing.T) *http.Request {
	r, err := http.NewRequest(http.MethodGet, authURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func newAPIRequest(t *testing.T) *http.Request {
	r, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func newDB(t *testing.T) aguis.UserDatabase {
	db, err := aguis.NewStructDB("", false, log.NewNopLogger())
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func checkResponse(t *testing.T, haveCode, wantCode int, body string) {
	if haveCode != wantCode {
		t.Errorf("have status code %d want %d", haveCode, wantCode)
	}

	if wantCode == http.StatusOK {
		return
	}

	mustContain := http.StatusText(wantCode)
	if !strings.Contains(body, mustContain) {
		t.Errorf("have %s which does not contain '%s'", body, mustContain)
	}
}

type testStore struct {
	store map[*http.Request]*sessions.Session
}

func newStore() *testStore {
	return &testStore{
		make(map[*http.Request]*sessions.Session),
	}
}

func (ts testStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	s := ts.store[r]
	if s == nil {
		s, err := ts.New(r, name)
		return s, err
	}
	return s, nil
}

func (ts testStore) New(r *http.Request, name string) (*sessions.Session, error) {
	s := sessions.NewSession(ts, name)
	s.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 30,
	}
	ts.store[r] = s
	return s, nil
}

func (ts testStore) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	ts.store[r] = s
	return nil
}
