package ci_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/log"
	"github.com/docker/docker/client"
)

var docker bool

func init() {
	if os.Getenv("DOCKER_TESTS") != "" {
		docker = true
	}
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		docker = false
	}
	if _, err := cli.Ping(context.Background()); err != nil {
		docker = false
	}
}

func dockerClient(t *testing.T) (*ci.Docker, func()) {
	docker, err := ci.NewDockerCI(log.Zap(true))
	if err != nil {
		t.Fatalf("Failed to set up docker client: %v", err)
	}
	return docker, func() { _ = docker.Close() }
}

func TestDocker(t *testing.T) {
	if !docker {
		t.SkipNow()
	}

	const (
		script     = `echo -n "hello world"`
		wantOut    = "hello world"
		image      = "golang:latest"
		dockerfile = "FROM golang:latest\n WORKDIR /quickfeed"
	)
	docker, closeFn := dockerClient(t)
	defer closeFn()

	out, err := docker.Run(context.Background(), &ci.Job{
		Name:       t.Name() + "-" + qtest.RandomString(t),
		Image:      image,
		Dockerfile: dockerfile,
		Commands:   []string{script},
	})
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("docker.Run(%#v) = %#v, want %#v", script, out, wantOut)
	}
}

func TestDockerBuild(t *testing.T) {
	if !docker {
		t.SkipNow()
	}

	const (
		script     = `echo -n "hello world"`
		wantOut    = "hello world"
		image      = "quickfeed:go"
		image2     = "golang:latest"
		dockerfile = `FROM golang:latest
		RUN apt update && apt install -y git bash build-essential && rm -rf /var/lib/apt/lists/*
		RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.41.1
		WORKDIR /quickfeed`
	)
	cmd := exec.Command("docker", "image", "rm", "--force", image, image2)
	dockerOut, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(dockerOut))

	docker, closeFn := dockerClient(t)
	defer closeFn()

	// To build an image, we need a job with both image name
	// and Dockerfile content.
	out, err := docker.Run(context.Background(), &ci.Job{
		Name:       t.Name() + "-" + qtest.RandomString(t),
		Image:      image,
		Dockerfile: dockerfile,
		Commands:   []string{script},
	})
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("docker.Run(%#v) = %#v, want %#v", script, out, wantOut)
	}
}

func TestDockerPull(t *testing.T) {
	if !docker {
		t.SkipNow()
	}

	const (
		script  = `python -c "print('Hello, world!')"`
		wantOut = "Hello, world!\n"
		image   = "python:latest"
	)
	cmd := exec.Command("docker", "image", "rm", "--force", image)
	dockerOut, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(dockerOut))

	docker, closeFn := dockerClient(t)
	defer closeFn()

	// To pull an image, we need only a job with image name;
	// no Dockerfile content should be provided when pulling.
	out, err := docker.Run(context.Background(), &ci.Job{
		Name:     t.Name() + "-" + qtest.RandomString(t),
		Image:    image,
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
		script     = `echo -n "hello," && sleep 10`
		wantOut    = `Container timeout. Please check for infinite loops or other slowness.`
		image      = "golang:latest"
		dockerfile = "FROM golang:latest\n WORKDIR /quickfeed"
	)

	// Note that the timeout value below is sensitive to startup time of the container.
	// If the timeout is too short, the Run() call may not reach the ContainerWait() call.
	// Hence, if this test fails, you may try to increase the timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()

	docker, closeFn := dockerClient(t)
	defer closeFn()

	out, err := docker.Run(ctx, &ci.Job{
		Name:       t.Name() + "-" + qtest.RandomString(t),
		Image:      image,
		Dockerfile: dockerfile,
		Commands:   []string{script},
	})
	t.Log("Expecting ERROR line above; not test failure")
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

func TestDockerOpenFileDescriptors(t *testing.T) {
	// This is mainly for debugging the 'too many open file descriptors' issue
	if !docker {
		t.SkipNow()
	}

	const (
		script        = `echo -n "hello, " && sleep 2 && echo -n "world!"`
		wantOut       = "hello, world!"
		image         = "golang:latest"
		numContainers = 5
		dockerfile    = "FROM golang:latest\n WORKDIR /quickfeed"
	)
	docker, closeFn := dockerClient(t)
	defer closeFn()

	errCh := make(chan error, numContainers)
	for i := 0; i < numContainers; i++ {
		go func(j int) {
			name := fmt.Sprintf(t.Name()+"-%d-%s", j, qtest.RandomString(t))
			out, err := docker.Run(context.Background(), &ci.Job{
				Name:       name,
				Image:      image,
				Dockerfile: dockerfile,
				Commands:   []string{script},
			})
			if err != nil {
				errCh <- err
			}
			if out != wantOut {
				t.Errorf("docker.Run(%#v) = %#v, want %#v", script, out, wantOut)
			}
			errCh <- nil
		}(i)
	}
	afterContainersStarted := countOpenFiles(t)

	for i := 0; i < numContainers; i++ {
		err := <-errCh
		if err != nil {
			t.Fatal(err)
		}
	}
	close(errCh)
	afterContainersFinished := countOpenFiles(t)
	if afterContainersFinished > afterContainersStarted {
		t.Errorf("finished %d > started %d", afterContainersFinished, afterContainersStarted)
	}
}

func countOpenFiles(t *testing.T) int {
	t.Helper()
	out, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("lsof -p %v", os.Getpid())).Output()
	if err != nil {
		t.Fatal(err)
	}
	return bytes.Count(out, []byte("\n"))
}
