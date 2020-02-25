package kube

import (
	"context"

	"k8s.io/client-go/util/workqueue"
)

// K8sgp is an implementation of the CI interface using K8s.
type K8sq struct {
	queue          workqueue.Interface
	containerImage string
}

//Add adds ...
func (kq *K8sq) Add(cmd []string) {
	kq.queue.Add(cmd)
}

//Run runs..
func (k *K8sq) Run(ctx context.Context) (string, error) {

	return "logs needed!", nil
}
