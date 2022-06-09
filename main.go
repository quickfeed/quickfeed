package main

import (
	"flag"
	"log"
	"mime"
	"net/http"
	"time"

	"github.com/autograde/quickfeed/ci"
	logq "github.com/autograde/quickfeed/log"
	"github.com/autograde/quickfeed/scm"
	"github.com/autograde/quickfeed/web"
	"github.com/autograde/quickfeed/web/auth"
	"github.com/autograde/quickfeed/web/auth/tokens"
	"github.com/autograde/quickfeed/web/config"
	"github.com/autograde/quickfeed/web/hooks"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

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
		baseURL        = flag.String("service.url", "127.0.0.1", "base service DNS name")
		dbFile         = flag.String("database.file", "qf.db", "database file")
		public         = flag.String("http.public", "public", "path to content to serve")
		httpAddr       = flag.String("http.addr", ":8080", "HTTP listen address")
		withEncryption = flag.Bool("e", false, "encrypt access tokens")
	)
	flag.Parse()

	logger := logq.Zap(true)
	defer logger.Sync()

	db, err := database.NewGormDB(*dbFile, logger)
	if err != nil {
		log.Fatalf("can't connect to database: %v\n", err)
	}

	runner, err := ci.NewDockerCI(logger)
	if err != nil {
		log.Fatalf("failed to set up docker client: %v\n", err)
	}
	defer runner.Close()

	// TODO(vera): replace by requesting and using the installation access token
	// Add application token for external applications (to allow invoking gRPC methods)
	// TODO(meling): this is a temporary solution, and we should find a better way to do this
	// token := os.Getenv("QUICKFEED_AUTH_TOKEN")
	// if len(token) > 16 {
	// 	auth.Add(token, 1)
	// 	log.Println("Added application token")
	// }

	serverConfig := config.NewConfig(*baseURL, *public, *httpAddr)
	logger.Sugar().Debugf("SERVER CONFIG: %+V", serverConfig)
	if *withEncryption {
		if err := serverConfig.ReadKey(false); err != nil {
			log.Fatal(err)
		}
	}

	scmMaker, err := scm.NewSCMMaker()
	if err != nil {
		log.Fatalf("failed to start GitHub app: %v\n", err)
	}
	id, secret := scmMaker.GetIDs()
	logger.Sugar().Debugf("Callback url from config: %s", serverConfig.Endpoints.BaseURL+serverConfig.Endpoints.PortNumber+serverConfig.Endpoints.CallbackURL) // tmp
	// TODO(vera): this part is specific to github, but doesn't have to be.
	// Idea: just pass scm maker to the enpoint handler and let it make config based on the provider.
	scmConfig := oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		Endpoint:     github.Endpoint,
		RedirectURL:  "https://127.0.0.1:8080/auth/github/callback/", // TODO(vera): get from config
	}
	tokenManager, err := tokens.NewTokenManager(db, config.TokenExpirationTime, serverConfig.Secrets.TokenSecret, *httpAddr)
	logger.Sugar().Debugf("Generated token manager, tokens to update: %v", tokenManager.GetTokens()) // tmp
	if err != nil {
		log.Fatalf("failed to make token manager: %v\n", err)
	}

	agService := web.NewAutograderService(logger, db, scmMaker, serverConfig, tokenManager, runner)
	agService.MakeSCMClients()

	apiServer, err := serverConfig.GenerateTLSApi(logger.Sugar(), db, tokenManager)
	if err != nil {
		log.Fatalf("failed to generate TLS grpc API: %v/n", err)
	}
	pb.RegisterAutograderServiceServer(apiServer, agService)

	grpcWebServer := grpcweb.WrapServer(apiServer)
	multiplexer := config.GrpcMultiplexer{
		grpcWebServer,
	}

	router := registerRouter(logger.Sugar(), db, multiplexer, runner, serverConfig, scmConfig, tokenManager)
	srv := &http.Server{
		Handler:      router,
		Addr:         serverConfig.Endpoints.BaseURL + serverConfig.Endpoints.PortNumber, // TODO(vera): fix port/localhost issue
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	// Serve the static handler over TLS
	log.Fatal(srv.ListenAndServeTLS(serverConfig.Paths.CertPath, serverConfig.Paths.CertKeyPath))
}

func registerRouter(logger *zap.SugaredLogger, db database.Database, mux config.GrpcMultiplexer, runner *ci.Docker, qfConfig *config.Config, authConfig oauth2.Config, tm *tokens.TokenManager) *http.ServeMux {
	ghHook := hooks.NewGitHubWebHook(logger, db, runner, qfConfig.Secrets.WebhookSecret)
	router := http.NewServeMux()
	router.Handle("/", mux.MultiplexerHandler(http.StripPrefix("/", http.FileServer(http.Dir("public")))))
	router.HandleFunc(qfConfig.Endpoints.LoginURL, auth.OAuth2Login(logger, db, authConfig, qfConfig.Secrets.CallbackSecret))
	router.HandleFunc(qfConfig.Endpoints.CallbackURL, auth.OAuth2Callback(logger, db, authConfig, tm, qfConfig))
	router.HandleFunc(qfConfig.Endpoints.LogoutURL, auth.OAuth2Logout(logger))
	router.HandleFunc(qfConfig.Endpoints.WebhookURL, hooks.HandleWebhook(ghHook))
	return router
}
