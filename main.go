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

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"

	"github.com/autograde/aguis/logger"
	"github.com/autograde/aguis/web/grpcservice"

	http "github.com/autograde/aguis/web/webserver"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
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
		httpAddr = flag.String("http.addr", ":8081", "HTTP listen address")
		public   = flag.String("http.public", "public", "directory to server static files from")

		buildscripts = flag.String("script.path", "buildscripts", "Directory with docker build scripts")

		dbFile = flag.String("database.file", tempFile("ag.db"), "database file")

		baseURL = flag.String("service.url", "localhost", "service base url")

		fake = flag.Bool("provider.fake", false, "enable fake provider")

		grpcAddr = flag.String("grpc.addr", ":9090", "gRPC listen address")
	)
	flag.Parse()

	l := logrus.New()
	l.Formatter = logger.NewDevFormatter(l.Formatter)

	db, err := database.NewGormDB("sqlite3", *dbFile, database.Logger{Logger: l})
	if err != nil {
		l.WithError(err).Fatal("could not connect to db")
	}
	defer func() {
		if dbErr := db.Close(); dbErr != nil {
			l.WithError(dbErr).Warn("error closing database")
		}
	}()

	/* Will start envoy in a docker container from image unless it is already running,
	if there is no image, will pull the envoy docker image, build a local envoy image with options from envoy.yaml
	 and start a container from the new image */
	go func() {
		envoy.StartEnvoy()
	}()

	// holds references for activated providers for current user token
	scms := make(map[string]scm.SCM)
	bh := web.BaseHookOptions{
		BaseURL: *baseURL,
		Secret:  os.Getenv("WEBHOOK_SECRET"),
	}
	go func() {
		http.NewWebServer(db, bh, l, *public, *httpAddr, *baseURL, *fake, *buildscripts, scms)
	}()

	lis, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAutograderServiceServer(grpcServer, grpcservice.NewAutograderService(db, scms, bh))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

func tempFile(name string) string {
	return filepath.Join(os.TempDir(), name)
}
