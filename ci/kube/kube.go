package kube

import (
	"context"
)

// Container contains the Base Docker Image and the Commands.
// These describe the task that has to be executed in the Container of a K8s Job object.
type Container struct {
	Image    string
	Commands []string
}

// KRunner contains methods for configuring a K8s Job Object to run the user-provided code in isolation.
type KRunner interface {
	// Run should synchronously execute the described Container and return the output.
	KRun(context.Context, *Container, string, string, string) (string, error)
}
