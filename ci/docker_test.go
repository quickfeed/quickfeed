package ci_test

import (
	"context"
	"errors"
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
	}, "")
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
		wantOut = `Container timeout. Please check for infinite loops or other slowness.`
	)

	// Note that the timeout value below is sensitive to startup time of the container.
	// If the timeout is too short, the Run() call may not reach the ContainerWait() call.
	// Hence, if this test fails, you may try to increase the timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	docker := &ci.Docker{}
	out, err := docker.Run(ctx, &ci.Job{
		Image:    "golang:latest",
		Commands: []string{script},
	}, "")
	if out != wantOut {
		t.Errorf("docker.Run(%#v) = %#v, want %#v", script, out, wantOut)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("docker.Run(%#v) = %#v, want %#v", script, err.Error(), context.DeadlineExceeded.Error())
	}
	if err == nil {
		t.Errorf("docker.Run(%#v) unexpectedly returned without error", script)
	}
}
