package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/doc"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/hookforward"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/courselog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/hooks"
	"github.com/quickfeed/quickfeed/web/manifest"
)

func main() {
	var (
		dbFile = flag.String("database.file", env.DatabasePath(), "database file")
		public = flag.String("http.public", env.PublicDir(), "path to content to serve")
		dev    = flag.Bool("dev", false, "run development server with self-signed certificates")
		secret = flag.Bool("secret", false, "force regeneration of JWT signing secret (will log out all users)")
		hook   = flag.Bool("hook", false, "forward GitHub webhook events to this server (requires 'make webhook-setup')")
	)
	flag.Parse()

	// Load environment variables from $QUICKFEED/.env.
	// Will not override variables already defined in the environment.
	const envFile = ".env"
	if err := env.Load(env.RootEnv(envFile)); err != nil {
		log.Fatal(err)
	}

	// Determine server type based on mode
	var srvFn web.ServerType
	if *dev {
		srvFn = web.NewDevelopmentServer
	} else {
		srvFn = web.NewProductionServer
	}

	// Ensure environment is ready: auth secret, domain validation, certificates (dev mode)
	if err := env.EnsureReady(*dev, envFile, *secret); err != nil {
		log.Fatal(err)
	}

	// If app data is missing, run the app creation flow
	if env.NeedsAppCreation() {
		if err := runAppCreation(envFile, *dev, srvFn); err != nil {
			log.Fatal(err)
		}
		// Reload environment after app creation
		if err := env.Load(env.RootEnv(envFile)); err != nil {
			log.Fatal(err)
		}
	}

	// Final validation: ensure we have everything needed
	if err := env.CheckAppData(); err != nil {
		log.Fatalf("Missing app configuration:\n%v", err)
	}

	log.Printf("Starting QuickFeed on %s", env.DomainWithPort())

	q, handler, err := initWebServer(*dbFile, *public)
	if err != nil {
		q.cleanup()
		log.Fatal(err)
	}
	defer q.cleanup()

	log.Print("Callback: ", auth.GetCallbackURL())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *dev {
		// Wrap handler with file watcher for live-reloading in development mode.
		handler = web.WatchHandler(ctx, handler)
	}
	if *hook {
		forwardWebhooks(ctx, q)
	}

	srv, err := srvFn(handler)
	if err != nil {
		log.Printf("Failed to start server: %v", err)
		return
	}

	go gracefulShutdown(ctx, srv)

	if err := srv.Serve(); err != nil {
		log.Printf("Failed to start QuickFeed server: %v", err)
		return
	}
	log.Println("QuickFeed shut down gracefully")
}

// runAppCreation runs the GitHub App creation flow after checking prerequisites.
func runAppCreation(envFile string, dev bool, srvFn web.ServerType) error {
	if err := checkDomain(); err != nil {
		return err
	}
	return manifest.CreateNewQuickFeedApp(srvFn, envFile, dev)
}

// gracefulShutdown blocks waiting for the context to be done (SIGINT or SIGTERM)
// and then attempts to gracefully shut down the server.
func gracefulShutdown(ctx context.Context, srv *web.Server) {
	<-ctx.Done()
	log.Print("Shutting down server...")
	shutDownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutDownCtx); err != nil {
		log.Printf("Graceful shutdown failed: %v", err)
	}
}

// forwardWebhooks starts forwarding GitHub webhook events to this server for
// every course organization in the database. Forwarding is a development
// convenience, so a failure is only a warning: the server is fully functional
// without it, and refusing to start would be a worse trade than running with
// webhook events delivered the way they are in production.
func forwardWebhooks(ctx context.Context, q *quickfeed) {
	orgs, err := q.organizations()
	if err != nil {
		q.logger.Warn("not forwarding webhook events", label.Error, err)
		return
	}
	if len(orgs) == 0 {
		q.logger.Warn("not forwarding webhook events: no course organizations in the database")
		return
	}
	err = hookforward.Start(ctx, q.logger, hookforward.Options{
		Orgs:   orgs,
		URL:    hookforward.URL(env.DomainWithPort()),
		Events: hooks.Events(),
		Secret: os.Getenv("QUICKFEED_WEBHOOK_SECRET"),
	})
	if err != nil {
		q.logger.Warn("not forwarding webhook events", label.Error, err)
	}
}

