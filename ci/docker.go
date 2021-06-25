package ci

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/autograde/quickfeed/kit/score"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"go.uber.org/zap"
)

var (
	containerTimeout = time.Duration(10 * time.Minute)
	maxToScan        = 1_000_000 // bytes
	maxLogSize       = 30_000    // bytes
	lastSegmentSize  = 1_000     // bytes
)

// Docker is an implementation of the CI interface using Docker.
type Docker struct {
	client *client.Client
	logger *zap.SugaredLogger
}

// NewDockerCI returns a runner to run CI tests.
func NewDockerCI(logger *zap.Logger) (*Docker, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &Docker{
		client: cli,
		logger: logger.Sugar(),
	}, nil
}

// Close ensures that the docker client is closed.
func (d *Docker) Close() error {
	if d.logger != nil {
		d.logger.Sync()
	}
	return d.client.Close()
}

// Run implements the CI interface. This method blocks until the job has been
// completed or an error occurs, e.g., the context times out.
func (d *Docker) Run(ctx context.Context, job *Job) (string, error) {
	if d.client == nil {
		return "", fmt.Errorf("cannot run job: %s; docker client not initialized", job.Name)
	}

	create := func() (container.ContainerCreateCreatedBody, error) {
		return d.client.ContainerCreate(ctx, &container.Config{
			Image: job.Image,
			Cmd:   []string{"/bin/bash", "-c", strings.Join(job.Commands, "\n")},
		}, nil, nil, job.Name)
	}

	resp, err := create()
	if err != nil {
		d.logger.Debugf("Failed to create container image '%s' for %s: %v\n", job.Image, job.Name, err)
		// if image not found locally, try to pull it
		if err := pullImage(ctx, d.client, job.Image); err != nil {
			return "", err
		}
		resp, err = create()
		if err != nil {
			return "", err
		}
	}

	if err := d.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	// wait until the container stops or context times out.
	_, err = d.client.ContainerWait(ctx, resp.ID)
	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			return "", err
		}

		// stop runaway container whose deadline was exceeded
		timeout := time.Duration(1 * time.Second)
		stopErr := d.client.ContainerStop(context.Background(), resp.ID, &timeout)
		if stopErr != nil {
			return "", stopErr
		}

		// remove the docker container (when stopped due to timeout) to prevent too many open files
		rmErr := d.client.ContainerRemove(context.Background(), resp.ID, types.ContainerRemoveOptions{})
		if rmErr != nil {
			return "", rmErr
		}

		// return message to user to be shown in the results log
		return "Container timeout. Please check for infinite loops or other slowness.", err
	}

	// extract the logs before removing the container below
	logReader, err := d.client.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
	})
	if err != nil {
		return "", err
	}

	// remove the container when finished to prevent too many open files
	err = d.client.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
	if err != nil {
		return "", err
	}

	var stdout bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, ioutil.Discard, logReader); err != nil {
		return "", err
	}
	if stdout.Len() > maxLogSize+lastSegmentSize {
		// converting to string here;
		// could be done more efficiently using stdout.Truncate(maxLogSize)
		// but then we wouldn't get the last part
		all := stdout.String()
		// find the last full line to keep before the truncate point
		startMiddleSegment := strings.LastIndex(all[0:maxLogSize], "\n") + 1
		// find the last full line to truncate and scan for score lines, before the last segment to output
		startLastSegment := strings.LastIndex(all[0:len(all)-lastSegmentSize], "\n") + 1

		middleSegment := all[startMiddleSegment:startLastSegment]
		// score lines will normally replace this string, unless too much output
		scoreLines := "too much output data to scan (skipping; fix your code)"
		// only scan if middle segment is less than maxToScan
		if len(middleSegment) < maxToScan {
			// find score lines in the middle segment that otherwise gets truncated
			scoreLines = findScoreLines(middleSegment)
		}
		return all[0:startMiddleSegment] + scoreLines + `

		...
		truncated output
		...

		` + all[startLastSegment:], nil
	}
	return stdout.String(), nil
}

func findScoreLines(lines string) string {
	scoreLines := make([]string, 0)
	for _, line := range strings.Split(lines, "\n") {
		// check if line has expected JSON score string
		if score.HasPrefix(line) {
			scoreLines = append(scoreLines, line)
		}
	}
	return strings.Join(scoreLines, "\n")
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
