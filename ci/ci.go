package ci

import (
	"context"
)

// Job describes how to execute a CI job.
type Job struct {
	// Name describes the running job; mainly used to name docker containers.
	Name string
	// Image names the image to use to run the job.
	Image string
	// BuildContext is a list of files to include in the build context.
	// These files are copied to the image's /quickfeed directory.
	// If the Dockerfile isn't present, the image is assumed to exist.
	// If the Dockerfile is present, the image is built from the Dockerfile.
	BuildContext map[string]string
	// BindDir is the directory to bind to the container's /quickfeed directory.
	BindDir string
	// Env is a list of environment variables to set for the job.
	Env []string
	// Commands is a list of shell commands to run as part of the job.
	Commands []string
}

// Runner contains methods for running user provided code in isolation.
type Runner interface {
	// Run should synchronously execute the described job and return the output.
	Run(context.Context, *Job) (string, error)
}
