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
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"go.uber.org/zap"
)

// DefaultContainerTimeout is the default timeout for running a container.
var DefaultContainerTimeout = time.Duration(10 * time.Minute)

const (
	Dockerfile      = "Dockerfile"
	QuickFeedPath   = "/quickfeed"
	GoModCache      = "/quickfeed-go-mod-cache"
	maxToScan       = 1_000_000 // bytes
	maxLogSize      = 30_000    // bytes
	lastSegmentSize = 1_000     // bytes
)

// Docker is an implementation of the CI interface using Docker.
type Docker struct {
	client *client.Client
	logger *zap.SugaredLogger
}

// NewDockerCI returns a runner to run CI tests.
func NewDockerCI(logger *zap.SugaredLogger) (*Docker, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}
	return &Docker{
		client: cli,
		logger: logger,
	}, nil
}

// Close ensures that the docker client is closed.
func (d *Docker) Close() error {
	var syncErr error
	if d.logger != nil {
		syncErr = d.logger.Sync()
	}
	closeErr := d.client.Close()
	return errors.Join(syncErr, closeErr)
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
	if err = d.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	d.logger.Infof("Waiting for container image '%s' for %s", job.Image, job.Name)
	msg, err := d.waitForContainer(ctx, job, resp.ID)
	if err != nil {
		return msg, err
	}

	d.logger.Infof("Done waiting for container image '%s' for %s", job.Image, job.Name)
	// extract the logs before removing the container below
	logReader, err := d.client.ContainerLogs(ctx, resp.ID, container.LogsOptions{
		ShowStdout: true,
	})
	if err != nil {
		return "", err
	}

	d.logger.Infof("Removing container image '%s' for %s", job.Image, job.Name)
	// remove the container when finished to prevent too many open files
	err = d.client.ContainerRemove(ctx, resp.ID, container.RemoveOptions{})
	if err != nil {
		return "", err
	}

	var stdout bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, io.Discard, logReader); err != nil {
		return "", err
	}
	if stdout.Len() > maxLogSize+lastSegmentSize {
		return truncateLog(&stdout, maxLogSize, lastSegmentSize, maxToScan), nil
	}
	return stdout.String(), nil
}

// createImage creates an image for the given job.
func (d *Docker) createImage(ctx context.Context, job *Job) (*container.CreateResponse, error) {
	if job.Image == "" {
		// image name should be specified in a run.sh file in the tests repository
		return nil, fmt.Errorf("no image name specified for '%s'", job.Name)
	}
	dockerFileContent := job.BuildContext[Dockerfile]
	if dockerFileContent != "" {
		d.logger.Infof("Removing image '%s' for '%s' prior to rebuild", job.Image, job.Name)
		resp, err := d.client.ImageRemove(ctx, job.Image, image.RemoveOptions{Force: true})
		if err != nil {
			d.logger.Debugf("Expected error (continuing): %v", err)
			// continue because we may not have an image to remove
		}
		for _, r := range resp {
			d.logger.Infof("Removed image '%s' for '%s'", r.Deleted, job.Name)
		}

		d.logger.Infof("Trying to build image: '%s' from Dockerfile", job.Image)
		// Log first line of Dockerfile
		d.logger.Infof("[%s] Dockerfile: %s ...", job.Image, dockerFileContent[:strings.Index(dockerFileContent, "\n")+1])
		if err := d.buildImage(ctx, job); err != nil {
			return nil, err
		}
	}

	var hostConfig *container.HostConfig
	if job.BindDir != "" {
		goModCacheSrc, err := moduleCachePath()
		if err != nil {
			return nil, err
		}
		hostConfig = &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: job.BindDir,
					Target: QuickFeedPath,
				},
				{
					Type:   mount.TypeBind,
					Source: goModCacheSrc,
					Target: GoModCache,
				},
			},
		}
	}

	create := func() (container.CreateResponse, error) {
		return d.client.ContainerCreate(ctx, &container.Config{
			Image: job.Image,
			User:  fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), // Run the image as the current user, e.g., quickfeed
			Env:   job.Env,                                        // Set default environment variables
			Cmd:   []string{"/bin/bash", "-c", strings.Join(job.Commands, "\n")},
		}, hostConfig, nil, nil, job.Name)
	}

	resp, err := create()
	switch {
	case errdefs.IsConflict(err):
		d.logger.Errorf("Image '%s' already being built for '%s': %v", job.Image, job.Name, err)
		return nil, ErrConflict
	case err != nil:
		d.logger.Errorf("Image '%s' not yet available for '%s': %v", job.Image, job.Name, err)
		d.logger.Infof("Trying to pull image: '%s' from remote repository", job.Image)
		if err := d.pullImage(ctx, job.Image); err != nil {
			return nil, err
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
			timeout := 1 // seconds to wait before forcefully killing the container
			stopErr := d.client.ContainerStop(context.Background(), respID, container.StopOptions{Timeout: &timeout})
			if stopErr != nil {
				return "", stopErr
			}
			// remove the docker container (when stopped due to timeout) to prevent too many open files
			rmErr := d.client.ContainerRemove(context.Background(), respID, container.RemoveOptions{})
			if rmErr != nil {
				return "", rmErr
			}
			// return message to user to be shown in the results log
			return "Container timeout. Please check for infinite loops or other slowness.", err
		}
	case status := <-statusCh:
		d.logger.Infof("Container: '%s' for %s: exited with status: %v", job.Image, job.Name, status.StatusCode)
	}
	return "", nil
}

// pullImage pulls an image from docker hub.
// This can be slow and should be avoided if possible.
func (d *Docker) pullImage(ctx context.Context, imageName string) error {
	progress, err := d.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer progress.Close()

	_, err = io.Copy(io.Discard, progress)
	return err
}

// buildImage builds and installs an image locally to be reused in a future run.
func (d *Docker) buildImage(ctx context.Context, job *Job) error {
	var buf bytes.Buffer
	tarWriter := tar.NewWriter(&buf)

	// Ensure consistent order of files in the tar archive
	for _, name := range slices.Sorted(maps.Keys(job.BuildContext)) {
		fileContents := []byte(job.BuildContext[name])
		if err := tarWriter.WriteHeader(&tar.Header{
			Name:     name,
			Mode:     0o777,
			Size:     int64(len(fileContents)),
			Typeflag: tar.TypeReg,
		}); err != nil {
			return err
		}
		if _, err := tarWriter.Write(fileContents); err != nil {
			return err
		}
	}
	if err := tarWriter.Close(); err != nil {
		return err
	}

	reader := bytes.NewReader(buf.Bytes())
	opts := build.ImageBuildOptions{
		Context:    reader,
		Dockerfile: Dockerfile,
		Tags:       []string{job.Image},
	}
	res, err := d.client.ImageBuild(ctx, reader, opts)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return printInfo(d.logger, res.Body)
}

func printInfo(logger *zap.SugaredLogger, rd io.Reader) error {
	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		out := &dockerJSON{}
		if err := json.Unmarshal([]byte(scanner.Text()), out); err != nil {
			return err
		}
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
