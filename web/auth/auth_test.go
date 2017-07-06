package auth_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"
	"github.com/autograde/aguis/web/auth"
	"github.com/gorilla/sessions"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
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

	db, cleanup := setup(t)
	defer cleanup()

	authHandler := auth.OAuth2Login(db)
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

	db, cleanup := setup(t)
	defer cleanup()

	authHandler := auth.OAuth2Callback(db)
	withSession := session.Middleware(store)(authHandler)
	err := withSession(c)
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Errorf("unexpected error type: %v", reflect.TypeOf(err))
	}

	assertCode(t, httpErr.Code, http.StatusBadRequest)
}

func TestPreAuthNoSession(t *testing.T) {
	testPreAuthLoggedIn(t, false, false, "github")
}

func TestPreAuthLoggedIn(t *testing.T) {
	testPreAuthLoggedIn(t, true, false, "github")
}

func TestPreAuthLoggedInNoDBUser(t *testing.T) {
	testPreAuthLoggedIn(t, true, true, "github")
}

func TestPreAuthLoggedInNewIdentity(t *testing.T) {
	testPreAuthLoggedIn(t, true, true, "gitlab")
}

func testPreAuthLoggedIn(t *testing.T, haveSession, existingUser bool, newProvider string) {
	const (
		provider = "github"
		remoteID = 0
		secret   = "secret"
	)
	shouldPass := !haveSession || newProvider != provider

	r := httptest.NewRequest(http.MethodGet, authURL, nil)
	w := httptest.NewRecorder()

	store := newStore()
	gothic.Store = store

	e := echo.New()
	c := e.NewContext(r, w)

	if haveSession {
		if err := store.login(c); err != nil {
			t.Error(err)
		}
	}

	db, cleanup := setup(t)
	defer cleanup()

	if existingUser {
		if _, err := db.NewUserFromRemoteIdentity(provider, remoteID, secret); err != nil {
			t.Fatal(err)
		}
		c.SetParamNames("provider")
		c.SetParamValues(newProvider)
	}

	authHandler := auth.PreAuth(db)(func(c echo.Context) error { return nil })
	withSession := session.Middleware(store)(authHandler)

	if err := withSession(c); err != nil {
		t.Error(err)
	}

	wantLocation := auth.Logout
	switch {
	case shouldPass:
		wantLocation = ""
	case existingUser:
		wantLocation = auth.Home
	}
	location := w.Header().Get("Location")
	if location != wantLocation {
		t.Errorf("have Location '%v' want '%v'", location, wantLocation)
	}

	wantCode := http.StatusFound
	if shouldPass {
		wantCode = http.StatusOK
	}

	assertCode(t, w.Code, wantCode)
}

func TestOAuth2LoginAuthenticated(t *testing.T) {
	testOAuth2Authenticated(t, auth.Home, auth.OAuth2Login)
}

func TestOAuth2CallbackAuthenticated(t *testing.T) {
	testOAuth2Authenticated(t, auth.Login, auth.OAuth2Callback)
}

func testOAuth2Authenticated(t *testing.T, wantLocation string, newHandler func(db database.Database) echo.HandlerFunc) {
	const userID = "1"

	r := httptest.NewRequest(http.MethodGet, authURL, nil)
	w := httptest.NewRecorder()

	store := newStore()
	gothic.Store = store

	fauxSession := faux.Session{ID: userID}
	s, _ := store.Get(r, fauxSessionName)
	s.Values[fauxSessionKey] = fauxSession.Marshal()
	if err := s.Save(r, w); err != nil {
		t.Error(err)
	}

	e := echo.New()
	c := e.NewContext(r, w)

	db, cleanup := setup(t)
	defer cleanup()

	authHandler := newHandler(db)
	withSession := session.Middleware(store)(authHandler)

	if err := withSession(c); err != nil {
		t.Error(err)
	}

	location := w.Header().Get("Location")
	if location != wantLocation {
		t.Errorf("have Location '%v' want '%v'", location, wantLocation)
	}
	assertCode(t, w.Code, http.StatusFound)
}

func TestAccessControl(t *testing.T) {
	const (
		provider = "github"
		remoteID = 0
		secret   = "secret"
	)

	r := httptest.NewRequest(http.MethodGet, authURL, nil)
	w := httptest.NewRecorder()

	store := newStore()

	e := echo.New()
	c := e.NewContext(r, w)

	db, cleanup := setup(t)
	defer cleanup()

	// Create a new user.
	if _, err := db.NewUserFromRemoteIdentity(provider, remoteID, secret); err != nil {
		t.Error(err)
	}

	m := auth.AccessControl(db, make(map[string]scm.SCM))
	protected := session.Middleware(store)(m(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
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

func setup(t *testing.T) (*database.GormDB, func()) {
	const (
		driver = "sqlite3"
		prefix = "testdb"
	)

	f, err := ioutil.TempFile(os.TempDir(), prefix)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	db, err := database.NewGormDB(driver, f.Name(), envSet("LOGDB"))
	if err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	return db, func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
		if err := os.Remove(f.Name()); err != nil {
			t.Error(err)
		}
	}
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
	s, err := ts.Get(c.Request(), auth.SessionKey)
	if err != nil {
		return err
	}
	s.Values[auth.UserKey] = &auth.UserSession{
		ID:        1,
		Providers: map[string]struct{}{"github": {}},
	}
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

func envSet(env string) bool {
	return os.Getenv(env) != ""
}
