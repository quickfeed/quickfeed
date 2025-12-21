package ci_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/sh"
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
	docker, err := ci.NewDockerCI(qtest.Logger(t))
	if err != nil {
		t.Fatalf("Failed to set up docker client: %v", err)
	}
	return docker, func() { _ = docker.Close() }
}

// deleteDockerImages deletes the given images.
// Used for tests that need fresh start, e.g., for pulling or building and image.
func deleteDockerImages(t *testing.T, images ...string) {
	t.Helper()
	args := append([]string{"image", "rm", "--force"}, images...)
	dockerOut, err := sh.OutputA("docker", args...)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(dockerOut))
}

func TestDocker(t *testing.T) {
	if !docker {
		t.SkipNow()
	}

	const (
		script  = `echo -n "hello world"`
		wantOut = "hello world"
		image   = "golang:latest"
	)
	docker, closeFn := dockerClient(t)
	defer closeFn()

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

func TestDockerMultilineScript(t *testing.T) {
	if !docker {
		t.SkipNow()
	}

	cmds := []string{
		`echo -n "hello world\n"`,
		`echo -n "join my world"`,
	}
	const (
		wantOut = "hello world\\njoin my world"
		image   = "golang:latest"
	)
	docker, closeFn := dockerClient(t)
	defer closeFn()

	out, err := docker.Run(context.Background(), &ci.Job{
		Name:     t.Name() + "-" + qtest.RandomString(t),
		Image:    image,
		Commands: cmds,
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
		script  = `ls /quickfeed`
		wantOut = "Dockerfile\nassignments\nrun.sh\ntests\n" // content of testdata (or /quickfeed inside the container)
		image   = "golang:latest"
	)
	docker, closeFn := dockerClient(t)
	defer closeFn()

	// bindDir is the ./testdata directory to map into /quickfeed.
	bindDir, err := filepath.Abs("./testdata")
	if err != nil {
		t.Fatal(err)
	}
	out, err := docker.Run(context.Background(), &ci.Job{
		Name:     t.Name() + "-" + qtest.RandomString(t),
		Image:    image,
		BindDir:  bindDir,
		Commands: []string{script},
	})
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("docker.Run(%#v) = %#v, want %#v", script, out, wantOut)
	}
}

func TestDockerEnvVars(t *testing.T) {
	if !docker {
		t.SkipNow()
	}

	envVars := []string{
		"TESTS=/quickfeed/tests",
		"ASSIGNMENTS=/quickfeed/assignments",
		"SUBMITTED=/quickfeed/submitted",
	}
	// check that the default environment variables are accessible from the container
	cmds := []string{
		`echo $TESTS`,
		`echo $ASSIGNMENTS`,
		`echo $SUBMITTED`,
	}

	const (
		wantOut = "/quickfeed/tests\n/quickfeed/assignments\n/quickfeed/submitted\n"
		image   = "golang:latest"
	)
	docker, closeFn := dockerClient(t)
	defer closeFn()

	// dir is the directory to map into /quickfeed in the docker container.
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "tests"), 0o700); err != nil {
		t.Error(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "assignments"), 0o700); err != nil {
		t.Error(err)
	}

	out, err := docker.Run(context.Background(), &ci.Job{
		Name:     t.Name() + "-" + qtest.RandomString(t),
		Image:    image,
		BindDir:  dir,
		Env:      envVars,
		Commands: cmds,
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
	deleteDockerImages(t, image, image2)

	docker, closeFn := dockerClient(t)
	defer closeFn()

	// To build an image, we need a job with both image name
	// and Dockerfile content.
	out, err := docker.Run(context.Background(), &ci.Job{
		Name:  t.Name() + "-" + qtest.RandomString(t),
		Image: image,
		BuildContext: map[string]string{
			ci.Dockerfile: dockerfile,
		},
		Commands: []string{script},
	})
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("docker.Run(%#v) = %#v, want %#v", script, out, wantOut)
	}
}

