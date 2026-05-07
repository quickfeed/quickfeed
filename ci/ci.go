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
	// Language specifies the programming language for the job.
	// When set to LanguageGo, Go-specific cache directories are mounted.
	// Parsed from the #language/ directive in the run script.
	Language string
	// BuildContext is a list of files to include in the Docker build context.
	// These files are available to the Dockerfile (e.g. via COPY/ADD) and can be
	// copied into the image, such as into the /quickfeed directory, if desired.
	// If the Dockerfile is present in the build context, the image is built from
	// the Dockerfile. Otherwise, the image is assumed to already exist.
	BuildContext map[string]string
	// BindDir is the directory to bind to the container's /quickfeed directory.
	BindDir string
	// ReadOnlyMounts maps host source paths to container target paths for read-only bind mounts.
	// These are mounted in addition to BindDir and are not affected by changes in the container.
	ReadOnlyMounts map[string]string
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
