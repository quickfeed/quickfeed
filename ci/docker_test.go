package ci_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/sh"
	"github.com/quickfeed/quickfeed/qlog"
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
	t.Helper()
	docker, err := ci.NewDockerCI(qlog.Logger(t))
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

func TestDockerMultilineScript(t *testing.T) {
	if !docker {
		t.SkipNow()
	}

	cmds := []string{
		`echo -n "hello world\n"`,
		`echo -n "join my world"`,
	}
	const (
		wantOut    = "hello world\\njoin my world"
		image      = "golang:latest"
		dockerfile = "FROM golang:latest\n WORKDIR /quickfeed"
	)
	docker, closeFn := dockerClient(t)
	defer closeFn()

	out, err := docker.Run(context.Background(), &ci.Job{
		Name:       t.Name() + "-" + qtest.RandomString(t),
		Image:      image,
		Dockerfile: dockerfile,
		Commands:   cmds,
	})
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("docker.Run(%#v) = %#v, want %#v", cmds, out, wantOut)
	}
}

// Note that this test will fail if the content of ./testdata changes.
func TestDockerBindDir(t *testing.T) {
	if !docker {
		t.SkipNow()
	}

	const (
		script     = `ls /quickfeed`
		wantOut    = "run.sh\n" // content of testdata (or /quickfeed inside the container)
		image      = "golang:latest"
		dockerfile = "FROM golang:latest\n WORKDIR /quickfeed" // TODO(meling) this is not needed when using a public image like golang:latest; remove this, and in other tests
	)
	docker, closeFn := dockerClient(t)
	defer closeFn()

	// bindDir is the ./testdata directory to map into /quickfeed.
	bindDir, err := filepath.Abs("./testdata")
	if err != nil {
		t.Fatal(err)
	}
	out, err := docker.Run(context.Background(), &ci.Job{
		Name:       t.Name() + "-" + qtest.RandomString(t),
		Image:      image,
		Dockerfile: dockerfile,
		BindDir:    bindDir,
		Commands:   []string{script},
	})
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("docker.Run(%#v) = %#v, want %#v", script, out, wantOut)
	}
}

// Note that this test will fail if the content of ./testdata changes.
func TestDockerEnvVars(t *testing.T) {
	if !docker {
		t.SkipNow()
	}

	envVars := []string{
		"TESTS=/quickfeed/tests",
		"ASSIGNMENTS=/quickfeed/assignments",
	}
	// check that the default environment variables are accessible from the container
	cmds := []string{
		`echo $TESTS`,
		`echo $ASSIGNMENTS`,
	}

	const (
		wantOut    = "/quickfeed/tests\n/quickfeed/assignments\n"
		image      = "golang:latest"
		dockerfile = "FROM golang:latest\n WORKDIR /quickfeed"
	)
	docker, closeFn := dockerClient(t)
	defer closeFn()

	// dir is the directory to map into /quickfeed in the docker container.
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "tests"), 0o700)
	os.Mkdir(filepath.Join(dir, "assignments"), 0o700)

	out, err := docker.Run(context.Background(), &ci.Job{
		Name:       t.Name() + "-" + qtest.RandomString(t),
		Image:      image,
		Dockerfile: dockerfile,
		BindDir:    dir,
		Env:        envVars,
		Commands:   cmds,
	})
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("docker.Run(%#v) = %#v, want %#v", cmds, out, wantOut)
	}
}

func TestDockerBuild(t *testing.T) {
	if !docker {
		t.SkipNow()
	}

	// TODO(meling) we should avoid using quickfeed:go in tests; we should instead build as a quickfeed:go_test or something to prevent that we overwrite a production image
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
	dockerOut, err := sh.OutputA("docker", "image", "rm", "--force", image, image2)
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

func TestDockerRunAsNonRoot(t *testing.T) {
	if !docker {
		t.SkipNow()
	}

	envVars := []string{
		"TESTS=/quickfeed/tests",
		"ASSIGNMENTS=/quickfeed/assignments",
	}

	const (
		script = `whoami
echo $TESTS
echo $ASSIGNMENTS
id
ls -la /
ls -la
echo "hello" > hello.txt
pwd

`
		wantOut    = "hello world"
		image      = "quickfeed:go"
		dockerfile = `FROM golang:latest
WORKDIR /quickfeed
`
	)
	dockerOut, err := sh.Output("docker image rm --force quickfeed:go")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(dockerOut))

	docker, closeFn := dockerClient(t)
	defer closeFn()

	// dir is the directory to map into /quickfeed in the docker container.
	// dir := t.TempDir()
	const d = "test-quickfeed"
	os.Mkdir(d, 0o700)
	dir, err := filepath.Abs(d)
	if err != nil {
		t.Fatal(err)
	}
	os.Mkdir(filepath.Join(dir, "tests"), 0o700)
	os.Mkdir(filepath.Join(dir, "assignments"), 0o700)

	out, err := docker.Run(context.Background(), &ci.Job{
		Name:       t.Name() + "-" + qtest.RandomString(t),
		Image:      image,
		Dockerfile: dockerfile,
		BindDir:    dir,
		Env:        envVars,
		Commands:   []string{script},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(out)

	// if out != wantOut {
	// 	t.Errorf("docker.Run(%#v) = %#v, want %#v", script, out, wantOut)
	// }
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
