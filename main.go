package main

import (
	"flag"
	"log"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/autograde/quickfeed/ci"
	logq "github.com/autograde/quickfeed/log"
	"github.com/autograde/quickfeed/web"
	"github.com/autograde/quickfeed/web/auth"
	"github.com/autograde/quickfeed/web/config"
	"github.com/improbable-eng/grpc-web/go/grpcweb"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	// Create some standard server metrics.
	grpcMetrics := grpc_prometheus.NewServerMetrics()

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

	reg.MustRegister(
		grpcMetrics,
		pb.AgFailedMethodsMetric,
		pb.AgMethodSuccessRateMetric,
		pb.AgResponseTimeByMethodsMetric,
	)
}

// Create a metrics registry.
var reg = prometheus.NewRegistry()

func main() {
	var (
		baseURL  = flag.String("service.url", "", "base service DNS name")
		dbFile   = flag.String("database.file", "qf.db", "database file")
		public   = flag.String("http.public", "public", "path to content to serve")
		httpAddr = flag.String("http.addr", ":8081", "HTTP listen address")
		// grpcAddr = flag.String("grpc.addr", ":9090", "gRPC listen address")
	)
	flag.Parse()

	logger := logq.Zap(true)
	defer logger.Sync()

	db, err := database.NewGormDB(*dbFile, logger)
	if err != nil {
		log.Fatalf("can't connect to database: %v\n", err)
	}

	// holds references for activated providers for current user token
	scms := auth.NewScms()
	// bh := web.BaseHookOptions{
	// 	BaseURL: *baseURL,
	// 	Secret:  os.Getenv("WEBHOOK_SECRET"),
	// }

	runner, err := ci.NewDockerCI(logger)
	if err != nil {
		log.Fatalf("failed to set up docker client: %v\n", err)
	}
	defer runner.Close()

	// TODO: find and replace (if possible) all occasions where this tiken is used
	// Add application token for external applications (to allow invoking gRPC methods)
	// TODO(meling): this is a temporary solution, and we should find a better way to do this
	token := os.Getenv("QUICKFEED_AUTH_TOKEN")
	if len(token) > 16 {
		auth.Add(token, 1)
		log.Println("Added application token")
	}

	serverConfig := config.NewConfig(*baseURL, *public, *httpAddr)

	// TODO: register GitHub App
	githubApp, err := config.NewApp()
	if err != nil {
		log.Fatalf("failed to start GitHub app: %v/n", err)
	}
	serverConfig.App = githubApp
	agService := web.NewAutograderService(logger, db, scms, serverConfig, runner)
	APIServer, err := serverConfig.GenerateTLSApi()
	if err != nil {
		log.Fatalf("failed to generate TLS grpc API: %v/n", err)
	}
	pb.RegisterAutograderServiceServer(APIServer, agService)

	grpcWebServer := grpcweb.WrapServer(APIServer)
	multiplexer := config.GrpcMultiplexer{
		grpcWebServer,
	}
	router := http.NewServeMux()
	staticHandler := http.FileServer(http.Dir(serverConfig.Endpoints.Public))
	router.Handle("/", multiplexer.MultiplexerHandler(staticHandler))

	//////////////////////////
	// TODO: register auth endpoints here: RegisterAuth(router, config.App or full config)

	//////////////////////////

	// Create a HTTP server and bind the router to it, and set wanted address
	srv := &http.Server{
		Handler:      router,
		Addr:         serverConfig.Endpoints.HttpAddress,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	// Serve the webapp over TLS
	log.Fatal(srv.ListenAndServeTLS(serverConfig.Paths.PemPath, serverConfig.Paths.KeyPath))

	///////////////////////////////////
	go web.New(agService, *public, *httpAddr)

	// lis, err := net.Listen("tcp", *grpcAddr)
	// if err != nil {
	// 	log.Fatalf("failed to start tcp listener: %v\n", err)
	// }
	// opt := grpc.ChainUnaryInterceptor(auth.UserVerifier(), pb.Interceptor(logger))
	// grpcServer := grpc.NewServer(opt)
	// Create a HTTP server for prometheus.
	// httpServer := &http.Server{
	// 	Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
	// 	Addr:    fmt.Sprintf("0.0.0.0:%d", 9097),
	// }
	// go func() {
	// 	if err := httpServer.ListenAndServe(); err != nil {
	// 		log.Fatal("Unable to start a http server.")
	// 	}
	// }()

	// pb.RegisterAutograderServiceServer(grpcServer, agService)
	// if err := grpcServer.Serve(lis); err != nil {
	// 	log.Fatalf("failed to start grpc server: %v\n", err)
	// }
}
