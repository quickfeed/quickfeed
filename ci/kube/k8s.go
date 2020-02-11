package kube

import (
	"context"
	"io"
	"io/ioutil"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/autograde/aguis/ci"
)

// K8s is an implementation of the CI interface using K8s.
type K8s struct {
}

//CreateJob runs the rescieved push from repository on the podes in our 3 nodes.
func (k *K8s) RunKubeJob(ctx context.Context, dockJob *ci.Job) (string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}

	//create a docker client to pull the image ?!
	dockCli, err := client.NewEnvClient()
	if err != nil {
		return "", err
	}

	//pull the docker image
	if err := pullImage(ctx, dockCli, dockJob.Image); err != nil {
		return "", err
	}

	//create job for every push ?!
	jobsClient := clientset.BatchV1().Jobs("agcicd")
	kubeJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "webhook-job",
			//Namespace: "agcicd",
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: int32Ptr(8),
			Parallelism:  int32Ptr(2), //TODO
			Completions:  int32Ptr(2), //TODO
			//ttlSecondsAfterFinished: 100
			//activeDeadlineSeconds: ?
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:    "webhook-job",
							Image:   dockJob.Image,
							Command: []string{"/bin/sh", "-c", strings.Join(dockJob.Commands, "\n")},
						},
					},
					RestartPolicy: apiv1.RestartPolicyOnFailure, //necessaray to set either onfailure or ?..
				},
			},
		},
	}
	result, err := jobsClient.Create(kubeJob)
	if err != nil {
		return "", err
	}

	return "TODO:logs needed to be retunrned for the job:" + result.Name, nil
}

//Result returns the result of recently push that are executed on the nodes ?
func (k *K8s) Result() string {
	result := ""
	return result
}

func pullImage(ctx context.Context, dockCli *client.Client, image string) error {
	progress, err := dockCli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer progress.Close()

	_, err = io.Copy(ioutil.Discard, progress)
	return err
}
