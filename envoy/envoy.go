package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type EnvoyConfig struct {
	Domain   string
	GRPCPort string
	HTTPPort string
}

func newEnvoyConfig(domain, GRPCPort, HTTPPort string) *EnvoyConfig {
	return &EnvoyConfig{
		Domain:   domain,
		GRPCPort: GRPCPort,
		HTTPPort: HTTPPort,
	}
}

//go:embed envoy.tmpl
var envoyTmpl embed.FS

// createEnvoyConfig creates the envoy.yaml config file.
func createEnvoyConfig(path string, data *EnvoyConfig) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	tmpl, err := template.ParseFS(envoyTmpl, "envoy.tmpl")
	if err != nil {
		return err
	}

	if err = tmpl.ExecuteTemplate(f, "envoy", data); err != nil {
		return err
	}
	return nil
}

var (
	genConfig              bool
	runEnvoy               bool
	envoyConfigPath        string
	_, pwd, _, _           = runtime.Caller(0)
	codePath               = path.Join(path.Dir(pwd), "..")
	env                    = filepath.Join(codePath, ".env")
	defaultEnvoyConfigPath = filepath.Join(path.Dir(pwd), "envoy.yaml")
)

// loadConfigEnv loads the  envoy config from the environment variables.
// It will not override a variable that already exists.
// Consider the .env file to set development vars or defaults.
func loadConfigEnv() (*EnvoyConfig, error) {
	err := godotenv.Load(env)
	if err != nil {
		return nil, err
	}
	return newEnvoyConfig(os.Getenv("DOMAIN"), os.Getenv("GRPC_PORT"), os.Getenv("HTTP_PORT")), nil
}

func main() {
	flag.BoolVar(&genConfig, "genconfig", false, "generate envoy config")
	flag.StringVar(&envoyConfigPath, "path", defaultEnvoyConfigPath, "envoy config path")
	flag.BoolVar(&runEnvoy, "run", false, "run envoy container")
	flag.Parse()

	config, err := loadConfigEnv()
	if err != nil {
		log.Fatal(err)
	}

	switch {
	case genConfig:
		err := createEnvoyConfig(envoyConfigPath, config)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("envoy config file created at", envoyConfigPath)
	case runEnvoy:
		// TODO: refactor startEnvoy
	default:
		fmt.Println("unknown command.")
	}
}

// StartEnvoy creates a Docker API client. If an envoy container is not running,
// it will be started from an image. If no image exists, it will pull an Envoy
// image from docker and build it with options from envoy.yaml.
// TODO(meling) since this runs in a separate goroutine it is actually bad practice
// to Panic or Fatal on error, since other goroutines may not exit cleanly.
// Instead it would be better to return an error and run synchronously.
func StartEnvoy(l *zap.Logger) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		l.Fatal("failed to start docker client", zap.Error(err))
	}

	// removes all stopped containers
	_, err = cli.ContainersPrune(ctx, filters.Args{})
	if err != nil {
		//
		l.Info("failed to prune unused containers", zap.Error(err))
	}

	// check for existing Envoy containers
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		l.Fatal("failed to retrieve docker container list", zap.Error(err))
	}
	for _, container := range containers {
		if container.Names[0] == "/envoy" {
			l.Info("Envoy container is already running")
			return
		}
	}

	if !hasEnvoyImage(ctx, l, cli) {
		// if there is no active Envoy image, we build it
		l.Info("building Envoy image...")
		// TODO(meling) use docker api to build image: "docker build -t ag_envoy -f ./envoy/Dockerfile ."
		out, err := exec.Command("/bin/sh", "./envoy/envoy.sh", "build").Output()
		if err != nil {
			l.Fatal("failed to execute bash script", zap.Error(err))
		}
		l.Debug("envoy.sh build", zap.String("output", string(out)))
	}
	l.Info("starting Envoy container...")
	// TODO(meling) use docker api to run image: "docker run --name=envoy -p 8080:8080 --net=host ag_envoy"
	out, err := exec.Command("/bin/sh", "./envoy/envoy.sh").Output()
	if err != nil {
		l.Fatal("failed to execute bash script", zap.Error(err))
	}
	l.Debug("envoy.sh", zap.String("output", string(out)))
}

// hasEnvoyImage returns true if the docker client has the latest Envoy image.
func hasEnvoyImage(ctx context.Context, l *zap.Logger, cli *client.Client) bool {
	l.Debug("no running Envoy container found")
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		l.Fatal("failed to retrieve docker image list", zap.Error(err))
	}
	l.Debug("checking for Autograder's Envoy image")
	for _, img := range images {
		l.Debug("found image", zap.Strings("repo", img.RepoTags))
		if len(img.RepoTags) > 0 && img.RepoTags[0] == "ag_envoy:latest" {
			l.Debug("found Envoy image")
			return true
		}
	}
	return false
}
