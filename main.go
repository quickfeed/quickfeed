package main

import (
	"flag"
	"log"
	"mime"
	"net"
	"os"
	"path/filepath"

	"github.com/autograde/aguis/envoy"
	"github.com/autograde/aguis/web"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"

	"github.com/autograde/aguis/logger"
	"github.com/autograde/aguis/web/grpc_service"

	http "github.com/autograde/aguis/web/webserver"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func init() {
	mustAddExtensionType := func(ext, typ string) {
		if err := mime.AddExtensionType(ext, typ); err != nil {
			panic(err)
		}
	}

	// On Windows, mime types are read from the registry, which often has
	// outdated content types. This enforces that the correct mime types
	// are used on all platforms.
	mustAddExtensionType(".html", "text/html")
	mustAddExtensionType(".css", "text/css")
	mustAddExtensionType(".js", "application/javascript")
	mustAddExtensionType(".jsx", "application/javascript")
	mustAddExtensionType(".map", "application/json")
	mustAddExtensionType(".ts", "application/x-typescript")
}

func main() {
	var (
		httpAddr = flag.String("http.addr", ":8081", "HTTP listen address")
		public   = flag.String("http.public", "public", "directory to server static files from")

		buildscripts = flag.String("script.path", "buildscripts", "Directory with docker build scripts")

		dbFile = flag.String("database.file", tempFile("ag.db"), "database file")

		baseURL = flag.String("service.url", "localhost", "service base url")

		fake = flag.Bool("provider.fake", false, "enable fake provider")

		grpcAddr = flag.String("grpc.addr", ":9090", "gRPC listen address")
	)
	flag.Parse()

	l := logrus.New()
	l.Formatter = logger.NewDevFormatter(l.Formatter)

	db, err := database.NewGormDB("sqlite3", *dbFile, database.Logger{Logger: l})
	if err != nil {
		l.WithError(err).Fatal("could not connect to db")
	}
	defer func() {
		if dbErr := db.Close(); dbErr != nil {
			l.WithError(dbErr).Warn("error closing database")
		}
	}()

	/* Will start envoy in a docker container from image unless it is already running,
	if there is no image, will pull the envoy docker image, build a local envoy image with options from envoy.yaml
	 and start a container from the new image */
	go func() {
		envoy.StartEnvoy()
	}()

	// holds references for activated providers for current user token
	scms := make(map[string]scm.SCM)
	bh := web.BaseHookOptions{
		BaseURL: *baseURL,
		Secret:  os.Getenv("WEBHOOK_SECRET"),
	}
<<<<<<< HEAD
=======

	store := newStore([]byte("secret"))
	gothic.Store = store
	e := newServer(l, store)
	enabled := enableProviders(l, *baseURL, *fake)
	registerWebhooks(l, e, db, bh.Secret, enabled, buildscripts)
	registerAuth(e, db)
	registerAPI(l, e, db, &bh)
	registerFrontend(e, entryPoint, *public)
	run(l, e, *httpAddr)
}

func newServer(l *logrus.Logger, store sessions.Store) *echo.Echo {
	e := echo.New()
	e.Logger = web.EchoLogger{Logger: l}
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

func enableProviders(l logrus.FieldLogger, baseURL string, fake bool) map[string]bool {
	enabled := make(map[string]bool)

	if ok := auth.EnableProvider(&auth.Provider{
		Name:          "github",
		KeyEnv:        "GITHUB_KEY",
		SecretEnv:     "GITHUB_SECRET",
		CallbackURL:   auth.GetCallbackURL(baseURL, "github"),
		StudentScopes: []string{},
		TeacherScopes: []string{"user", "repo", "delete_repo"},
	}, func(key, secret, callback string, scopes ...string) goth.Provider {
		return github.New(key, secret, callback, scopes...)
	}); ok {
		enabled["github"] = true
	} else {
		l.WithFields(logrus.Fields{
			"provider": "github",
			"enabled":  false,
		}).Warn("environment variables not set")
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
		l.WithFields(logrus.Fields{
			"provider": "gitlab",
			"enabled":  false,
		}).Warn("environment variables not set")
	}

	if fake {
		l.Warn("fake provider enabled")
		goth.UseProviders(&auth.FakeProvider{
			Callback: auth.GetCallbackURL(baseURL, "fake"),
		})
		goth.UseProviders(&auth.FakeProvider{
			Callback: auth.GetCallbackURL(baseURL, "fake-teacher"),
		})
	}

	return enabled
}

func registerWebhooks(logger logrus.FieldLogger, e *echo.Echo, db database.Database, secret string, enabled map[string]bool, buildscripts *string) {
	webhooks.DefaultLog = web.WebhookLogger{FieldLogger: logger}

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

func registerAuth(e *echo.Echo, db database.Database) {
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

	oauth2 := e.Group("/auth/:provider", withProvider, auth.PreAuth(db))
	oauth2.GET("", auth.OAuth2Login(db))
	oauth2.GET("/callback", auth.OAuth2Callback(db))
	e.GET("/logout", auth.OAuth2Logout())
}

func registerAPI(l logrus.FieldLogger, e *echo.Echo, db database.Database, bh *web.BaseHookOptions) {
	// Source code management clients indexed by access token.
	scms := make(map[string]scm.SCM)

	api := e.Group("/api/v1")
	api.Use(auth.AccessControl(db, scms))

	var providers []string
	for _, provider := range goth.GetProviders() {
		if !strings.HasSuffix(provider.Name(), auth.TeacherSuffix) {
			providers = append(providers, provider.Name())
		}
	}
	api.GET("/providers", func(c echo.Context) error {
		return c.JSONPretty(http.StatusOK, &providers, "\t")
	})

	api.GET("/user", web.GetSelf())

	users := api.Group("/users")
	users.GET("", web.GetUsers(db))
	users.GET("/:uid", web.GetUser(db))
	users.PATCH("/:uid", web.PatchUser(db))
	users.GET("/:uid/courses", web.ListCoursesWithEnrollment(db))
	users.GET("/:uid/courses/:cid/group", web.GetGroupByUserAndCourse(db))

	courses := api.Group("/courses")
	courses.GET("", web.ListCourses(db))
	courses.POST("", web.NewCourse(l, db, bh))
	courses.GET("/:cid", web.GetCourse(db))

	courses.POST("/:cid/refresh", web.RefreshCourse(l, db))
	// TODO: Pass in webhook URLs and secrets for each registered provider.
	// TODO: Check if webhook exists and if not create a new one.
	courses.PUT("/:cid", web.UpdateCourse(db))
	courses.GET("/:cid/users", web.GetEnrollmentsByCourse(db))
	// TODO: Check if user is a member of a course, returns 404 or enrollment status.
	courses.GET("/:cid/users/:uid", echo.NotFoundHandler)
	courses.GET("/:cid/users/:uid/submissions", web.ListSubmissions(db))
	courses.POST("/:cid/users/:uid", web.CreateEnrollment(db))
	courses.PATCH("/:cid/users/:uid", web.UpdateEnrollment(db))
	courses.GET("/:cid/assignments", web.ListAssignments(db))
	// TODO: Endpoints needs to be fixed.
	courses.GET("/:cid/assignments/:aid/submission", web.GetSubmission(db))
	courses.GET("/:cid/submissions", web.ListSubmissions(db))
	courses.POST("/:cid/groups", web.NewGroup(db))
	courses.PUT("/:cid/groups/:gid", web.UpdateGroup(l, db))
	courses.GET("/:cid/groups", web.GetGroups(db))
	courses.GET("/:cid/groups/:gid/submissions", web.ListGroupSubmissions(db))
	courses.GET("/:cid/courseinformation", web.GetCourseInformationURL(db))
	courses.GET("/:cid/repositoryurl", web.GetRepositoryURL(db))

	submissions := api.Group("/submissions")
	submissions.PATCH("/:sid", web.UpdateSubmission(db))

	groups := api.Group("/groups")
	groups.GET("/:gid", web.GetGroup(db))
	groups.PATCH("/:gid", web.PatchGroup(l, db))
	groups.DELETE("/:gid", web.DeleteGroup(db))

	api.POST("/directories", web.ListDirectories())
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

func run(l logrus.FieldLogger, e *echo.Echo, httpAddr string) {
>>>>>>> origin/master
	go func() {
		http.NewWebServer(db, bh, l, *public, *httpAddr, *baseURL, *fake, *buildscripts, scms)
	}()

	lis, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAutograderServiceServer(grpcServer, grpc_service.NewAutograderService(db, scms, bh))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

func tempFile(name string) string {
	return filepath.Join(os.TempDir(), name)
}
