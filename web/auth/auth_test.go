package auth_test

import (
	"net/http"
	"testing"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/steinfletcher/apitest"
)

const (
	testSecret     = "top-secret"
	user           = "/user"
	authGithub     = "/auth/github"
	callbackGithub = "/auth/callback/github"
	loginToken     = "/login/oauth/access_token"
)

func TestOAuth2Login(t *testing.T) {
	logger := qtest.Logger(t)
	authConfig := auth.NewGitHubConfig(&scm.Config{})
	// Incorrect request method.
	apitest.New().HandlerFunc(auth.OAuth2Login(logger, authConfig, "")).
		Post(auth.Auth).
		Expect(t).
		Status(http.StatusUnauthorized).
		End()
	// No existing auth cookie.
	apitest.New().HandlerFunc(auth.OAuth2Login(logger, authConfig, "")).
		Get(auth.Auth).
		Expect(t).
		Status(http.StatusTemporaryRedirect).
		End()
	// Outdated auth cookie with expected name should not break API.
	apitest.New().HandlerFunc(auth.OAuth2Login(logger, authConfig, "")).
		Get(auth.Auth).
		Cookie(auth.CookieName, "empty").
		Expect(t).
		Status(http.StatusTemporaryRedirect).
		End()
}

func TestOAuth2LoginRedirect(t *testing.T) {
	logger := qtest.Logger(t)
	authConfig := auth.NewGitHubConfig(&scm.Config{})

	apitest.New().HandlerFunc(auth.OAuth2Login(logger, authConfig, "")).
		Get(authGithub).
		Expect(t).
		Status(http.StatusTemporaryRedirect).
		Assert(func(res *http.Response, _ *http.Request) error {
			fullURL, err := res.Location()
			if err != nil {
				return err
			}
			redirectURL := fullURL.Path
			wantRedirectURL := "/login/oauth/authorize"
			if redirectURL != wantRedirectURL {
				t.Errorf("got redirect URL: %v, want %v", redirectURL, wantRedirectURL)
			}
			return nil
		}).
		End()
}

func TestOAuth2Callback(t *testing.T) {
	userJSON := `{"id": 1, "email": "mail", "name": "No name Last name", "login": "test"}`
	logger := qtest.Logger(t)
	authConfig := auth.NewGitHubConfig(&scm.Config{})
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}

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

	apitest.New().Mocks(mockTokenExchange, mockUserExchange).
		HandlerFunc(auth.OAuth2Callback(logger, db, tm, authConfig, testSecret)).
		Get(callbackGithub).
		Query("state", testSecret).
		Query("code", "test code").
		Expect(t).
		Status(http.StatusFound).
		HeaderPresent(auth.SetCookie).
		End()

	user, err := db.GetUser(1)
	if err != nil {
		t.Fatal(err)
	}
	if user.GetLogin() != "test" {
		t.Fatalf("incorrect user login: expected 'test', got %s", user.GetName())
	}
}

func TestOAuth2CallbackUserExchange(t *testing.T) {
	logger := qtest.Logger(t)
	authConfig := auth.NewGitHubConfig(&scm.Config{})
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}

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

	apitest.New().Mocks(mockTokenExchange, mockEmptyUserInfo).
		HandlerFunc(auth.OAuth2Callback(logger, db, tm, authConfig, testSecret)).
		Get(callbackGithub).
		Query("state", testSecret).
		Query("code", "test code").
		Expect(t).
		Status(http.StatusUnauthorized).
		HeaderNotPresent(auth.SetCookie).
		End()
	apitest.New().Mocks(mockTokenExchange, mockEmptyResponseBody).
		HandlerFunc(auth.OAuth2Callback(logger, db, tm, authConfig, testSecret)).
		Get(callbackGithub).
		Query("state", testSecret).
		Query("code", "test code").
		Expect(t).
		Status(http.StatusUnauthorized).
		HeaderNotPresent(auth.SetCookie).
		End()
	apitest.New().Mocks(mockTokenExchange, mockBadRequestStatus).
		HandlerFunc(auth.OAuth2Callback(logger, db, tm, authConfig, testSecret)).
		Get(callbackGithub).
		Query("state", testSecret).
		Query("code", "test code").
		Expect(t).
		Status(http.StatusUnauthorized).
		HeaderNotPresent(auth.SetCookie).
		End()

	checkNoUsersInDB(db, t)
}

