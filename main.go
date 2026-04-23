package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/doc"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/manifest"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func init() {
	mustAddExtensionType := func(ext, typ string) {
		if err := mime.AddExtensionType(ext, typ); err != nil {
			panic(err)
		}
	}

	// On Windows, mime types are read from the registry, which often has
	// outdated content qf. This enforces that the correct mime types
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
		dbFile = flag.String("database.file", env.DatabasePath(), "database file")
		public = flag.String("http.public", env.PublicDir(), "path to content to serve")
		dev    = flag.Bool("dev", false, "run development server with self-signed certificates")
		secret = flag.Bool("secret", false, "force regeneration of JWT signing secret (will log out all users)")
	)
	flag.Parse()

	// Load environment variables from $QUICKFEED/.env.
	// Will not override variables already defined in the environment.
	const envFile = ".env"
	if err := env.Load(env.RootEnv(envFile)); err != nil {
		log.Fatal(err)
	}
	// Important to load the environment variables before initializing the provider.
	env.InitProvider()

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

	handler, cleanup, err := initWebServer(*dbFile, *public)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	log.Print("Callback: ", auth.GetCallbackURL())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *dev {
		// Wrap handler with file watcher for live-reloading in development mode.
		handler = web.WatchHandler(ctx, handler)
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

// runAppCreation runs the GitHub app creation flow after checking prerequisites.
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

// initWebServer initializes the QuickFeed web server components.
func initWebServer(dbFile, public string) (http.Handler, func(), error) {
	q := &quickfeed{}
	var err error

	q.logger, err = qlog.Zap()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize logger: %v", err)
	}

	q.db, err = database.NewGormDB(dbFile, q.logger)
	if err != nil {
		return nil, q.cleanup, fmt.Errorf("failed to connect to database: %v", err)
	}

	q.runner, err = ci.NewDockerCI(q.logger.Sugar())
	if err != nil {
		return nil, q.cleanup, fmt.Errorf("failed to set up docker client: %v", err)
	}

	tm, err := auth.NewTokenManager(q.db)
	if err != nil {
		return nil, q.cleanup, err
	}

	scmMgr, err := scm.NewSCMManager()
	if err != nil {
		return nil, q.cleanup, err
	}

	qfService := web.NewQuickFeedService(q.logger, q.db, scmMgr, q.runner, tm)
	// Register HTTP endpoints and webhooks
	router := qfService.RegisterRouter(os.Getenv("QUICKFEED_WEBHOOK_SECRET"), public)

	return h2c.NewHandler(router, &http2.Server{}), q.cleanup, nil
}

type quickfeed struct {
	logger *zap.Logger
	db     *database.GormDB
	runner *ci.Docker
}

func (q *quickfeed) cleanup() {
	var err error
	if q.runner != nil {
		if e := q.runner.Close(); e != nil {
			err = fmt.Errorf("failed to close runner: %w", e)
		}
	}
	if q.db != nil {
		if e := q.db.Close(); e != nil {
			err = errors.Join(err, fmt.Errorf("failed to close database: %w", e))
		}
	}
	if q.logger != nil {
		if e := q.logger.Sync(); e != nil {
			err = errors.Join(err, fmt.Errorf("failed to sync logger: %w", e))
		}
	}
	if err != nil {
		log.Printf("Cleanup error:\n%v", err)
	}
}

func checkDomain() error {
	if env.IsDomainLocal() {
		msg := `
WARNING: You are creating a GitHub app on a local or private domain: %q.
This is only for development purposes.
In this mode, QuickFeed will not be able to receive webhook events from GitHub.
To receive webhook events, you must run QuickFeed on a public domain or use a tunneling service like ngrok.
`
		fmt.Printf(msg, env.Domain())
		fmt.Printf("Read more here: %s\n\n", doc.DeployURL)
		fmt.Print("Do you want to continue? (Y/n) ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "Y" && answer != "y" {
			return fmt.Errorf("aborting %s GitHub App creation", env.AppName())
		}
	}
	return nil
}
