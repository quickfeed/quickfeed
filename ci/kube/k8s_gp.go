package kube

import (
	"context"

	"github.com/autograde/aguis/ci"

	"k8s.io/client-go/util/workqueue"
)

// K8sgp is an implementation of the CI interface using K8s.
type K8sgp struct {
	queue          workqueue.Interface
	containerImage string
}

//ExtractCmd extraxts the commands part of the container
func (k *K8sgp) ExtractCmd(ctx context.Context, c *ci.Job) {

}

//Run runs..
func (k *K8sgp) Run(ctx context.Context) (string, error) {

	return "logs needed!", nil
}
