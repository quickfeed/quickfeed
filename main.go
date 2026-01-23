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
	"github.com/quickfeed/quickfeed/internal/cert"
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
		dbFile  = flag.String("database.file", env.DatabasePath(), "database file")
		public  = flag.String("http.public", env.PublicDir(), "path to content to serve")
		dev     = flag.Bool("dev", false, "run development server with self-signed certificates")
		genCert = flag.Bool("gen", false, "generate self-signed certificates for development")
		newApp  = flag.Bool("new", false, "create new GitHub app")
		secret  = flag.Bool("secret", false, "create new secret for JWT signing")
	)
	flag.Parse()

	// Load environment variables from $QUICKFEED/.env.
	// Will not override variables already defined in the environment.
	const envFile = ".env"
	if err := env.Load(env.RootEnv(envFile)); err != nil {
		log.Fatal(err)
	}

	// Handle certificate generation separately from server startup
	if *genCert {
		if err := generateCertificates(); err != nil {
			log.Fatalf("Failed to generate certificates: %v", err)
		}
		return
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
		if *dev {
			log.Print("To generate self-signed certificates, run: quickfeed -gencert")
		}
		return
	}

	go gracefulShutdown(ctx, srv)

	if err := srv.Serve(); err != nil {
		log.Printf("Failed to start QuickFeed server: %v", err)
		return
	}
	log.Println("QuickFeed shut down gracefully")
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

// generateCertificates generates self-signed certificates for development.
// The certificates are stored in $HOME/.quickfeed/certs/ by default.
func generateCertificates() error {
	log.Printf("Generating self-signed certificates for domain: %s", env.Domain())
	if err := cert.GenerateSelfSignedCert(cert.Options{
		FullchainFile: env.FullchainFile(),
		PrivKeyFile:   env.PrivKeyFile(),
		CAFile:        env.CAFile(),
		Hosts:         env.Domain(),
	}); err != nil {
		return fmt.Errorf("failed to generate self-signed certificates: %w", err)
	}
	log.Printf("Certificates successfully generated at: %s", env.CertPath())
	log.Print("Adding certificate to local trust store (may require elevated privileges on some systems)")
	if err := cert.AddTrustedCert(env.CAFile()); err != nil {
		return fmt.Errorf("failed to install self-signed certificate: %w", err)
	}
	log.Print("Certificate successfully added to trust store")
	return nil
}
