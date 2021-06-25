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
	"path/filepath"
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

	create := func() (container.ContainerCreateCreatedBody, error) {
		return d.client.ContainerCreate(ctx, &container.Config{
			Image: job.Image,
			Cmd:   []string{"/bin/bash", "-c", strings.Join(job.Commands, "\n")},
		}, nil, nil, nil, job.Name)
	}

	resp, err := create()
	if err != nil {
		d.logger.Errorf("Failed to create container image '%s' for %s: %v", job.Image, job.Name, err)
		// if image not found locally, try to pull it
		if err := d.pullImage(ctx, job.Image); err != nil {
			d.logger.Errorf("Failed to pull image '%s' from docker.io: %v", job.Image, err)
			if err := d.buildImage(ctx, job.Image); err != nil {
				return "", err
			}
		}
		resp, err = create()
		if err != nil {
			return "", err
		}
	}

	d.logger.Infof("Created container image '%s' for %s", job.Image, job.Name)
	if err := d.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	// wait until the container stops or context times out.
	statusCh, errCh := d.client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			d.logger.Errorf("Failed to stop container image '%s' for %s: %v", job.Image, job.Name, err)
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
	case <-statusCh:
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
func (d *Docker) pullImage(ctx context.Context, image string) error {
	d.logger.Infof("Pulling Docker image: '%s' from docker.io", image)
	progress, err := d.client.ImagePull(ctx, "docker.io/library/"+image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer progress.Close()

	_, err = io.Copy(ioutil.Discard, progress)
	return err
}

// buildImage builds and installs an image locally to be reused in a future run.
func (d *Docker) buildImage(ctx context.Context, image string) error {
	tag := image[strings.Index(image, ":")+1:]
	dockerfile := filepath.Join("scripts", tag, "Dockerfile")
	d.logger.Infof("Building image: '%s' from %s", image, dockerfile)

	contents, err := ioutil.ReadFile(dockerfile)
	if err != nil {
		return err
	}
	header := &tar.Header{
		Name:     "Dockerfile",
		Mode:     0o777,
		Size:     int64(len(contents)),
		Typeflag: tar.TypeReg,
	}
	var buf bytes.Buffer
	tarWriter := tar.NewWriter(&buf)
	if err = tarWriter.WriteHeader(header); err != nil {
		return err
	}
	if _, err = tarWriter.Write(contents); err != nil {
		return err
	}
	if err = tarWriter.Close(); err != nil {
		return err
	}

	reader := bytes.NewReader(buf.Bytes())
	opts := types.ImageBuildOptions{
		Context:    reader,
		Dockerfile: "Dockerfile",
		Tags:       []string{image},
	}
	res, err := d.client.ImageBuild(ctx, reader, opts)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	err = print(d.logger, res.Body)
	if err != nil {
		return err
	}
	return nil
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
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
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
