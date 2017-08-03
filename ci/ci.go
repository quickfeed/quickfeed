package ci

import "context"

// Job describes how to execute a CI job.
type Job struct {
	Image    string
	Commands []string
}

// Runner contains methods for running user provided code in isolation.
type Runner interface {
	// Run should synchronously execute the described job and return the output.
	Run(context.Context, *Job) (string, error)
}
