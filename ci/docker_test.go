package ci_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/autograde/quickfeed/ci"
	"github.com/docker/docker/client"
)

var docker bool

func init() {
	if os.Getenv("DOCKER_TESTS") != "" {
		docker = true
	}
	cli, err := client.NewEnvClient()
	if err != nil {
		docker = false
	}
	if _, err := cli.Ping(context.Background()); err != nil {
		docker = false
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

	docker := &ci.Docker{}
	out, err := docker.Run(context.Background(), &ci.Job{
		Image:    "golang:latest",
		Commands: []string{script},
	}, "", 0)
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("docker.Run(%#v) = %#v, want %#v", script, out, wantOut)
	}
}

func TestDockerTimeout(t *testing.T) {
	if !docker {
		t.SkipNow()
	}

	const (
		script  = `echo -n "hello," && sleep 10`
		wantOut = `Container timed out after 100ms.
Please check for infinite loops or other slowness.`
	)

	docker := &ci.Docker{}
	out, err := docker.Run(context.Background(), &ci.Job{
		Image:    "golang:latest",
		Commands: []string{script},
	}, "", 100*time.Millisecond)
	if out == "" {
		t.Errorf("docker.Run(%#v) = %#v, want %#v", script, out, wantOut)
	}
	if err == nil {
		t.Errorf("docker.Run(%#v) unexpectedly returned without error", script)
	}
}