func TestOAuth2CallbackTokenExchange(t *testing.T) {
	logger := qtest.Logger(t)
	authConfig := auth.NewGitHubConfig(&scm.Config{})
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}

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
	apitest.New().Mocks(mockEmptyAccessToken).
		HandlerFunc(auth.OAuth2Callback(logger, db, tm, authConfig, testSecret)).
		Get(callbackGithub).
		Query("state", testSecret).
		Query("code", "test code").
		Expect(t).
		Status(http.StatusUnauthorized).
		HeaderNotPresent(auth.SetCookie).
		End()
	// No values in the request body.
	apitest.New().Mocks(mockEmptyResponseBody).
		HandlerFunc(auth.OAuth2Callback(logger, db, tm, authConfig, testSecret)).
		Get(callbackGithub).
		Query("state", testSecret).
		Query("code", "test code").
		Expect(t).
		Status(http.StatusUnauthorized).
		HeaderNotPresent(auth.SetCookie).
		End()

	checkNoUsersInDB(db, t)
}

func TestOAuth2CallbackBadRequest(t *testing.T) {
	logger := qtest.Logger(t)
	authConfig := auth.NewGitHubConfig(&scm.Config{})
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}
	// Wrong request method.
	apitest.New().HandlerFunc(auth.OAuth2Callback(logger, db, tm, authConfig, testSecret)).
		Post(callbackGithub).
		Query("state", testSecret).
		Query("code", "test code").
		Expect(t).
		Status(http.StatusUnauthorized).
		End()
	// Incorrect secret code.
	apitest.New().HandlerFunc(auth.OAuth2Callback(logger, db, tm, authConfig, testSecret)).
		Get(callbackGithub).
		Query("state", "not a secret").
		Query("code", "test code").
		Expect(t).
		Status(http.StatusUnauthorized).
		End()
	// Empty exchange code.
	apitest.New().HandlerFunc(auth.OAuth2Callback(logger, db, tm, authConfig, testSecret)).
		Get(callbackGithub).
		Query("state", testSecret).
		Query("code", "").
		Expect(t).
		Status(http.StatusUnauthorized).
		End()
	// Request with empty body content.
	apitest.New().HandlerFunc(auth.OAuth2Callback(logger, db, tm, authConfig, testSecret)).
		Get(callbackGithub).
		Expect(t).
		Status(http.StatusUnauthorized).
		End()

	checkNoUsersInDB(db, t)
}

func TestOAuth2Logout(t *testing.T) {
	apitest.New().HandlerFunc(auth.OAuth2Logout()).
		Get(auth.Logout).
		// Make sure an outdated auth cookie with a correct name does not break API.
		Cookie(auth.CookieName, "empty").
		Expect(t).
		Status(http.StatusFound).
		Cookies(
			apitest.NewCookie(auth.CookieName).
				Value("").
				MaxAge(-1),
		).
		End()
}

func TestSanitizeNext(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty string", "", "/"},
		{"whitespace only", "   ", "/"},
		{"root path", "/", "/"},
		{"simple path", "/dashboard", "/dashboard"},
		{"path with trailing slash", "/dashboard/", "/dashboard"},
		{"path with dot", "/./dashboard", "/dashboard"},
		{"path with double slash", "//dashboard", "/"},
		{"path with backslash", `/dash\board`, "/"},
		{"absolute URL http", "http://example.com", "/"},
		{"absolute URL https", "https://example.com/path", "/"},
		{"absolute URL with path", "https://example.com/dashboard", "/"},
		{"relative path", "dashboard", "/"},
		{"path with ..", "/foo/../bar", "/bar"},
		{"path with multiple ..", "/foo/../../bar", "/bar"},
		{"path with spaces", "   /foo/bar   ", "/foo/bar"},
		{"path with query", "/foo/bar?baz=1", "/foo/bar?baz=1"},
		{"path with fragment", "/foo/bar#section", "/foo/bar#section"},
		{"dot only", ".", "/"},
		{"slash dot", "/.", "/"},
		{"empty after clean", "", "/"},
		{"path with encoded slash", "/foo%2Fbar", "/foo%2Fbar"},
		{"path with backslash", "/foo\\bar", "/"},
		{"protocol relative URL", "//example.com", "/"},
		{"protocol backslash URL", "\\example.com", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := auth.SanitizeNext(tt.in)
			if got != tt.want {
				t.Errorf("SanitizeNext(%q) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
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