func TestDockerBuildRebuild(t *testing.T) {
	if !docker {
		t.SkipNow()
	}

	const (
		script     = `echo -n "hello world"`
		script2    = `echo -n "hello quickfeed"`
		wantOut    = "hello world"
		wantOut2   = "hello quickfeed"
		image      = "dat320:latest"
		image2     = "golang:latest"
		dockerfile = `FROM golang:latest
RUN apt update && apt install -y git bash build-essential && rm -rf /var/lib/apt/lists/*
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.41.1
WORKDIR /quickfeed`
		dockerfile2 = `FROM golang:latest
RUN apt update && apt install -y git bash build-essential && rm -rf /var/lib/apt/lists/*
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.42.1
WORKDIR /quickfeed`
	)

	docker, closeFn := dockerClient(t)
	defer closeFn()

	out, err := docker.Run(context.Background(), &ci.Job{
		Name:  t.Name() + "-" + qtest.RandomString(t),
		Image: image,
		BuildContext: map[string]string{
			ci.Dockerfile: dockerfile,
		},
		Commands: []string{script},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out != wantOut {
		t.Errorf("docker.Run(%#v) = %#v, want %#v", script, out, wantOut)
	}

	out2, err := docker.Run(context.Background(), &ci.Job{
		Name:  t.Name() + "-" + qtest.RandomString(t),
		Image: image,
		BuildContext: map[string]string{
			ci.Dockerfile: dockerfile2,
		},
		Commands: []string{script2},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out2 != wantOut2 {
		t.Errorf("docker.Run(%#v) = %#v, want %#v", script2, out2, wantOut2)
	}
}

func TestDockerBuildWithModuleCache(t *testing.T) {
	if !docker {
		t.SkipNow()
	}

	const (
		// List module files and verify go.mod content, then check that go mod download works
		script = `cat go.mod && go mod download && echo "modules downloaded successfully"`
		image  = "testmod:latest"
		goMod  = `module github.com/quickfeed/build-context

go 1.24

require github.com/relab/container v0.0.0-20251028224705-baa7b7c5c895 // indirect
`
		goSum = `github.com/relab/container v0.0.0-20251028224705-baa7b7c5c895 h1:IiS1KzQwmZsL5fnrvSpdMypSKqu7wcqSO7DGjWTr6bs=
github.com/relab/container v0.0.0-20251028224705-baa7b7c5c895/go.mod h1:oLZXG1NirJWzF2fMEeMUC6OLiMn6RtCChzUz1jtF/qs=
`
		dockerfile = `FROM golang:latest
WORKDIR /quickfeed
COPY go.mod go.sum ./
RUN go mod download
`
	)
	deleteDockerImages(t, image)

	docker, closeFn := dockerClient(t)
	defer closeFn()

	out, err := docker.Run(context.Background(), &ci.Job{
		Name:  t.Name() + "-" + qtest.RandomString(t),
		Image: image,
		BuildContext: map[string]string{
			ci.Dockerfile: dockerfile,
			"go.mod":      goMod,
			"go.sum":      goSum,
		},
		Commands: []string{script},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify that go.mod content is present in output
	if !strings.Contains(out, "github.com/quickfeed/build-context") {
		t.Errorf("expected go.mod content in output, got: %s", out)
	}
	// Verify that modules were downloaded successfully
	if !strings.Contains(out, "modules downloaded successfully") {
		t.Errorf("expected successful module download message in output, got: %s", out)
	}

	deleteDockerImages(t, image)
}

func TestDockerRunAsNonRoot(t *testing.T) {
	if !docker {
		t.SkipNow()
	}

	envVars := []string{
		"HOME=/quickfeed",
		"TESTS=/quickfeed/tests",
		"ASSIGNMENTS=/quickfeed/assignments",
	}
	wantOut := []string{
		"HOME: /quickfeed",
		"/quickfeed/tests",
		"/quickfeed/.cache/go-build",
		"=== RUN   TestX",
		"x_test.go:10: hallo",
		"--- PASS: TestX ",
	}

	const (
		script = `echo "HOME: $HOME"
echo "hello" > hello.txt
cd tests
cat << EOF > go.mod
module tests

go 1.19
EOF
pwd
go env GOCACHE
go test -v
`
		image      = "quickfeed:go"
		dockerfile = `FROM golang:latest
WORKDIR /quickfeed
`
	)

	docker, closeFn := dockerClient(t)
	defer closeFn()

	// dir is the directory to map into /quickfeed in the docker container.
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "tests"), 0o700); err != nil {
		t.Error(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "assignments"), 0o700); err != nil {
		t.Error(err)
	}

	xTestGo, err := os.ReadFile("testdata/tests/x_test.go")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "tests", "x_test.go"), xTestGo, 0o600); err != nil {
		t.Fatal(err)
	}

	out, err := docker.Run(context.Background(), &ci.Job{
		Name:  t.Name() + "-" + qtest.RandomString(t),
		Image: image,
		BuildContext: map[string]string{
			ci.Dockerfile: dockerfile,
		},
		BindDir:  dir,
		Env:      envVars,
		Commands: []string{script},
	})
	if err != nil {
		t.Fatal(err)
	}

	checkOwner(t, filepath.Join(dir, "hello.txt"))
	for _, line := range wantOut {
		if !strings.Contains(out, line) {
			t.Errorf("Expected %q not found in output: %q", line, out)
		}
	}

	if t.Failed() {
		// Print output from container.
		t.Log(out)
		// Print output from local filesystem (non-container).
		out2, err := sh.Output("ls -l " + dir)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(out2)
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
	deleteDockerImages(t, image)

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

func TestDockerPullFromNonDockerHubRepositories(t *testing.T) {
	if !docker {
		t.SkipNow()
	}
	const (
		script  = `echo "Hello, world!"`
		wantOut = "Hello, world!\n"
		image   = "mcr.microsoft.com/dotnet/sdk:6.0"
	)
	deleteDockerImages(t, image)

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
		script  = `echo -n "hello," && sleep 10`
		wantOut = `Container timeout. Please check for infinite loops or other slowness.`
		image   = "golang:latest"
	)

	// Note that the timeout value below is sensitive to startup time of the container.
	// If the timeout is too short, the Run() call may not reach the ContainerWait() call.
	// Hence, if this test fails, you may try to increase the timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()

	docker, closeFn := dockerClient(t)
	defer closeFn()

	out, err := docker.Run(ctx, &ci.Job{
		Name:     t.Name() + "-" + qtest.RandomString(t),
		Image:    image,
		Commands: []string{script},
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
	)
	docker, closeFn := dockerClient(t)
	defer closeFn()

	errCh := make(chan error, numContainers)
	for i := 0; i < numContainers; i++ {
		go func(j int) {
			name := fmt.Sprintf(t.Name()+"-%d-%s", j, qtest.RandomString(t))
			out, err := docker.Run(context.Background(), &ci.Job{
				Name:     name,
				Image:    image,
				Commands: []string{script},
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
