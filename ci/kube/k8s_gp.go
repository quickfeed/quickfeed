package kube

import (
	"context"
	"strings"

	//corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/docker/docker/client"

	"github.com/autograde/aguis/ci"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
)

// K8sgp is an implementation of the CI interface using K8s.
type K8sgp struct {
	queue workqueue.Interface
}

//Put the job in a working queue
func (k *K8sgp) AddJob(ctx context.Context, cijob *ci.Job) {

	//create a docker client to pull the image ?!
	dockCli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	//pull the docker image
	if err := pullImage(ctx, dockCli, cijob.Image); err != nil {
		panic(err)
	}
	cijob.Commands = []string{"/bin/sh", "-c", strings.Join(cijob.Commands, "\n")}

	inn := &ci.Job{Image: cijob.Image, Commands: cijob.Commands}
	k.queue.Add(inn)

}

//Run runs..
func (k *K8sgp) Run(d ci.Docker, dockJob *ci.Job, ctx context.Context) (string, error) {
	q := k.queue
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}

	if q.Len() != 0 {
		getJob, _ := q.Get()
		for {

			//create job for every push ?!
			jobsClient := clientset.BatchV1().Jobs("agcicd")
			kubeJob := &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name: "webhook-job",
					//Namespace: "agcicd",
				},
				Spec: batchv1.JobSpec{
					BackoffLimit: int32Ptr(10),
					Parallelism:  int32Ptr(3),
					//ttlSecondsAfterFinished: 100
					Template: apiv1.PodTemplateSpec{
						Spec: apiv1.PodSpec{
							Containers: []apiv1.Container{
								{
									Name:  "webhook-job",
									Image: dockJob.Image, //TODO
									//Command: getJob. TODO
								},
							},
							RestartPolicy: apiv1.RestartPolicyOnFailure, //necessaray to set either onfailure or ?..
						},
					},
				},
			}
		}
	}
}
