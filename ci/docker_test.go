package ci_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
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
	})
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
	})
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

func TestDockerLogLimit(t *testing.T) {
	// This is just for testing
	t.SkipNow()
	const maxLogSize = 4
	const lastSegmentSize = 5
	logReader := strings.NewReader("want only that some small last thing")
	var stdout bytes.Buffer
	n, err := io.Copy(&stdout, logReader)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("x(%d=%d): %s", n, stdout.Len(), stdout.String())
	if stdout.Len() > maxLogSize {
		all := stdout.String()
		t.Logf("%s ONLY %s", all[0:maxLogSize], all[len(all)-lastSegmentSize:])
	}
	t.Logf("x(%d=%d): %s", n, stdout.Len(), stdout.String())
}
