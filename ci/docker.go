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
	"log/slog"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/client"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
)

// DefaultContainerTimeout is the default timeout for running a container.
var DefaultContainerTimeout = time.Duration(10 * time.Minute)

const (
	Dockerfile      = "Dockerfile"
	QuickFeedPath   = "/quickfeed"
	maxToScan       = 1_000_000 // bytes
	maxLogSize      = 30_000    // bytes
	lastSegmentSize = 1_000     // bytes
	imageLabel      = "image"
	jobLabel        = "job"
)

// Docker is an implementation of the CI interface using Docker.
type Docker struct {
	client *client.Client
}

// NewDockerCI returns a runner to run CI tests.
func NewDockerCI() (*Docker, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &Docker{client: cli}, nil
}

// Close ensures that the docker client is closed.
func (d *Docker) Close() error {
	return d.client.Close()
}

// Run implements the CI interface. This method blocks until the job has been
// completed or an error occurs, e.g., the context times out.
func (d *Docker) Run(ctx context.Context, job *Job) (string, error) {
	if d.client == nil {
		return "", fmt.Errorf("cannot run job: %s; docker client not initialized", job.Name)
	}
	// Scope every record for this job, including those from the helpers below.
	ctx, logger := qlog.WithLogger(ctx, imageLabel, job.Image, jobLabel, job.Name)

	resp, err := d.createImage(ctx, job)
	if err != nil {
		return "", err
	}
	logger.Info("created container")
	if _, err = d.client.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return "", err
	}

	logger.Info("waiting for container")
	msg, err := d.waitForContainer(ctx, resp.ID)
	if err != nil {
		return msg, err
	}

	logger.Info("container completed")
	// extract the logs before removing the container below
	logReader, err := d.client.ContainerLogs(ctx, resp.ID, client.ContainerLogsOptions{
		ShowStdout: true,
	})
	if err != nil {
		return "", err
	}

	logger.Info("removing container")
	// remove the container when finished to prevent too many open files
	_, err = d.client.ContainerRemove(ctx, resp.ID, client.ContainerRemoveOptions{})
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
func (d *Docker) createImage(ctx context.Context, job *Job) (*client.ContainerCreateResult, error) {
	logger := qlog.FromContext(ctx)
	if job.Image == "" {
		// image name should be specified in a run.sh file in the tests repository
		return nil, fmt.Errorf("no image name specified for '%s'", job.Name)
	}
	dockerFileContent := job.BuildContext[Dockerfile]
	if dockerFileContent != "" {
		logger.Info("removing image before rebuild")
		resp, err := d.client.ImageRemove(ctx, job.Image, client.ImageRemoveOptions{Force: true})
		if err != nil {
			logger.Debug("image was not present before rebuild", label.Error, err)
			// continue because we may not have an image to remove
		}
		for _, r := range resp.Items {
			logger.Info("removed image", "removed_image", r.Deleted)
		}

		// Log the first line of the Dockerfile with the build record.
		logger.Info("building image from Dockerfile", "dockerfile_first_line", dockerFileContent[:strings.Index(dockerFileContent, "\n")+1])
		if err := d.buildImage(ctx, job); err != nil {
			return nil, err
		}
	}

	var hostConfig *container.HostConfig
	if job.BindDir != "" {
		mounts := []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: job.BindDir,
				Target: QuickFeedPath,
			},
		}
		if cfg, ok := languages[job.Language]; ok {
			for target, pathFn := range cfg.cacheDirs {
				src, err := pathFn()
				if err != nil {
					return nil, err
				}
				mounts = append(mounts, mount.Mount{
					Type:   mount.TypeBind,
					Source: src,
					Target: target,
				})
			}
		}
		for _, src := range slices.Sorted(maps.Keys(job.ReadOnlyMounts)) {
			mounts = append(mounts, mount.Mount{
				Type:     mount.TypeBind,
				Source:   src,
				Target:   job.ReadOnlyMounts[src],
				ReadOnly: true,
			})
		}
		hostConfig = &container.HostConfig{
			Mounts: mounts,
		}
	}

	create := func() (client.ContainerCreateResult, error) {
		return d.client.ContainerCreate(ctx, client.ContainerCreateOptions{
			Config: &container.Config{
				Image: job.Image,
				User:  fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), // Run the image as the current user, e.g., quickfeed
				Env:   job.Env,                                        // Set default environment variables
				Cmd:   []string{"/bin/bash", "-c", strings.Join(job.Commands, "\n")},
			},
			HostConfig: hostConfig,
			Name:       job.Name,
		})
	}

	resp, err := create()
	switch {
	case errdefs.IsConflict(err):
		logger.Error("image already being built", label.Error, err)
		return nil, ErrConflict
	case err != nil:
		logger.Error("image not available locally", label.Error, err)
		logger.Info("pulling image")
		if err := d.pullImage(ctx, job.Image); err != nil {
			return nil, err
		}
		// try to create the container again
		resp, err = create()
	}
	return &resp, err
}

// waitForContainer waits until the container stops or context times out.
func (d *Docker) waitForContainer(ctx context.Context, respID string) (string, error) {
	logger := qlog.FromContext(ctx)
	wait := d.client.ContainerWait(ctx, respID, client.ContainerWaitOptions{Condition: container.WaitConditionNotRunning})
	select {
	case err := <-wait.Error:
		if err != nil {
			logger.Error("failed to stop container", label.Error, err)
			if !errors.Is(err, context.DeadlineExceeded) {
				return "", err
			}
			// The job context has expired, so use a detached deadline to bound the
			// complete stop-and-remove sequence. The Docker stop timeout separately
			// controls the daemon's grace period before it forcefully kills the container.
			timeout := 1 // seconds to wait before forcefully killing the container
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			defer cancel()
			_, stopErr := d.client.ContainerStop(cleanupCtx, respID, client.ContainerStopOptions{Timeout: &timeout})
			if stopErr != nil {
				return "", stopErr
			}
			// remove the docker container (when stopped due to timeout) to prevent too many open files
			_, rmErr := d.client.ContainerRemove(cleanupCtx, respID, client.ContainerRemoveOptions{})
			if rmErr != nil {
				return "", rmErr
			}
			// return message to user to be shown in the results log
			return "Container timeout. Please check for infinite loops or other slowness.", err
		}
	case status := <-wait.Result:
		logger.Info("container exited", "exit_status", status.StatusCode)
	}
	return "", nil
}

// pullImage pulls an image from docker hub.
// This can be slow and should be avoided if possible.
func (d *Docker) pullImage(ctx context.Context, imageName string) error {
	progress, err := d.client.ImagePull(ctx, imageName, client.ImagePullOptions{})
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
	opts := client.ImageBuildOptions{
		Context:    reader,
		Dockerfile: Dockerfile,
		Tags:       []string{job.Image},
	}
	res, err := d.client.ImageBuild(ctx, reader, opts)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return printInfo(qlog.FromContext(ctx), res.Body)
}

// printInfo logs the Docker build output, one record per line, at debug level;
// a single build emits many lines.
func printInfo(logger *slog.Logger, rd io.Reader) error {
	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		out := &dockerJSON{}
		if err := json.Unmarshal([]byte(scanner.Text()), out); err != nil {
			return err
		}
		if out.Error != "" {
			return errors.New(out.Error)
		}
		logger.Debug("docker build output", "output", out.String())
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
