package web_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/web"
	"github.com/autograde/aguis/web/auth"
	"github.com/labstack/echo"
)

const (
	userURL   = "/user"
	usersURL  = "/users"
	user1URL  = "/users/1"
	apiPrefix = "/api/v1"
)

func TestGetSelf(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, userURL, nil)
	w := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(r, w)
	c.Set(auth.UserKey, &models.User{ID: 1})

	userHandler := web.GetSelf()
	if err := userHandler(c); err != nil {
		t.Error(err)
	}

	location := w.Header().Get("Location")
	if location != apiPrefix+user1URL {
		t.Errorf("have Location '%v' want '%v'", location, apiPrefix+user1URL)
	}
	assertCode(t, w.Code, http.StatusFound)
}

func TestGetUser(t *testing.T) {
	const (
		rID1      = 1
		secret1   = "123"
		provider1 = "github"
		remoteID1 = 10
	)

	db, cleanup := setup(t)
	defer cleanup()

	dbuser, err := db.NewUserFromRemoteIdentity(provider1, rID1, secret1)
	if err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodGet, "/users/:id", web.GetUser(db))

	r := httptest.NewRequest(http.MethodGet, user1URL, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	// Prepare context with user request.
	router.Find(http.MethodGet, user1URL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}

	var foundUser *models.User
	if err := json.Unmarshal(w.Body.Bytes(), &foundUser); err != nil {
		t.Fatal(err)
	}

	// Access token should be stripped.
	dbuser.RemoteIdentities[0].AccessToken = ""
	if !reflect.DeepEqual(foundUser, dbuser) {
		t.Errorf("have user %+v want %+v", foundUser, dbuser)
	}
	assertCode(t, w.Code, http.StatusFound)
}

func TestGetUsers(t *testing.T) {
	const (
		rID1      = 1
		secret1   = "123"
		provider1 = "github"
		remoteID1 = 10

		rID2      = 2
		secret2   = "456"
		provider2 = "gitlab"
		remoteID2 = 20
	)

	db, cleanup := setup(t)
	defer cleanup()

	user1, err := db.NewUserFromRemoteIdentity(provider1, rID1, secret1)
	if err != nil {
		t.Fatal(err)
	}
	user2, err := db.NewUserFromRemoteIdentity(provider2, rID2, secret2)
	if err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	r := httptest.NewRequest(http.MethodGet, usersURL, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)

	h := web.GetUsers(db)
	if err := h(c); err != nil {
		t.Error(err)
	}

	var foundUsers []*models.User
	if err := json.Unmarshal(w.Body.Bytes(), &foundUsers); err != nil {
		t.Fatal(err)
	}

	// Remote identities should not be loaded.
	user1.RemoteIdentities = nil
	user2.RemoteIdentities = nil
	wantUsers := []*models.User{user1, user2}
	if !reflect.DeepEqual(foundUsers, wantUsers) {
		t.Errorf("have users %+v want %+v", foundUsers, wantUsers)
	}

	assertCode(t, w.Code, http.StatusFound)
}