// initWebServer initializes the QuickFeed web server components.
// It returns the quickfeed struct even on error, so that the caller can
// release whatever was set up before the failure.
func initWebServer(dbFile, public string) (*quickfeed, http.Handler, error) {
	q := &quickfeed{}
	var err error

	operator := qlog.New(os.Stderr)
	q.courseLogs, err = courselog.NewStore(env.CourseLogDir(), operator)
	if err != nil {
		return q, nil, fmt.Errorf("setting up course log store: %w", err)
	}
	q.logger = qlog.WithSink(operator, courselog.NewHandler(q.courseLogs))
	qlog.SetDefault(q.logger)

	q.db, err = database.NewGormDB(dbFile, q.logger)
	if err != nil {
		return q, nil, fmt.Errorf("connecting to database: %w", err)
	}

	q.runner, err = ci.NewDockerCI()
	if err != nil {
		return q, nil, fmt.Errorf("setting up docker client: %w", err)
	}

	tm, err := auth.NewTokenManager(q.db)
	if err != nil {
		return q, nil, err
	}

	scmMgr, err := scm.NewSCMManager()
	if err != nil {
		return q, nil, err
	}

	qfService := web.NewQuickFeedService(q.logger, q.db, scmMgr, q.runner, tm, q.courseLogs)
	// Register HTTP endpoints and webhooks
	router := qfService.RegisterRouter(os.Getenv("QUICKFEED_WEBHOOK_SECRET"), public)

	return q, router, nil
}

type quickfeed struct {
	logger     *slog.Logger
	db         *database.GormDB
	runner     *ci.Docker
	courseLogs *courselog.Store
}

// organizations returns the distinct SCM organization names of the courses in
// the database, in sorted order.
func (q *quickfeed) organizations() ([]string, error) {
	courses, err := q.db.GetCourses()
	if err != nil {
		return nil, fmt.Errorf("loading courses: %w", err)
	}
	orgs := make(map[string]bool)
	for _, course := range courses {
		if org := course.GetScmOrganizationName(); org != "" {
			orgs[org] = true
		}
	}
	return slices.Sorted(maps.Keys(orgs)), nil
}

func (q *quickfeed) cleanup() {
	var err error
	if q.runner != nil {
		if e := q.runner.Close(); e != nil {
			err = fmt.Errorf("closing runner: %w", e)
		}
	}
	if q.db != nil {
		if e := q.db.Close(); e != nil {
			err = errors.Join(err, fmt.Errorf("closing database: %w", e))
		}
	}
	if q.courseLogs != nil {
		if e := q.courseLogs.Close(); e != nil {
			err = errors.Join(err, fmt.Errorf("closing course log store: %w", e))
		}
	}
	if err != nil {
		log.Printf("Cleanup error:\n%v", err)
	}
}

func checkDomain() error {
	if env.IsDomainLocal() {
		msg := `
WARNING: You are creating a GitHub App on a local or private domain: %q.
This is only for development purposes.
In this mode, QuickFeed will not be able to receive webhook events from GitHub.
To receive webhook events, you must run QuickFeed on a public domain or use a tunneling service like ngrok.
`
		fmt.Printf(msg, env.Domain())
		fmt.Printf("Read more here: %s\n\n", doc.DeployURL)
		fmt.Print("Do you want to continue? (Y/n) ")
		var answer string
		if _, err := fmt.Scanln(&answer); err != nil && err.Error() != "unexpected newline" {
			return fmt.Errorf("failed to read answer: %w", err)
		}
		if answer != "Y" && answer != "y" {
			return fmt.Errorf("aborting %s GitHub App creation", env.AppName())
		}
	}
	return nil
}
