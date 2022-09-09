package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/cert"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"github.com/quickfeed/quickfeed/web/manifest"
	"golang.org/x/crypto/acme/autocert"
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
		baseURL  = flag.String("service.url", "", "base service DNS name")
		dbFile   = flag.String("database.file", "qf.db", "database file")
		public   = flag.String("http.public", "public", "path to content to serve")
		httpAddr = flag.String("http.addr", ":8081", "HTTP listen address")
		dev      = flag.Bool("dev", false, "running server locally")
		newApp   = flag.Bool("new", false, "create new GitHub app")
	)
	flag.Parse()

	if *dev {
		*baseURL = "127.0.0.1" + *httpAddr
	}

	// Load environment variables from $QUICKFEED/.env.
	// Will not override variables already defined in the environment.
	if err := env.Load(""); err != nil {
		log.Fatal(err)
	}

	if *newApp {
		if err := manifest.StartAppCreationFlow(*httpAddr); err != nil {
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
		BaseURL: *baseURL,
		Secret:  os.Getenv("WEBHOOK_SECRET"),
	}

	scmConfig, err := scm.NewSCMConfig()
	if err != nil {
		log.Fatal(err)
	}

	tokenManager, err := auth.NewTokenManager(db, *baseURL)
	if err != nil {
		log.Fatal(err)
	}
	authConfig := auth.NewGitHubConfig(*baseURL, scmConfig)
	logger.Sugar().Debug("CALLBACK: ", authConfig.RedirectURL)
	scmManager := scm.NewSCMManager(scmConfig)

	runner, err := ci.NewDockerCI(logger.Sugar())
	if err != nil {
		log.Fatalf("Failed to set up docker client: %v", err)
	}
	defer runner.Close()

	certFile := env.CertFile()
	certKey := env.KeyFile()
	qfService := web.NewQuickFeedService(logger, db, scmManager, bh, runner)

	// Register HTTP endpoints and webhooks
	router := qfService.RegisterRouter(tokenManager, authConfig, *public)

	// Create an HTTP server for prometheus.
	httpServer := interceptor.MetricsServer(9097)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start a http server: %v", err)
		}
	}()
	muxServer := &http.Server{
		Handler:           h2c.NewHandler(router, &http2.Server{}),
		Addr:              *httpAddr,
		ReadHeaderTimeout: 3 * time.Second, // to prevent Slowloris (CWE-400)
		WriteTimeout:      2 * time.Minute,
		ReadTimeout:       2 * time.Minute,
	}
	if *dev {
		logger.Sugar().Debugf("Starting server in development mode on %s", *httpAddr)
		if err := muxServer.ListenAndServeTLS(certFile, certKey); err != nil {
			log.Fatalf("Failed to start grpc server: %v", err)
			return
		}
	} else {
		logger.Sugar().Debugf("Starting server in production mode on %s", *baseURL)
	}
	whitelist, err := env.Whitelist()
	if err != nil {
		log.Fatalf("Failed to get whitelist: %v", err)
	}
	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(env.CertPath()),
		HostPolicy: autocert.HostWhitelist(
			whitelist...,
		),
	}
	muxServer.TLSConfig = certManager.TLSConfig()

	// Redirect all HTTP traffic to HTTPS.
	go func() {
		redirectSrv := &http.Server{
			Handler:           certManager.HTTPHandler(nil),
			Addr:              ":http",
			ReadHeaderTimeout: 3 * time.Second, // to prevent Slowloris (CWE-400)
		}
		if err := redirectSrv.ListenAndServe(); err != nil {
			log.Printf("Failed to start redirect http server: %v", err)
			return
		}
	}()

	// Start the HTTPS server.
	if err := muxServer.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Failed to start grpc server: %v", err)
	}
}

func prodServer(addr string) (*http.Server, error) {
	whitelist, err := env.Whitelist()
	if err != nil {
		return nil, fmt.Errorf("failed to get whitelist: %w", err)
	}
	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(env.CertPath()),
		HostPolicy: autocert.HostWhitelist(
			whitelist...,
		),
	}
	return &http.Server{
		Addr:         addr,
		WriteTimeout: 2 * time.Minute,
		ReadTimeout:  2 * time.Minute,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MaxVersion:     tls.VersionTLS13,
			MinVersion:     tls.VersionTLS12,
		},
	}, nil
}

// devServer returns a http.Server with self-signed certificates for development-use only.
func devServer(addr string) (*http.Server, error) {
	certificate, err := tls.LoadX509KeyPair(env.CertFile(), env.KeyFile())
	if err != nil {
		// Couldn't load credentials; generate self-signed certificates.
		log.Println("Generating self-signed certificates.")
		if err := cert.GenerateSelfSignedCert(cert.Options{
			KeyFile:  env.KeyFile(),
			CertFile: env.CertFile(),
			Hosts:    env.Domain(),
		}); err != nil {
			return nil, fmt.Errorf("failed to generate self-signed certificates: %w", err)
		}
		log.Printf("Certificates successfully generated at: %s", env.CertPath())
	} else {
		log.Println("Existing credentials successfully loaded.")
	}
	return &http.Server{
		Addr:         addr,
		WriteTimeout: 2 * time.Minute,
		ReadTimeout:  2 * time.Minute,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{certificate},
		},
	}, nil
}
