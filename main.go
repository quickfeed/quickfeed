package main

import (
	"context"
	"flag"
	"log"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"google.golang.org/grpc"
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

	logger, err := qlog.Zap()
	if err != nil {
		log.Fatalf("Can't initialize logger: %v", err)
	}
	defer logger.Sync()

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
	var grpcServer *grpc.Server
	unaryOptions := grpc.ChainUnaryInterceptor(
		interceptor.Metrics(),
		interceptor.Validation(logger),
		interceptor.UnaryUserVerifier(logger.Sugar(), tokenManager),
		interceptor.AccessControl(logger.Sugar(), tokenManager),
		interceptor.TokenRefresher(logger.Sugar(), tokenManager),
	)
	streamOptions := grpc.ChainStreamInterceptor(interceptor.StreamUserVerifier(logger.Sugar(), tokenManager))
	if *dev {
		logger.Sugar().Debugf("Starting server in development mode on %s", *httpAddr)
		// In development, the server itself must maintain a TLS session.
		grpcServer, err = web.GRPCServerWithCredentials(certFile, certKey, unaryOptions, streamOptions)
		if err != nil {
			log.Fatalf("Failed to generate gRPC server credentials: %v", err)
		}
	} else {
		logger.Sugar().Debugf("Starting server in production mode on %s", *baseURL)
		// In production, the envoy proxy will manage TLS certificates, and
		// the gRPC server must be started without credentials.
		grpcServer = web.GRPCServer(unaryOptions, streamOptions)
	}
	qfService := web.NewQuickFeedService(logger, db, scmManager, bh, runner)
	if err := qfService.InitSCMs(context.Background()); err != nil {
		log.Fatalf("Failed to initialize SCM clients: %v", err)
	}

	qf.RegisterQuickFeedServiceServer(grpcServer, qfService)
	if err = web.VerifyAccessControlMethods(grpcServer); err != nil {
		log.Fatal(err)
	}

	multiplexer := web.GrpcMultiplexer{
		MuxServer: grpcweb.WrapServer(grpcServer),
	}

	// Register HTTP endpoints and webhooks
	router := qfService.RegisterRouter(tokenManager, authConfig, multiplexer, *public)

	// Create an HTTP server for prometheus.
	httpServer := interceptor.MetricsServer(9097)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start a http server: %v", err)
		}
	}()
	muxServer := &http.Server{
		Handler:      router,
		Addr:         *httpAddr,
		WriteTimeout: 2 * time.Minute,
		ReadTimeout:  2 * time.Minute,
	}
	if *dev {
		if err := muxServer.ListenAndServeTLS(certFile, certKey); err != nil {
			log.Fatalf("Failed to start grpc server: %v", err)
			return
		}
	}
	if err := muxServer.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start grpc server: %v", err)
	}
}
