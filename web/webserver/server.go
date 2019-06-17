package web

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/web"
	"github.com/autograde/aguis/web/auth"
	"github.com/gorilla/sessions"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
	"go.uber.org/zap"

	webhooks "gopkg.in/go-playground/webhooks.v3"
	whgithub "gopkg.in/go-playground/webhooks.v3/github"
	whgitlab "gopkg.in/go-playground/webhooks.v3/gitlab"
)

// NewWebServer starts a new web server
func NewWebServer(db *database.GormDB, bh web.BaseHookOptions, l *zap.Logger, public, httpAddr string, baseURL string, fake bool, buildscripts string, scms *web.Scms) {
	entryPoint := filepath.Join(public, "index.html")
	if _, err := os.Stat(entryPoint); os.IsNotExist(err) {
		l.Fatal("file not found", zap.String("path", entryPoint))
	}

	store := newStore([]byte("secret"))
	gothic.Store = store
	e := newServer(l, store)

	enabled := enableProviders(l, baseURL, fake)
	registerWebhooks(l, e, db, bh.Secret, enabled, &buildscripts)
	registerAuth(l, e, db, scms)

	registerFrontend(e, entryPoint, public)
	runWebServer(l, e, httpAddr)
}

func newServer(l *zap.Logger, store sessions.Store) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(
		middleware.Recover(),
		web.Logger(l),
		middleware.Secure(),
		session.Middleware(store),
	)
	return e
}

func newStore(keyPairs ...[]byte) sessions.Store {
	store := sessions.NewCookieStore(keyPairs...)
	store.Options.HttpOnly = true
	store.Options.Secure = true
	return store
}

func enableProviders(l *zap.Logger, baseURL string, fake bool) map[string]bool {
	enabled := make(map[string]bool)

	if ok := auth.EnableProvider(&auth.Provider{
		Name:          "github",
		KeyEnv:        "GITHUB_KEY",
		SecretEnv:     "GITHUB_SECRET",
		CallbackURL:   auth.GetCallbackURL(baseURL, "github"),
		StudentScopes: []string{},
		TeacherScopes: []string{"user", "repo", "delete_repo", "admin:org"},
	}, func(key, secret, callback string, scopes ...string) goth.Provider {
		return github.New(key, secret, callback, scopes...)
	}); ok {
		enabled["github"] = true
	} else {
		l.Debug("environment variable not set for github")
	}

	if ok := auth.EnableProvider(&auth.Provider{
		Name:          "gitlab",
		KeyEnv:        "GITLAB_KEY",
		SecretEnv:     "GITLAB_SECRET",
		CallbackURL:   auth.GetCallbackURL(baseURL, "gitlab"),
		StudentScopes: []string{"read_user"},
		TeacherScopes: []string{"api"},
	}, func(key, secret, callback string, scopes ...string) goth.Provider {
		return gitlab.New(key, secret, callback, scopes...)
	}); ok {
		enabled["gitlab"] = true
	} else {
		l.Debug("environment variable not set for gitlab")
	}

	if fake {
		l.Debug("fake provider enabled")
		goth.UseProviders(&auth.FakeProvider{
			Callback: auth.GetCallbackURL(baseURL, "fake"),
		})
		goth.UseProviders(&auth.FakeProvider{
			Callback: auth.GetCallbackURL(baseURL, "fake-teacher"),
		})
	}

	return enabled
}

func registerWebhooks(logger *zap.Logger, e *echo.Echo, db database.Database, secret string, enabled map[string]bool, buildscripts *string) {
	webhooks.DefaultLog = web.WebhookLogger{Logger: logger}

	docker := ci.Docker{
		Endpoint: envString("DOCKER_HOST", "http://localhost:4243"),
		Version:  envString("DOCKER_VERSION", "1.30"),
	}

	ghHook := whgithub.New(&whgithub.Config{Secret: secret})
	if enabled["github"] {
		ghHook.RegisterEvents(web.GithubHook(logger, db, &docker, *buildscripts), whgithub.PushEvent)
	}
	glHook := whgitlab.New(&whgitlab.Config{Secret: secret})
	if enabled["gitlab"] {
		glHook.RegisterEvents(web.GitlabHook(logger), whgitlab.PushEvents)
	}

	e.POST("/hook/:provider/events", func(c echo.Context) error {
		var hook webhooks.Webhook
		provider := c.Param("provider")
		if !enabled[provider] {
			return echo.ErrNotFound
		}

		switch provider {
		case "github":
			hook = ghHook
		case "gitlab":
			hook = glHook
		default:
			panic("registered provider is missing corresponding webhook")
		}
		webhooks.Handler(hook).ServeHTTP(c.Response(), c.Request())
		return nil
	})
}

func registerAuth(logger *zap.Logger, e *echo.Echo, db database.Database, scms *web.Scms) {
	// makes the oauth2 provider available in the request query so that
	// markbates/goth/gothic.GetProviderName can find it.
	withProvider := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			qv := c.Request().URL.Query()
			qv.Set("provider", c.Param("provider"))
			c.Request().URL.RawQuery = qv.Encode()
			return next(c)
		}
	}

	oauth2 := e.Group("/auth/:provider", withProvider, auth.PreAuth(logger, db))
	oauth2.GET("", auth.OAuth2Login(logger, db))
	oauth2.GET("/callback", auth.OAuth2Callback(logger, db))
	e.GET("/logout", auth.OAuth2Logout(logger))

	api := e.Group("/api/v1")
	api.Use(auth.AccessControl(logger, db, scms))
	api.GET("/user", web.GetSelf(db))
}

func registerFrontend(e *echo.Echo, entryPoint, public string) {
	index := func(c echo.Context) error {
		return c.File(entryPoint)
	}
	e.GET("/app", index)
	e.GET("/app/*", index)

	// TODO: Whitelisted files only.
	e.Static("/", public)
}

func runWebServer(l *zap.Logger, e *echo.Echo, httpAddr string) {
	srvErr := e.Start(httpAddr)
	if srvErr == http.ErrServerClosed {
		l.Warn("shutting down the server")
		return
	}
	l.Fatal("failed to start server", zap.Error(srvErr))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		l.Fatal("failure during server shutdown", zap.Error(err))
	}
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
