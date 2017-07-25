package main

import (
	"context"
	"flag"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/logger"
	"github.com/autograde/aguis/scm"
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

	webhooks "gopkg.in/go-playground/webhooks.v3"
	whgithub "gopkg.in/go-playground/webhooks.v3/github"
	whgitlab "gopkg.in/go-playground/webhooks.v3/gitlab"
)

func main() {
	var (
		httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")
		public   = flag.String("http.public", "public", "directory to server static files from")

		baseURL = flag.String("service.url", "localhost", "service base url")

		fake = flag.Bool("provider.fake", false, "enable fake provider")
	)
	flag.Parse()

	setDefaultMimeTypes()

	e := echo.New()
	l := logrus.New()
	l.Formatter = logger.NewDevFormatter(l.Formatter)
	e.Logger = web.EchoLogger{Logger: l}

	entryPoint := filepath.Join(*public, "index.html")
	if !fileExists(entryPoint) {
		l.WithField("path", entryPoint).Warn("could not find file")
	}

	store := sessions.NewCookieStore([]byte("secret"))
	store.Options.HttpOnly = true
	store.Options.Secure = true
	gothic.Store = store

	if ok := auth.EnableProvider(&auth.Provider{
		Name:          "github",
		KeyEnv:        "GITHUB_KEY",
		SecretEnv:     "GITHUB_SECRET",
		CallbackURL:   getCallbackURL(*baseURL, "github"),
		StudentScopes: []string{},
		TeacherScopes: []string{"user", "repo"},
	}, func(key, secret, callback string, scopes ...string) goth.Provider {
		return github.New(key, secret, callback, scopes...)
	}); !ok {
		l.WithFields(logrus.Fields{
			"provider": "github",
			"enabled":  false,
		}).Warn("environment variables not set")
	}
	if ok := auth.EnableProvider(&auth.Provider{
		Name:          "gitlab",
		KeyEnv:        "GITLAB_KEY",
		SecretEnv:     "GITLAB_SECRET",
		CallbackURL:   getCallbackURL(*baseURL, "gitlab"),
		StudentScopes: []string{"read_user"},
		TeacherScopes: []string{"api"},
	}, func(key, secret, callback string, scopes ...string) goth.Provider {
		return gitlab.New(key, secret, callback, scopes...)
	}); !ok {
		l.WithFields(logrus.Fields{
			"provider": "gitlab",
			"enabled":  false,
		}).Warn("environment variables not set")
	}

	if *fake {
		l.Warn("fake provider enabled")
		goth.UseProviders(&auth.FakeProvider{Callback: getCallbackURL(*baseURL, "fake")})
	}

	e.HideBanner = true
	e.Use(
		middleware.Recover(),
		web.Logger(l),
		middleware.Secure(),
		session.Middleware(store),
	)

	db, err := database.NewGormDB("sqlite3", tempFile("agdb.db"), database.Logger{Logger: l})
	defer db.Close()

	if err != nil {
		l.WithError(err).Fatal("could not connect to db")
	}

	e.GET("/logout", auth.OAuth2Logout())

	ghHook := whgithub.New(&whgithub.Config{Secret: os.Getenv("GITHUB_HOOK_SECRET")})
	ghHook.RegisterEvents(web.GithubHook, whgithub.PushEvent)

	glHook := whgitlab.New(&whgitlab.Config{Secret: os.Getenv("GITLAB_HOOK_SECRET")})
	glHook.RegisterEvents(web.GitlabHook, whgitlab.PushEvents)

	e.POST("/hook/:provider/events", func(c echo.Context) error {
		var hook webhooks.Webhook
		switch c.Param("provider") {
		case "github":
			hook = ghHook
		case "gitlab":
			hook = glHook
		default:
			return echo.ErrNotFound
		}
		webhooks.Handler(hook).ServeHTTP(c.Response(), c.Request())
		return nil
	})

	oauth2 := e.Group("/auth/:provider", withProvider, auth.PreAuth(db))
	oauth2.GET("", auth.OAuth2Login(db))
	oauth2.GET("/callback", auth.OAuth2Callback(db))

	// Source code management clients indexed by access token.
	scms := make(map[string]scm.SCM)

	api := e.Group("/api/v1")
	api.Use(auth.AccessControl(db, scms))

	api.GET("/user", web.GetSelf())

	users := api.Group("/users")
	users.GET("", web.GetUsers(db))
	users.GET("/:uid", web.GetUser(db))
	users.PATCH("/:uid", web.PatchUser(db))
	users.GET("/:uid/courses", web.ListCoursesWithEnrollment(db))

	courses := api.Group("/courses")
	courses.GET("", web.ListCourses(db))
	// TODO: Pass in webhook URLs and secrets for each registered provider.
	courses.POST("", web.NewCourse(l, db))
	courses.GET("/:cid", web.GetCourse(db))
	// TODO: Pass in webhook URLs and secrets for each registered provider.
	// TODO: Check if webhook exists and if not create a new one.
	courses.PUT("/:cid", web.UpdateCourse(db))
	courses.GET("/:cid/users", web.GetEnrollmentsByCourse(db))
	// TODO: Check if user is a member of a course, returns 404 or enrollment status.
	courses.GET("/:cid/users/:uid", echo.NotFoundHandler)
	courses.PUT("/:cid/users/:uid", web.SetEnrollment(db))
	courses.GET("/:cid/assignments", web.ListAssignments(db))

	api.POST("/directories", web.ListDirectories())

	index := func(c echo.Context) error {
		return c.File(entryPoint)
	}
	e.GET("/app", index)
	e.GET("/app/*", index)

	// TODO: Whitelisted files only.
	e.Static("/", *public)

	go func() {
		if err := e.Start(*httpAddr); err == http.ErrServerClosed {
			l.Warn("shutting down the server")
			return
		}
		l.WithError(err).Fatal("could not start server")
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		l.WithError(err).Fatal("failure during server shutdown")
	}
}

// In Windows, mime.type loads the file extensions from registry which
// usually has the wrong content-type associated with the file extension.
// This will enforce the correct types for the most used mime types
func setDefaultMimeTypes() {
	mime.AddExtensionType(".html", "text/html")
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")

	// Useful for debugging in browser
	mime.AddExtensionType(".jsx", "application/javascript")
	mime.AddExtensionType(".map", "application/json")
	mime.AddExtensionType(".ts", "application/x-typescript")
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

func getEventsURL(baseURL, provider string) string {
	return getURL(baseURL, "hook", provider, "events")
}

func getCallbackURL(baseURL, provider string) string {
	return getURL(baseURL, "auth", provider, "callback")
}

func getURL(baseURL, route, provider, endpoint string) string {
	return "https://" + baseURL + "/" + route + "/" + provider + "/" + endpoint
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
