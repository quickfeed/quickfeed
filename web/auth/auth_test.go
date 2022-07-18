package auth_test

import (
	"net/http"
	"testing"

	"github.com/markbates/goth"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/steinfletcher/apitest"
	"gotest.tools/assert"
)

const (
	testSecret = "topsecret"
)

func init() {
	goth.UseProviders(&auth.FakeProvider{
		Callback: "/auth/callback/fake",
	})
}

func TestOAuth2Login(t *testing.T) {
	tests := map[string]int{
		"/auth/":               http.StatusUnauthorized,
		"/auth":                http.StatusUnauthorized,
		"/auth/wrong-provider": http.StatusUnauthorized,
		"/github":              http.StatusUnauthorized,
		"/auth/github":         http.StatusTemporaryRedirect,
		"/auth/github/":        http.StatusTemporaryRedirect,
	}
	logger := qlog.Logger(t)
	authConfig, err := auth.NewConfig("")
	if err != nil {
		t.Fatal(err)
	}

	for requestURL, status := range tests {
		apitest.New().Debug().
			HandlerFunc(auth.OAuth2Login(logger, authConfig, "")).
			Get(requestURL).
			// Make sure an outdated session with a correct name does not break API.
			Cookie("session", "empty").
			Expect(t).
			Status(status).
			HeaderPresent("Set-Cookie").
			End()
	}
}

func TestOAuth2LoginRedirect(t *testing.T) {
	logger := qlog.Logger(t)
	authConfig, err := auth.NewConfig("")
	if err != nil {
		t.Fatal(err)
	}

	apitest.New().Debug().
		HandlerFunc(auth.OAuth2Login(logger, authConfig, "")).
		Get("/auth/github").
		Expect(t).
		Status(http.StatusTemporaryRedirect).
		Assert(func(res *http.Response, req *http.Request) error {
			fullURL, err := res.Location()
			if err != nil {
				return err
			}
			redirectURL := fullURL.Path
			assert.Equal(t, redirectURL, "/login/oauth/authorize")
			return nil
		}).
		End()
}

func TestOAuth2Callback(t *testing.T) {
	userJSON := `{"id": 1, "email": "mail", "name": "Noname Lastname", "login": "test"}`
	logger := qtest.Logger(t)
	authConfig, err := auth.NewConfig("")
	if err != nil {
		t.Fatal(err)
	}
	scms := auth.NewScms()
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	mockTokenExchange := apitest.NewMock().
		Post("/login/oauth/access_token").
		RespondWith().
		Body(`{"access_token": "test_token"}`).
		Status(http.StatusOK).
		End()
	mockUserExchange := apitest.NewMock().
		Get("/user").
		RespondWith().
		Body(userJSON).
		Status(http.StatusOK).
		End()

	apitest.New().Debug().
		Mocks(mockTokenExchange, mockUserExchange).
		HandlerFunc(auth.OAuth2Callback(logger, db, authConfig, scms, testSecret)).
		Get("/auth/callback/github").
		Query("state", testSecret).
		Query("code", "testcode").
		Expect(t).
		Status(http.StatusFound).
		End()

	user, err := db.GetUser(1)
	if err != nil {
		t.Fatal(err)
	}
	if user.Login != "test" {
		t.Fatalf("incorrect user login: expected 'test', got %s", user.Name)
	}
}

func TestOAuth2CallbackBadRequest(t *testing.T) {
	logger := qtest.Logger(t)
	authConfig, err := auth.NewConfig("")
	if err != nil {
		t.Fatal(err)
	}
	scms := auth.NewScms()
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	// Wrong request method.
	apitest.New().Debug().
		HandlerFunc(auth.OAuth2Callback(logger, db, authConfig, scms, testSecret)).
		Post("/auth/callback/github").
		Query("state", testSecret).
		Query("code", "testcode").
		Expect(t).
		Status(http.StatusUnauthorized).
		End()
	// Incorrect secret code.
	apitest.New().Debug().
		HandlerFunc(auth.OAuth2Callback(logger, db, authConfig, scms, testSecret)).
		Get("/auth/callback/github").
		Query("state", "not a secret").
		Query("code", "testcode").
		Expect(t).
		Status(http.StatusUnauthorized).
		End()
	// Empty exchange code.
	apitest.New().Debug().
		HandlerFunc(auth.OAuth2Callback(logger, db, authConfig, scms, testSecret)).
		Get("/auth/callback/github").
		Query("state", testSecret).
		Query("code", "").
		Expect(t).
		Status(http.StatusUnauthorized).
		End()
}

func TestOAuth2Logout(t *testing.T) {
	apitest.New().Debug().
		HandlerFunc(auth.OAuth2Logout(qlog.Logger(t))).
		Get("/logout").
		// Make sure an outdated session with a correct name does not break API.
		Cookie("session", "empty").
		Expect(t).
		Status(http.StatusFound).
		Cookies(
			apitest.NewCookie("session").MaxAge(-1),
		).
		End()
}
