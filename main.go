package main

import (
	"context"
	"errors"
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
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
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
		dbFile   = flag.String("database.file", "qf.db", "database file")
		public   = flag.String("http.public", "public", "path to content to serve")
		httpAddr = flag.String("http.addr", ":443", "HTTP listen address")
		dev      = flag.Bool("dev", false, "run development server with self-signed certificates")
		newApp   = flag.Bool("new", false, "create new GitHub app")
	)
	flag.Parse()

	// Load environment variables from $QUICKFEED/.env.
	// Will not override variables already defined in the environment.
	if err := env.Load(""); err != nil {
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
	log.Printf("Starting QuickFeed on %s%s", env.Domain(), *httpAddr)

	if *newApp {
		if err := createNewQuickFeedApp(srvFn, *httpAddr); err != nil {
			log.Fatal(err)
		}
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
		Secret:  os.Getenv("WEBHOOK_SECRET"),
	}

	scmConfig, err := scm.NewSCMConfig()
	if err != nil {
		log.Fatal(err)
	}

	tokenManager, err := auth.NewTokenManager(db)
	if err != nil {
		log.Fatal(err)
	}
	authConfig := auth.NewGitHubConfig(env.Domain(), scmConfig)
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

	// Create an HTTP server for prometheus.
	httpServer := interceptor.MetricsServer(9097)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start a http server: %v", err)
		}
	}()

	srv, err := srvFn(*httpAddr, handler)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
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

func createNewQuickFeedApp(srvFn web.ServerType, httpAddr string) error {
	if env.HasAppID() {
		return errors.New(".env already contains App information")
	}
	if env.Domain() == "127.0.0.1" {
		fmt.Printf("WARNING: You are creating an app on %s. Only for development purposes. Continue? (Y/n) ", env.Domain())
		var answer string
		fmt.Scanln(&answer)
		if !(answer == "Y" || answer == "y") {
			return fmt.Errorf("aborting %s GitHub App creation", env.AppName())
		}
	}

	m := manifest.New(env.Domain())
	server, err := srvFn(httpAddr, m.Handler())
	if err != nil {
		return err
	}
	return m.StartAppCreationFlow(server)
}
