package kube

import (
	"fmt"
	"context"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"

	"github.com/autograde/aguis/ci"
)

// K8s is an implementation of the CI interface using K8s.
type K8s struct {
	queue PriorityQueue
}

//CreateJob runs the rescieved push from repository on the podes in our 3 nodes.
func (k *K8s) CreateJob(d ci.Docker, dockJob *ci.Job, ctx context.Context) {
	config, err := rest.InClusterConfig()
	check(err)

	clientset, err := kubernetes.NewForConfig(config)
	check(err)

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
	jobsClient := clientset.BatchV1().Jobs()
	kubeJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-job",
			Namespace: "agcicd",
		},
		Spec: batchv1.JobSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "webhook-job",
							Image: dockJob.Image,
							Command: []string{"/bin/sh", "-c", strings.Join(dockJob.Commands, "\n")}
						},
					},
					RestartPolicy:    apiv1.RestartPolicyOnFailure, //necessaray to set either onfailure or ?..
				},
			},
		},
	}

	fmt.Println("Creating job... ")
	result1, err := jobsClient.Create(kubeJob)
	check(err)
	fmt.Printf("Created job %q.\n", result1.Name)

}

//Result returns the result of recently push that are executed on the nodes ?
func (k *K8s) Result() string {
	result := ""
	return result
}

func pullImage(ctx context.Context, dockCli *client.Client, image string) error {
	progress, err := dockCli.ImagePull(ctx, image, types.ImagePullOptions{})
	check(err)

	defer progress.Close()

	_, err = io.Copy(ioutil.Discard, progress)
	return err
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}