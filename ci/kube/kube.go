package kube

import (
	"context"
)

type KubeConf struct {
	ConfigFlag *string //flag that addresses to the kubeconfig file path
}

// Job describes how to execute a CI job.
type PodContainer struct {
	BaseImage    string
	ContainerCmd []string
}

// Runner contains methods for running user provided code in isolation.
type KubeRunner interface {
	// Run should synchronously execute the described job and return the output.
	RunKubeJob(context.Context, *PodContainer, string, string, *KubeConf) (string, error)
}
