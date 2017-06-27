package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/web/auth"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/faux"
)

const (
	authURL         = "/auth?provider=faux"
	logoutURL       = "/logout?provider=faux"
	fauxSessionName = "faux" + gothic.SessionName
	fauxSessionKey  = "faux"
)

func init() {
	goth.UseProviders(&faux.Provider{})
}

func TestOAuth2Logout(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, logoutURL, nil)
	w := httptest.NewRecorder()

	store := newStore()
	gothic.Store = store

	e := echo.New()
	c := e.NewContext(r, w)

	fauxSession := faux.Session{}
	s, _ := store.Get(r, fauxSessionName)
	s.Values[fauxSessionKey] = fauxSession.Marshal()
	if err := s.Save(r, w); err != nil {
		t.Error(err)
	}

	if err := store.login(c); err != nil {
		t.Error(err)
	}

	ns := len(store.store[r].Values)
	// Want gothic session and user session.
	if ns != 2 {
		t.Errorf("have %d sessions want %d", ns, 2)
	}

	authHandler := auth.OAuth2Logout()
	withSession := session.Middleware(store)(authHandler)

	if err := withSession(c); err != nil {
		t.Error(err)
	}

	ns = len(store.store[r].Values)
	// Sessions should be cleared.
	if ns != 0 {
		t.Errorf("have %d sessions want %d", ns, 0)
	}
}

func TestOAuth2LoginRedirect(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, authURL, nil)
	w := httptest.NewRecorder()

	store := newStore()
	gothic.Store = store

	e := echo.New()
	c := e.NewContext(r, w)

	authHandler := auth.OAuth2Login(newDB(t))
	withSession := session.Middleware(store)(authHandler)
	if err := withSession(c); err != nil {
		t.Error(err)
	}

	assertCode(t, w.Code, http.StatusTemporaryRedirect)
}

func TestOAuth2CallbackUnauthorized(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, authURL, nil)
	w := httptest.NewRecorder()

	store := newStore()
	gothic.Store = store

	e := echo.New()
	c := e.NewContext(r, w)

	authHandler := auth.OAuth2Callback(newDB(t))
	withSession := session.Middleware(store)(authHandler)
	if err := withSession(c); err != echo.ErrUnauthorized {
		t.Errorf("have error '%s' want '%s'", err, echo.ErrUnauthorized)
	}
}

func TestOAuth2LoginLoggedIn(t *testing.T) {
	testOAuth2LoggedIn(t, auth.OAuth2Login)
}

func TestOAuth2CallbackLoggedIn(t *testing.T) {
	testOAuth2LoggedIn(t, auth.OAuth2Callback)
}

func testOAuth2LoggedIn(t *testing.T, newHandler func(db database.UserDatabase) echo.HandlerFunc) {
	r := httptest.NewRequest(http.MethodGet, authURL, nil)
	w := httptest.NewRecorder()

	store := newStore()
	gothic.Store = store

	e := echo.New()
	c := e.NewContext(r, w)

	if err := store.login(c); err != nil {
		t.Error(err)
	}

	authHandler := newHandler(newDB(t))
	withSession := session.Middleware(store)(authHandler)

	if err := withSession(c); err != nil {
		t.Error(err)
	}

	assertCode(t, w.Code, http.StatusFound)
}

func TestOAuth2LoginAuthenticated(t *testing.T) {
	testOAuth2Authenticated(t, auth.OAuth2Login)
}

func TestOAuth2CallbackAuthenticated(t *testing.T) {
	testOAuth2Authenticated(t, auth.OAuth2Callback)
}

func testOAuth2Authenticated(t *testing.T, newHandler func(db database.UserDatabase) echo.HandlerFunc) {
	r := httptest.NewRequest(http.MethodGet, authURL, nil)
	w := httptest.NewRecorder()

	store := newStore()
	gothic.Store = store

	fauxSession := faux.Session{}
	s, _ := store.Get(r, fauxSessionName)
	s.Values[fauxSessionKey] = fauxSession.Marshal()
	if err := s.Save(r, w); err != nil {
		t.Error(err)
	}

	e := echo.New()
	c := e.NewContext(r, w)

	authHandler := newHandler(newDB(t))
	withSession := session.Middleware(store)(authHandler)

	if err := withSession(c); err != nil {
		t.Error(err)
	}
}

func TestAccessControl(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, authURL, nil)
	w := httptest.NewRecorder()

	store := newStore()

	e := echo.New()
	c := e.NewContext(r, w)

	m := auth.AccessControl()
	protected := session.Middleware(store)(m(func(c echo.Context) error {
		return c.String(http.StatusOK, "protected")
	}))

	// User is not logged in.
	if err := protected(c); err != echo.ErrUnauthorized {
		t.Error(err)
	}

	if err := store.login(c); err != nil {
		t.Error(err)
	}

	// User is logged in.
	if err := protected(c); err != nil {
		t.Error(err)
	}
}

func newDB(t *testing.T) database.UserDatabase {
	db, err := database.NewStructDB("", false, log.NewNopLogger())
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func assertCode(t *testing.T, haveCode, wantCode int) {
	if haveCode != wantCode {
		t.Errorf("have status code %d want %d", haveCode, wantCode)
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

func (ts testStore) login(c echo.Context) error {
	s, err := ts.Get(c.Request(), "session")
	if err != nil {
		return err
	}
	s.Values["userid"] = 0
	return s.Save(c.Request(), c.Response())
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
