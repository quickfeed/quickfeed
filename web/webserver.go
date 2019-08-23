package web

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/web/auth"
	"github.com/gorilla/sessions"
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

// New starts a new web server
func New(ags *AutograderService, public, httpAddr, scriptPath string, fake bool) {
	entryPoint := filepath.Join(public, "index.html")
	if _, err := os.Stat(entryPoint); os.IsNotExist(err) {
		ags.logger.Fatalf("file note found %s", entryPoint)
	}

	store := newStore([]byte("secret"))
	gothic.Store = store
	e := newServer(ags.logger, store)

	enabled := enableProviders(ags.logger, ags.bh.BaseURL, fake)
	registerWebhooks(ags, e, enabled, scriptPath)
	registerAuth(ags, e)

	registerFrontend(e, entryPoint, public)
	runWebServer(ags.logger, e, httpAddr)
}

func newServer(l *zap.SugaredLogger, store sessions.Store) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(
		middleware.Recover(),
		Logger(l.Desugar()),
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

func enableProviders(l *zap.SugaredLogger, baseURL string, fake bool) map[string]bool {
	enabled := make(map[string]bool)

	if ok := auth.EnableProvider(&auth.Provider{
		Name:          "github",
		KeyEnv:        "GITHUB_KEY",
		SecretEnv:     "GITHUB_SECRET",
		CallbackURL:   auth.GetCallbackURL(baseURL, "github"),
		StudentScopes: []string{},
		TeacherScopes: []string{"user", "repo", "delete_repo", "admin:org", "admin:org_hook"},
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

func registerWebhooks(ags *AutograderService, e *echo.Echo, enabled map[string]bool, scriptPath string) {
	webhooks.DefaultLog = WebhookLogger{SugaredLogger: ags.logger}

	docker := ci.Docker{
		Endpoint: envString("DOCKER_HOST", "http://localhost:4243"),
		Version:  envString("DOCKER_VERSION", "1.30"),
	}

	ghHook := whgithub.New(&whgithub.Config{Secret: ags.bh.Secret})
	if enabled["github"] {
		ghHook.RegisterEvents(GithubHook(ags.logger, ags.db, &docker, scriptPath), whgithub.PushEvent)
	}
	glHook := whgitlab.New(&whgitlab.Config{Secret: ags.bh.Secret})
	if enabled["gitlab"] {
		glHook.RegisterEvents(GitlabHook(ags.logger), whgitlab.PushEvents)
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

func registerAuth(ags *AutograderService, e *echo.Echo) {
	logger := ags.logger.Desugar()
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

	oauth2 := e.Group("/auth/:provider", withProvider, auth.PreAuth(logger, ags.db))
	oauth2.GET("", auth.OAuth2Login(logger, ags.db))
	oauth2.GET("/callback", auth.OAuth2Callback(logger, ags.db))
	e.GET("/logout", auth.OAuth2Logout(logger))

	api := e.Group("/api/v1")
	api.Use(auth.AccessControl(logger, ags.db, ags.scms))
	api.GET("/user", GetSelf(ags.db))
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

func runWebServer(l *zap.SugaredLogger, e *echo.Echo, httpAddr string) {
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
