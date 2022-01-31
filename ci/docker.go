package ci

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"go.uber.org/zap"
)

var (
	DefaultContainerTimeout = time.Duration(10 * time.Minute)
	maxToScan               = 1_000_000 // bytes
	maxLogSize              = 30_000    // bytes
	lastSegmentSize         = 1_000     // bytes
)

// Docker is an implementation of the CI interface using Docker.
type Docker struct {
	client *client.Client
	logger *zap.SugaredLogger
}

// NewDockerCI returns a runner to run CI tests.
func NewDockerCI(logger *zap.Logger) (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
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

	resp, err := d.createImage(ctx, job)
	if err != nil {
		return "", err
	}
	d.logger.Infof("Created container image '%s' for %s", job.Image, job.Name)
	if err := d.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	d.logger.Infof("Waiting for container image '%s' for %s", job.Image, job.Name)
	msg, err := d.waitForContainer(ctx, job, resp.ID)
	if err != nil {
		return msg, err
	}

	d.logger.Infof("Done waiting for container image '%s' for %s", job.Image, job.Name)
	// extract the logs before removing the container below
	logReader, err := d.client.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
	})
	if err != nil {
		return "", err
	}

	d.logger.Infof("Removing container image '%s' for %s", job.Image, job.Name)
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
		return truncateLog(&stdout, maxLogSize, lastSegmentSize, maxToScan), nil
	}
	return stdout.String(), nil
}

// createImage creates an image for the given job.
func (d *Docker) createImage(ctx context.Context, job *Job) (*container.ContainerCreateCreatedBody, error) {
	if job.Image == "" {
		// image name should be specified in a run.sh file in the tests repository
		return nil, fmt.Errorf("no image name specified for '%s'", job.Name)
	}

	create := func() (container.ContainerCreateCreatedBody, error) {
		return d.client.ContainerCreate(ctx, &container.Config{
			Image: job.Image,
			Cmd:   []string{"/bin/bash", "-c", strings.Join(job.Commands, "\n")},
		}, nil, nil, nil, job.Name)
	}

	resp, err := create()
	if err != nil {
		d.logger.Infof("Image '%s' not yet available for '%s': %v", job.Image, job.Name, err)

		if job.Dockerfile != "" {
			d.logger.Infof("Trying to build image: '%s' from Dockerfile", job.Image)
			if err := d.buildImage(ctx, job); err != nil {
				return nil, err
			}
		} else {
			d.logger.Infof("Trying to pull image: '%s' from docker.io", job.Image)
			if err := d.pullImage(ctx, job.Image); err != nil {
				return nil, err
			}
		}
		// try to create the container again
		resp, err = create()
	}
	return &resp, err
}

// waitForContainer waits until the container stops or context times out.
func (d *Docker) waitForContainer(ctx context.Context, job *Job, respID string) (string, error) {
	statusCh, errCh := d.client.ContainerWait(ctx, respID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			d.logger.Errorf("Failed to stop container image '%s' for %s: %v", job.Image, job.Name, err)
			if !errors.Is(err, context.DeadlineExceeded) {
				return "", err
			}
			// stop runaway container whose deadline was exceeded
			timeout := time.Duration(1 * time.Second)
			stopErr := d.client.ContainerStop(context.Background(), respID, &timeout)
			if stopErr != nil {
				return "", stopErr
			}
			// remove the docker container (when stopped due to timeout) to prevent too many open files
			rmErr := d.client.ContainerRemove(context.Background(), respID, types.ContainerRemoveOptions{})
			if rmErr != nil {
				return "", rmErr
			}
			// return message to user to be shown in the results log
			return "Container timeout. Please check for infinite loops or other slowness.", err
		}
	case <-statusCh:
	}
	return "", nil
}

// pullImage pulls an image from docker hub.
// This can be slow and should be avoided if possible.
func (d *Docker) pullImage(ctx context.Context, image string) error {
	progress, err := d.client.ImagePull(ctx, "docker.io/library/"+image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer progress.Close()

	_, err = io.Copy(ioutil.Discard, progress)
	return err
}

// buildImage builds and installs an image locally to be reused in a future run.
func (d *Docker) buildImage(ctx context.Context, job *Job) error {
	dockerFileContents := []byte(job.Dockerfile)
	header := &tar.Header{
		Name:     "Dockerfile",
		Mode:     0o777,
		Size:     int64(len(dockerFileContents)),
		Typeflag: tar.TypeReg,
	}
	var buf bytes.Buffer
	tarWriter := tar.NewWriter(&buf)
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}
	if _, err := tarWriter.Write(dockerFileContents); err != nil {
		return err
	}
	if err := tarWriter.Close(); err != nil {
		return err
	}

	reader := bytes.NewReader(buf.Bytes())
	opts := types.ImageBuildOptions{
		Context:    reader,
		Dockerfile: "Dockerfile",
		Tags:       []string{job.Image},
	}
	res, err := d.client.ImageBuild(ctx, reader, opts)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return print(d.logger, res.Body)
}

func print(logger *zap.SugaredLogger, rd io.Reader) error {
	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		out := &dockerJSON{}
		json.Unmarshal([]byte(scanner.Text()), out)
		if out.Error != "" {
			return errors.New(out.Error)
		}
		logger.Info(out)
	}
	return scanner.Err()
}

type dockerJSON struct {
	Status string `json:"status"`
	ID     string `json:"id"`
	Stream string `json:"stream"`
	Error  string `json:"error"`
}

func (s dockerJSON) String() string {
	if len(s.Status) > 0 {
		return s.Status + s.ID
	}
	return strings.TrimSpace(s.Stream)
}
