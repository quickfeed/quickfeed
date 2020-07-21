package ci

import (
	"bytes"
	"context"
	"errors"
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

	// wait until the container stops or context times out.
	_, err = cli.ContainerWait(ctx, resp.ID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			timeout := time.Duration(1 * time.Second)
			// ContainerStop seems to work better than ContainerKill in my tests.
			stopErr := cli.ContainerStop(context.Background(), resp.ID, &timeout)
			if stopErr != nil {
				// if we failed to stop, report it up the chain
				err = stopErr
			}
			userMsg := "Container timeout. Please check for infinite loops or other slowness."
			return userMsg, err
		}
		return "", err
	}

	logReader, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
	})
	if err != nil {
		return "", err
	}

	var stdout bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, ioutil.Discard, logReader); err != nil {
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
