package auth_test

import (
	"net/http"
	"testing"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/steinfletcher/apitest"
	"gotest.tools/assert"
)

const (
	testSecret     = "top-secret"
	user           = "/user"
	authGithub     = "/auth/github"
	callbackGithub = "/auth/callback/github"
	loginToken     = "/login/oauth/access_token"
)

func TestOAuth2Login(t *testing.T) {
	logger := qlog.Logger(t)
	authConfig := auth.NewGitHubConfig("", "TEST", "TEST")
	// Incorrect request method.
	apitest.New().Debug().
		HandlerFunc(auth.OAuth2Login(logger, authConfig, "")).
		Post(web.Auth).
		Expect(t).
		Status(http.StatusUnauthorized).
		End()
	// No existing session cookie.
	apitest.New().Debug().
		HandlerFunc(auth.OAuth2Login(logger, authConfig, "")).
		Get(web.Auth).
		Expect(t).
		Status(http.StatusTemporaryRedirect).
		End()
	// Outdated session with expected name should not break API.
	apitest.New().Debug().
		HandlerFunc(auth.OAuth2Login(logger, authConfig, "")).
		Get(web.Auth).
		Cookie("session", "empty").
		Expect(t).
		Status(http.StatusTemporaryRedirect).
		End()
}

func TestOAuth2LoginRedirect(t *testing.T) {
	logger := qlog.Logger(t)
	authConfig := auth.NewGitHubConfig("", "TEST", "TEST")

	apitest.New().Debug().
		HandlerFunc(auth.OAuth2Login(logger, authConfig, "")).
		Get(authGithub).
		Expect(t).
		Status(http.StatusTemporaryRedirect).
		Assert(func(res *http.Response, _ *http.Request) error {
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
	userJSON := `{"id": 1, "email": "mail", "name": "No name Last name", "login": "test"}`
	logger := qtest.Logger(t)
	authConfig := auth.NewGitHubConfig("", "TEST", "TEST")
	scms := scm.NewScms()
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	mockTokenExchange := apitest.NewMock().
		Post(loginToken).
		RespondWith().
		Body(`{"access_token": "test_token"}`).
		Status(http.StatusOK).
		End()
	mockUserExchange := apitest.NewMock().
		Get(user).
		RespondWith().
		Body(userJSON).
		Status(http.StatusOK).
		End()

	apitest.New().Debug().
		Mocks(mockTokenExchange, mockUserExchange).
		HandlerFunc(auth.OAuth2Callback(logger, db, authConfig, scms, testSecret)).
		Get(callbackGithub).
		Query("state", testSecret).
		Query("code", "test code").
		Expect(t).
		Status(http.StatusFound).
		HeaderPresent("Set-Cookie").
		End()

	user, err := db.GetUser(1)
	if err != nil {
		t.Fatal(err)
	}
	if user.Login != "test" {
		t.Fatalf("incorrect user login: expected 'test', got %s", user.Name)
	}
}

func TestOAuth2CallbackUserExchange(t *testing.T) {
	logger := qtest.Logger(t)
	authConfig := auth.NewGitHubConfig("", "TEST", "TEST")
	scms := scm.NewScms()
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	mockTokenExchange := apitest.NewMock().
		Post(loginToken).
		RespondWith().
		Body(`{"access_token": "test_token"}`).
		Status(http.StatusOK).
		End()
	mockEmptyUserInfo := apitest.NewMock().
		Get(user).
		RespondWith().
		Body(`userID: "none"`).
		Status(http.StatusOK).
		End()
	mockEmptyResponseBody := apitest.NewMock().
		Get(user).
		RespondWith().
		Status(http.StatusOK).
		End()
	mockBadRequestStatus := apitest.NewMock().
		Get(user).
		RespondWith().
		Body(`userID: "none"`).
		Status(http.StatusBadRequest).
		End()

	apitest.New().Debug().
		Mocks(mockTokenExchange, mockEmptyUserInfo).
		HandlerFunc(auth.OAuth2Callback(logger, db, authConfig, scms, testSecret)).
		Get(callbackGithub).
		Query("state", testSecret).
		Query("code", "test code").
		Expect(t).
		Status(http.StatusUnauthorized).
		HeaderNotPresent("Set-Cookie").
		End()
	apitest.New().Debug().
		Mocks(mockTokenExchange, mockEmptyResponseBody).
		HandlerFunc(auth.OAuth2Callback(logger, db, authConfig, scms, testSecret)).
		Get(callbackGithub).
		Query("state", testSecret).
		Query("code", "test code").
		Expect(t).
		Status(http.StatusUnauthorized).
		HeaderNotPresent("Set-Cookie").
		End()
	apitest.New().Debug().
		Mocks(mockTokenExchange, mockBadRequestStatus).
		HandlerFunc(auth.OAuth2Callback(logger, db, authConfig, scms, testSecret)).
		Get(callbackGithub).
		Query("state", testSecret).
		Query("code", "test code").
		Expect(t).
		Status(http.StatusUnauthorized).
		HeaderNotPresent("Set-Cookie").
		End()

	checkNoUsersInDB(db, t)
}

func TestOAuth2CallbackTokenExchange(t *testing.T) {
	logger := qtest.Logger(t)
	authConfig := auth.NewGitHubConfig("", "TEST", "TEST")
	scms := scm.NewScms()
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	mockEmptyAccessToken := apitest.NewMock().
		Post(loginToken).
		RespondWith().
		Body(`{"access_token": ""}`).
		Status(http.StatusOK).
		End()
	mockEmptyResponseBody := apitest.NewMock().
		Post(loginToken).
		RespondWith().
		Status(http.StatusOK).
		End()
	// Token value is an empty string.
	apitest.New().Debug().
		Mocks(mockEmptyAccessToken).
		HandlerFunc(auth.OAuth2Callback(logger, db, authConfig, scms, testSecret)).
		Get(callbackGithub).
		Query("state", testSecret).
		Query("code", "test code").
		Expect(t).
		Status(http.StatusUnauthorized).
		HeaderNotPresent("Set-Cookie").
		End()
	// No values in the request body.
	apitest.New().Debug().
		Mocks(mockEmptyResponseBody).
		HandlerFunc(auth.OAuth2Callback(logger, db, authConfig, scms, testSecret)).
		Get(callbackGithub).
		Query("state", testSecret).
		Query("code", "test code").
		Expect(t).
		Status(http.StatusUnauthorized).
		HeaderNotPresent("Set-Cookie").
		End()

	checkNoUsersInDB(db, t)
}

func TestOAuth2CallbackBadRequest(t *testing.T) {
	logger := qtest.Logger(t)
	authConfig := auth.NewGitHubConfig("", "TEST", "TEST")
	scms := scm.NewScms()
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	// Wrong request method.
	apitest.New().Debug().
		HandlerFunc(auth.OAuth2Callback(logger, db, authConfig, scms, testSecret)).
		Post(callbackGithub).
		Query("state", testSecret).
		Query("code", "test code").
		Expect(t).
		Status(http.StatusUnauthorized).
		End()
	// Incorrect secret code.
	apitest.New().Debug().
		HandlerFunc(auth.OAuth2Callback(logger, db, authConfig, scms, testSecret)).
		Get(callbackGithub).
		Query("state", "not a secret").
		Query("code", "test code").
		Expect(t).
		Status(http.StatusUnauthorized).
		End()
	// Empty exchange code.
	apitest.New().Debug().
		HandlerFunc(auth.OAuth2Callback(logger, db, authConfig, scms, testSecret)).
		Get(callbackGithub).
		Query("state", testSecret).
		Query("code", "").
		Expect(t).
		Status(http.StatusUnauthorized).
		End()
	// Request with empty body content.
	apitest.New().Debug().
		HandlerFunc(auth.OAuth2Callback(logger, db, authConfig, scms, testSecret)).
		Get(callbackGithub).
		Expect(t).
		Status(http.StatusUnauthorized).
		End()

	checkNoUsersInDB(db, t)
}

func TestOAuth2Logout(t *testing.T) {
	apitest.New().Debug().
		HandlerFunc(auth.OAuth2Logout(qlog.Logger(t))).
		Get(web.Logout).
		// Make sure an outdated session with a correct name does not break API.
		Cookie("session", "empty").
		Expect(t).
		Status(http.StatusFound).
		Cookies(
			apitest.NewCookie("session").MaxAge(-1),
		).
		End()
}

func checkNoUsersInDB(db database.Database, t *testing.T) {
	users, err := db.GetUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(users) > 0 {
		t.Fatalf("Expected empty database, got %d users", len(users))
	}
}
