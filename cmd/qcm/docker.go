package main

import (
	"context"
	"fmt"
	"time"

	"github.com/quickfeed/quickfeed/ci"
)

// dockerPingTimeout bounds the check that the Docker daemon answers. A daemon
// that is running answers at once; one that is not makes the connection fail
// at once as well, so the timeout only guards against a hung daemon.
const dockerPingTimeout = 5 * time.Second

// dockerHelp is printed when the Docker daemon cannot be reached.
const dockerHelp = `qcm runs the tests in Docker, exactly as the QuickFeed server does.
Start Docker (Docker Desktop, or the docker daemon) and run the command again.`

// pinger is the part of ci.Docker that reports whether the daemon is
// reachable; see checkDocker.
type pinger interface {
	Ping(context.Context) error
}

// newDockerRunner returns a Docker runner after confirming that the daemon
// answers. Without the confirmation a stopped Docker would be reported as a
// failure of every check or test run attempted, each with its own error
// records, instead of once and clearly before any work is started.
func newDockerRunner(ctx context.Context) (*ci.Docker, error) {
	runner, err := ci.NewDockerCI()
	if err != nil {
		return nil, fmt.Errorf("creating docker client: %w", err)
	}
	if err := checkDocker(ctx, runner); err != nil {
		_ = runner.Close()
		return nil, err
	}
	return runner, nil
}

// checkDocker fails with instructions if the Docker daemon does not answer.
func checkDocker(ctx context.Context, docker pinger) error {
	ctx, cancel := context.WithTimeout(ctx, dockerPingTimeout)
	defer cancel()
	if err := docker.Ping(ctx); err != nil {
		return fmt.Errorf("the Docker daemon is not available: %w\n\n%s", err, dockerHelp)
	}
	return nil
}
