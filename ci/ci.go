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
	// Commands is a list of shell commands to run as part of the job.
	Commands []string
}

// Runner contains methods for running user provided code in isolation.
type Runner interface {
	// Run should synchronously execute the described job and return the output.
	Run(context.Context, *Job) (string, error)
	BuildImage(context.Context, string, string, string) error
}
