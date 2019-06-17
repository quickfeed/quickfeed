package envoy

import (
	"context"
	"os/exec"

	"github.com/docker/docker/api/types/filters"
	"go.uber.org/zap"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// StartEnvoy creates a Docker API client. If an envoy container is not running,
// it will be started from an image. If no image exists, it will pull an Envoy
// image from docker and build it with options from envoy.yaml.
//TODO(meling) since this runs in a separate goroutine it is actually bad practice
//to Panic or Fatal on error, since other goroutines may not exit cleanly.
//Instead it would be better to return an error and run synchronously.
func StartEnvoy(l *zap.Logger) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
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
		//TODO(meling) use docker api to build image: "docker build -t ag_envoy -f ./envoy/envoy.Dockerfile ."
		out, err := exec.Command("/bin/sh", "./envoy/envoy.sh", "build").Output()
		if err != nil {
			l.Fatal("failed to execute bash script", zap.Error(err))
		}
		l.Debug("envoy.sh build", zap.String("output", string(out)))
	}
	l.Info("starting Envoy container...")
	//TODO(meling) use docker api to run image: "docker run --name=envoy -p 8080:8080 --net=host ag_envoy"
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
		if img.RepoTags[0] == "ag_envoy:latest" {
			l.Debug("found Envoy image")
			return true
		}
	}
	return false
}
