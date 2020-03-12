package kube

import (
	"context"
)

// Job describes how to execute a CI job.
type PodContainer struct {
	Image    string
	Commands []string
}

// Runner contains methods for running user provided code in isolation.
type KubeRunner interface {
	// RunKubeJob should synchronously execute the described job and return the output.
	RunKubeJob(context.Context, *PodContainer, string, string /* , string */) (string /* , string */, error)
}
