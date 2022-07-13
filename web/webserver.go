package web

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/gorilla/sessions"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/rand"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/hooks"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// timeouts for http server
var (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 5 * time.Minute
)

type GrpcMultiplexer struct {
	*grpcweb.WrappedGrpcServer
}

// ServerWithCredentials starts a new gRPC server with credentials
// generated from TLS certificates.
func ServerWithCredentials(logger *zap.Logger, certFile, certKey string) (*grpc.Server, error) {
	// Generate TLS credentials from certificates

	cred, err := credentials.NewServerTLSFromFile(certFile, certKey)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer(
		grpc.Creds(cred),
		grpc.ChainUnaryInterceptor(
			auth.UserVerifier(),
			qf.Interceptor(logger),
		),
	)
	return s, nil
}

// MuxHandler routes HTTP and gRPC requests.
func (m *GrpcMultiplexer) MuxHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.IsGrpcWebRequest(r) {
			log.Printf("MUX: got GRPC request: %v", r)
			m.ServeHTTP(w, r)
			return
		}
		// log.Printf("MUX: got HTTP request: %v", r)
		next.ServeHTTP(w, r)
	})
}

// RegisterRouter registers http endpoints for authentication API and GitHub webhooks.
func RegisterRouter(logger *zap.SugaredLogger, db database.Database, authConfig *auth.AuthConfig, scms *auth.Scms, mux GrpcMultiplexer, static, secret string) *http.ServeMux {
	// Register hooks
	// TODO

	// Serve static files
	router := http.NewServeMux()
	assets := http.FileServer(http.Dir("public/assets"))
	dist := http.FileServer(http.Dir("public/dist"))

	router.Handle("/", mux.MuxHandler(http.StripPrefix("/", assets)))
	router.Handle("/assets/", mux.MuxHandler(http.StripPrefix("/assets/", assets)))
	router.Handle("/static/", mux.MuxHandler(http.StripPrefix("/static/", dist)))

	// Register auth endpoints
	router.HandleFunc("/auth/", auth.OAuth2Login(logger, db, authConfig, secret))
	router.HandleFunc("/auth/callback/", auth.OAuth2Callback(logger, db, authConfig, scms, secret))
	// logout

	return router
}

// New starts a new web server
func New(ags *QuickFeedService, public, httpAddr string) {
	entryPoint := filepath.Join(public, "assets/index.html")
	if _, err := os.Stat(entryPoint); os.IsNotExist(err) {
		ags.logger.Fatalf("file not found %s", entryPoint)
	}

	secret := rand.String()
	store := newStore([]byte(secret))
	gothic.Store = store
	e := newServer(ags, store)

	enabled := enableProviders(ags.logger, ags.bh.BaseURL)
	registerWebhooks(ags, e, enabled)
	// registerAuth(ags, e)

	registerFrontend(e, entryPoint, public)
	runWebServer(ags.logger, e, httpAddr)
}

func newServer(ags *QuickFeedService, store sessions.Store) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(
		middleware.Recover(),
		Logger(ags.logger.Desugar()),
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

func enableProviders(l *zap.SugaredLogger, baseURL string) map[string]bool {
	enabled := make(map[string]bool)

	if ok := auth.EnableProvider(&auth.Provider{
		Name:          "github",
		KeyEnv:        "GITHUB_KEY",
		SecretEnv:     "GITHUB_SECRET",
		CallbackURL:   auth.GetCallbackURL(baseURL, "github"),
		StudentScopes: []string{"repo:invite"},
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

func registerWebhooks(ags *QuickFeedService, e *echo.Echo, enabled map[string]bool) {
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

// func registerAuth(ags *QuickFeedService, e *echo.Echo) {
// 	// makes the oauth2 provider available in the request query so that
// 	// markbates/goth/gothic.GetProviderName can find it.
// 	withProvider := func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {
// 			qv := c.Request().URL.Query()
// 			qv.Set("provider", c.Param("provider"))
// 			c.Request().URL.RawQuery = qv.Encode()
// 			return next(c)
// 		}
// 	}

// 	oauth2 := e.Group("/auth/:provider", withProvider, auth.PreAuth(ags.logger, ags.db))
// 	oauth2.GET("", auth.OAuth2Login(ags.logger, ags.db))
// 	oauth2.GET("/callback", auth.OAuth2Callback(ags.logger, ags.db, ags.scms))
// 	e.GET("/logout", auth.OAuth2Logout(ags.logger))
// }

func registerFrontend(e *echo.Echo, entryPoint, public string) {
	index := func(c echo.Context) error {
		return c.File(entryPoint)
	}

	e.GET("/", index)
	e.GET("/*", index)
	// TODO: Whitelisted files only.
	e.Static("/static", filepath.Join(public, "dist"))
	e.Static("/assets", filepath.Join(public, "assets"))
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
