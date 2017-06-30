package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/web"
	"github.com/autograde/aguis/web/auth"
	githubHandlers "github.com/autograde/aguis/web/github"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/bitbucket"
	"github.com/markbates/goth/providers/faux"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
)

func main() {
	var (
		httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")
		public   = flag.String("http.public", "public", "directory to server static files from")

		baseURL = flag.String("service.url", "localhost", "service base url")
	)
	flag.Parse()

	e := echo.New()
	e.Logger.SetLevel(log.DEBUG)

	entryPoint := filepath.Join(*public, "index.html")
	if !fileExists(entryPoint) {
		e.Logger.Warnj(log.JSON{
			"path": entryPoint,
			"err":  "could not find file",
		})
	}

	store := sessions.NewCookieStore([]byte("secret"))
	store.Options.HttpOnly = true
	store.Options.Secure = true
	gothic.Store = store

	// TODO: Only register if env set.
	goth.UseProviders(
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), getCallbackURL(*baseURL, "github"), "user"),
		bitbucket.New(os.Getenv("BITBUCKET_KEY"), os.Getenv("BITBUCKET_SECRET"), getCallbackURL(*baseURL, "bitbucket")),
		gitlab.New(os.Getenv("GITLAB_KEY"), os.Getenv("GITLAB_SECRET"), getCallbackURL(*baseURL, "gitlab")),
	)
	if _, err := goth.GetProvider((&faux.Provider{}).Name()); err == nil {
		log.Fatal("faux provider enabled in production")
	}

	e.HideBanner = true
	e.Use(
		middleware.Logger(),
		middleware.Recover(),
		middleware.Secure(),
		session.Middleware(store),
	)

	db, err := database.NewStructDB(tempFile("agdb.db"), false, e.Logger)

	if err != nil {
		log.Fatalj(log.JSON{
			"message": "could not connect to db",
			"err":     err,
		})
	}

	e.GET("/logout", func(c echo.Context) error {
		return c.Redirect(http.StatusTemporaryRedirect, "/auth/github/logout")
	})

	oauth2 := e.Group("/auth/:provider", withProvider)
	oauth2.GET("", auth.OAuth2Login(db))
	oauth2.GET("/callback", auth.OAuth2Callback(db))
	oauth2.GET("/logout", auth.OAuth2Logout())

	api := e.Group("/api/v1")
	api.Use(auth.AccessControl(db))

	api.GET("/courses", web.ListCourses(db))
	api.POST("/courses", web.NewCourse(db))

	githubAPI := api.Group("/github")
	githubAPI.GET("/organizations", githubHandlers.ListOrganizations())

	index := func(c echo.Context) error {
		return c.File(entryPoint)
	}
	e.GET("/app", index)
	e.GET("/app/*", index)

	// TODO: Whitelisted files only.
	e.Static("/", *public)

	go func() {
		if err := e.Start(*httpAddr); err == http.ErrServerClosed {
			e.Logger.Warn("shutting down the server")
			return
		}
		e.Logger.Fatalj(log.JSON{
			"message": "could not start server",
			"err":     err,
		})
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatalj(log.JSON{
			"message": "failure during server shutdown",
			"err":     err,
		})
	}
}

// makes the oauth2 provider available in the request query so that
// markbates/goth/gothic.GetProviderName can find it.
func withProvider(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		qv := c.Request().URL.Query()
		qv.Set("provider", c.Param("provider"))
		c.Request().URL.RawQuery = qv.Encode()
		return next(c)
	}
}

func getCallbackURL(baseURL string, provider string) string {
	return "https://" + baseURL + "/auth/" + provider + "/callback"
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}

func tempFile(name string) string {
	return filepath.Join(os.TempDir(), name)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
