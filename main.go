package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/autograde/quickfeed/ci"
	//"github.com/autograde/quickfeed/envoy"
	"github.com/autograde/quickfeed/web"
	"github.com/autograde/quickfeed/web/auth"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
		baseURL    = flag.String("service.url", "", "base service DNS name")
		dbFile     = flag.String("database.file", "qf.db", "database file")
		public     = flag.String("http.public", "public", "path to content to serve")
		httpAddr   = flag.String("http.addr", ":8081", "HTTP listen address")
		grpcAddr   = flag.String("grpc.addr", ":9090", "gRPC listen address")
		scriptPath = flag.String("script.path", "ci/scripts", "path to continuous integration scripts")
		fake       = flag.Bool("provider.fake", false, "enable fake provider")
	)
	flag.Parse()

	cfg := zap.NewDevelopmentConfig()
	// database logging is only enabled if the LOGDB environment variable is set
	cfg = database.GormLoggerConfig(cfg)
	// add colorization
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// we only want stack trace enabled for panic level and above
	logger, err := cfg.Build(zap.AddStacktrace(zapcore.PanicLevel))
	if err != nil {
		log.Fatalf("can't initialize logger: %v\n", err)
	}
	defer logger.Sync()

	db, err := database.NewGormDB("sqlite3", *dbFile, database.NewGormLogger(logger))
	if err != nil {
		log.Fatalf("can't connect to database: %v\n", err)
	}
	defer func() {
		if dbErr := db.Close(); dbErr != nil {
			log.Printf("error closing database: %v\n", dbErr)
		}
	}()

	// start envoy in a docker container; fetch envoy docker image if necessary
	//go envoy.StartEnvoy(logger)

	// holds references for activated providers for current user token
	scms := auth.NewScms()
	bh := web.BaseHookOptions{
		BaseURL: *baseURL,
		Secret:  os.Getenv("WEBHOOK_SECRET"),
	}

	runner, err := ci.NewDockerCI()
	if err != nil {
		log.Fatalf("failed to set up docker client: %v\n", err)
	}
	defer runner.Close()

	agService := web.NewAutograderService(logger, db, scms, bh, runner)
	go web.New(agService, *public, *httpAddr, *scriptPath, *fake)

	lis, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		log.Fatalf("failed to start tcp listener: %v\n", err)
	}

	//  Experimental
	accessControl := AccessControl{db}
	opt := grpc.UnaryInterceptor(accessControl.UserVerifier)
	//opt := grpc.UnaryInterceptor(pb.Interceptor(logger))
	grpcServer := grpc.NewServer(opt)

	// Create a HTTP server for prometheus.
	httpServer := &http.Server{
		Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
		Addr:    fmt.Sprintf("0.0.0.0:%d", 9097),
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatal("Unable to start a http server.")
		}
	}()

	pb.RegisterAutograderServiceServer(grpcServer, agService)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to start grpc server: %v\n", err)
	}
}

// TODO: Move somewhere else that makes sense
type AccessControl struct {
	db *database.GormDB
}

// UserVerifier looks up a user (passed by AccessToken) in the database, and translates it into a User ID
// Modifies the context to include the User ID to be passed along to the actual gRPC method.
func (pc *AccessControl) UserVerifier(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response interface{}, err error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("could not grab metadata from context")
	}

	user, err := pc.db.GetUserByAccessToken(meta.Get("user")[0])
	if err != nil {
		return nil, errors.New("could not associate token with a user")
	}

	meta.Set("user", strconv.FormatUint(user, 10))

	edited := metadata.NewOutgoingContext(ctx, meta)

	return handler(edited, req)
}
