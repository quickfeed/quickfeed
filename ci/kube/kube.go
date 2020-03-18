package kube

import (
	"context"
)

//Container describes how to execute a CI job.
type Container struct {
	Image    string
	Commands []string
}

//KRunner contains methods for running user provided code in isolation.
type KRunner interface {
	// RunKubeJob should synchronously execute the described job and return the output.
	KRun(context.Context, *Container, string, string, string) (string, error)
}
