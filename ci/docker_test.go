package ci_test

import (
	"context"
	"os"
	"testing"

	"github.com/autograde/aguis/ci"
	"github.com/docker/docker/client"
)

var docker bool
var host, version string

func init() {
	host = envString("DOCKER_HOST", "http://localhost:4243")
	version = envString("DOCKER_VERSION", "1.30")

	if os.Getenv("DOCKER_TESTS") != "" {
		docker = true
	}

	cli, err := client.NewClient(host, version, nil, nil)
	if err != nil {
		docker = false
	}
	if _, err := cli.Ping(context.Background()); err != nil {
		docker = false
	}
}

func newDockerCI() *ci.Docker {
	return &ci.Docker{
		Endpoint: host,
		Version:  version,
	}
}

func TestDocker(t *testing.T) {
	if !docker {
		t.SkipNow()
	}

	const (
		script  = `echo -n "hello world"`
		wantOut = "hello world"
	)

	docker := newDockerCI()
	out, err := docker.Run(context.Background(), &ci.Job{
		Image:    "golang:1.8.3",
		Commands: []string{script},
	})
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("have %#v want %#v", out, wantOut)
	}
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
