package ci

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Docker is an implementation of the CI interface using Docker.
type Docker struct {
	Endpoint string
	Version  string
}

var containerTimeout = time.Duration(10 * time.Minute)

// Run implements the CI interface. This method blocks until the job has been
// completed or an error occurs, e.g., the context times out.
func (d *Docker) Run(ctx context.Context, job *Job, user string, timeout time.Duration) (string, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return "", err
	}

	create := func() (container.ContainerCreateCreatedBody, error) {
		return cli.ContainerCreate(ctx, &container.Config{
			Image: job.Image,
			Cmd:   []string{"/bin/bash", "-c", strings.Join(job.Commands, "\n")},
		}, nil, nil, user)
	}

	resp, err := create()
	if err != nil {
		// if image not found locally, try to pull it
		if err := pullImage(ctx, cli, job.Image); err != nil {
			return "", err
		}
		resp, err = create()
		if err != nil {
			return "", err
		}
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	// will wait until the container stops
	waitCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	if timeout < 1 {
		timeout = containerTimeout
	}

	select {
	case err := <-errCh:
		// container failed with error; return
		return "", err
	case <-time.After(timeout):
		// force kill container after predefined time interval
		err = cli.ContainerKill(ctx, resp.ID, "SIGTERM")
		userErr := fmt.Sprintf("Container timed out after %v.\nPlease check for infinite loops or other slowness.", timeout)
		return userErr, fmt.Errorf("container timed out after %v: %w", timeout, err)
	case <-waitCh:
		// container finished gracefully; fallthrough
	}

	r, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
	})
	if err != nil {
		return "", err
	}

	var stdout bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, ioutil.Discard, r); err != nil {
		return "", err
	}
	return stdout.String(), nil
}

// pullImage pulls an image from docker hub; this can be slow and should be
// avoided if possible.
func pullImage(ctx context.Context, cli *client.Client, image string) error {
	progress, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer progress.Close()

	_, err = io.Copy(ioutil.Discard, progress)
	return err
}
