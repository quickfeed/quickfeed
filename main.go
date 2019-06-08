package main

import (
	"flag"
	"log"
	"mime"
	"net"
	"os"
	"path/filepath"

	"github.com/autograde/aguis/envoy"
	"github.com/autograde/aguis/web"
	"go.uber.org/zap"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"

	"github.com/autograde/aguis/web/grpcservice"

	http "github.com/autograde/aguis/web/webserver"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"google.golang.org/grpc"
)

func init() {
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
}

func main() {
	var (
		httpAddr  = flag.String("http.addr", ":8081", "HTTP listen address")
		public    = flag.String("http.public", "public", "directory to server static files from")
		ciScripts = flag.String("script.path", "ci/scripts", "Directory with docker CI scripts")
		dbFile    = flag.String("database.file", tempFile("ag.db"), "database file")
		baseURL   = flag.String("service.url", "localhost", "service base url")
		fake      = flag.Bool("provider.fake", false, "enable fake provider")
		grpcAddr  = flag.String("grpc.addr", ":9090", "gRPC listen address")
	)
	flag.Parse()

	//TODO(meling) make dev flag to switch between dev/production logging
	// lg, err := zap.NewProduction()
	lg, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize logger: %v\n", err)
	}
	defer lg.Sync()

	//TODO(meling) how to connect the main logger with the GormLogger; now they are independent;
	db, err := database.NewGormDB("sqlite3", *dbFile, database.NewGormLogger())
	if err != nil {
		log.Fatalf("can't connect to database: %v\n", err)
	}
	defer func() {
		if dbErr := db.Close(); dbErr != nil {
			log.Printf("error closing database: %v\n", dbErr)
		}
	}()

	// start envoy in a docker container; fetch envoy docker image if necessary
	go envoy.StartEnvoy(lg)

	// holds references for activated providers for current user token
	scms := web.NewScms()
	bh := web.BaseHookOptions{
		BaseURL: *baseURL,
		Secret:  os.Getenv("WEBHOOK_SECRET"),
	}
	go http.NewWebServer(db, bh, lg, *public, *httpAddr, *baseURL, *fake, *ciScripts, scms)

	lis, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		log.Fatalf("failed to start tcp listener: %v\n", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAutograderServiceServer(grpcServer, grpcservice.NewAutograderService(db, scms, bh))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to start grpc server: %v\n", err)
	}
}

func tempFile(name string) string {
	return filepath.Join(os.TempDir(), name)
}
