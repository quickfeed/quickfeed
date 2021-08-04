package web

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/autograde/quickfeed/web/auth"
	"github.com/autograde/quickfeed/web/hooks"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
	"go.uber.org/zap"
)

// timeouts for http server
var (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 5 * time.Minute
)

// New starts a new web server
func New(ags *AutograderService, public, httpAddr string) {
	entryPoint := filepath.Join(public, "index.html")
	if _, err := os.Stat(entryPoint); os.IsNotExist(err) {
		ags.logger.Fatalf("file not found %s", entryPoint)
	}

	store := newStore([]byte("secret"))
	gothic.Store = store
	e := newServer(ags, store)

	enabled := enableProviders(ags.logger, ags.bh.BaseURL)
	registerWebhooks(ags, e, enabled)
	registerAuth(ags, e)

	registerFrontend(e, entryPoint, public)
	runWebServer(ags.logger, e, httpAddr)
}

func newServer(ags *AutograderService, store sessions.Store) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(
		middleware.Recover(),
		Logger(ags.logger.Desugar()),
		middleware.Secure(),
		session.Middleware(store),
		auth.AccessControl(ags.logger.Desugar(), ags.db, ags.scms),
	)
	return e
}

func newStore(keyPairs ...[]byte) sessions.Store {
	store := sessions.NewCookieStore(keyPairs...)
	store.Options.HttpOnly = true
	store.Options.Secure = true
	return store
}

func enableProviders(l *zap.SugaredLogger, baseURL string) map[string]bool {
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

	return enabled
}

func registerWebhooks(ags *AutograderService, e *echo.Echo, enabled map[string]bool) {
	if enabled["github"] {
		ghHook := hooks.NewGitHubWebHook(ags.logger, ags.db, ags.runner, ags.bh.Secret)
		e.POST("/hook/github/events", func(c echo.Context) error {
			ghHook.Handle(c.Response(), c.Request())
			return nil
		})
	}
	if enabled["gitlab"] {
		// TODO(meling) fix gitlab
		glHook := hooks.NewGitHubWebHook(ags.logger, ags.db, ags.runner, ags.bh.Secret)
		e.POST("/hook/gitlab/events", func(c echo.Context) error {
			glHook.Handle(c.Response(), c.Request())
			return nil
		})
	}
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
	e.Server.WriteTimeout = writeTimeout
	e.Server.ReadTimeout = readTimeout
	e.Server.IdleTimeout = idleTimeout

	e.TLSServer.ReadTimeout = readTimeout
	e.TLSServer.WriteTimeout = writeTimeout
	e.TLSServer.IdleTimeout = idleTimeout

	srvErr := e.Start(httpAddr)
	if srvErr == http.ErrServerClosed {
		l.Warn("shutting down the server")
		return
	}
	l.Fatal("failed to start server", zap.Error(srvErr))
	// TODO(meling) pretty sure the following lines are unreachable; check

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		l.Fatal("failure during server shutdown", zap.Error(err))
	}
}
