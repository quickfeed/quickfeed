package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"mime"
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
		newApp = flag.Bool("new", false, "create new GitHub app")
		secret = flag.Bool("secret", false, "create new secret for JWT signing")
	)
	flag.Parse()

	// Load environment variables from $QUICKFEED/.env.
	// Will not override variables already defined in the environment.
	const envFile = ".env"
	if err := env.Load(env.RootEnv(envFile)); err != nil {
		log.Fatal(err)
	}
	if env.Domain() == "localhost" {
		log.Fatal(`Domain "localhost" is unsupported; use "127.0.0.1" instead.`)
	}

	var srvFn web.ServerType
	if *dev {
		srvFn = web.NewDevelopmentServer
	} else {
		srvFn = web.NewProductionServer
	}
	log.Printf("Starting QuickFeed on %s", env.DomainWithPort())

	if *newApp {
		if err := manifest.ReadyForAppCreation(envFile, checkDomain); err != nil {
			log.Fatal(err)
		}
		if err := manifest.CreateNewQuickFeedApp(srvFn, envFile, *dev); err != nil {
			log.Fatal(err)
		}
	}
	if *secret {
		log.Println("Generating new random secret for signing JWT tokens...")
		if err := env.NewAuthSecret(envFile); err != nil {
			log.Fatalf("Failed to save secret: %v", err)
		}
	}
	if *secret || *newApp {
		// Refresh environment variables
		if err := env.Load(env.RootEnv(envFile)); err != nil {
			log.Fatal(err)
		}
	}
	if env.AuthSecret() == "" {
		log.Fatal("Required QUICKFEED_AUTH_SECRET is not set")
	}

	logger, err := qlog.Zap()
	if err != nil {
		log.Fatalf("Can't initialize logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	db, err := database.NewGormDB(*dbFile, logger)
	if err != nil {
		log.Fatalf("Can't connect to database: %v", err)
	}

	// Holds references for activated providers for current user token
	bh := web.BaseHookOptions{
		BaseURL: env.Domain(),
		Secret:  os.Getenv("QUICKFEED_WEBHOOK_SECRET"),
	}

	scmConfig, err := scm.NewSCMConfig()
	if err != nil {
		log.Fatal(err)
	}

	tokenManager, err := auth.NewTokenManager(db)
	if err != nil {
		log.Fatal(err)
	}
	authConfig := auth.NewGitHubConfig(scmConfig)
	log.Print("Callback: ", authConfig.RedirectURL)
	scmManager := scm.NewSCMManager(scmConfig)

	runner, err := ci.NewDockerCI(logger.Sugar())
	if err != nil {
		log.Fatalf("Failed to set up docker client: %v", err)
	}
	defer runner.Close()

	qfService := web.NewQuickFeedService(logger, db, scmManager, bh, runner)
	// Register HTTP endpoints and webhooks
	router := qfService.RegisterRouter(tokenManager, authConfig, *public)
	handler := h2c.NewHandler(router, &http2.Server{})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *dev {
		// Wrap handler with file watcher
		// for live-reloading in development mode.
		handler = web.WatchHandler(ctx, handler)
	}

	srv, err := srvFn(handler)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Graceful shutdown failed: %v", err)
		}
	}()

	if err := srv.Serve(); err != nil {
		log.Fatalf("Failed to start QuickFeed server: %v", err)
	}
	log.Println("QuickFeed shut down gracefully")
}

func checkDomain() error {
	if env.Domain() == "127.0.0.1" {
		msg := `
WARNING: You are creating a GitHub app on "127.0.0.1".
This is only for development purposes.
In this mode, QuickFeed will not be able to receive webhook events from GitHub.
To receive webhook events, you must run QuickFeed on a public domain or use a tunneling service like ngrok.
`
		fmt.Println(msg)
		fmt.Printf("Read more here: %s\n\n", doc.DeployURL)
		fmt.Print("Do you want to continue? (Y/n) ")
		var answer string
		fmt.Scanln(&answer)
		if !(answer == "Y" || answer == "y") {
			return fmt.Errorf("aborting %s GitHub App creation", env.AppName())
		}
	}
	return nil
}
