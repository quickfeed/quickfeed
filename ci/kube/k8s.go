package kube

import (
	"bytes"
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
	Endpoint string
	Version  string
}

//CreateJob runs the rescieved push from repository on the podes in our 3 nodes.
//dockJob is the container that will be creted using the base client docker image and commands that will run.
//id is a unique string for each job object
func (k *K8s) RunKubeJob(ctx context.Context, dockJob *ci.Job, id string) (string, error) {
	//only for inside the cluster configurations ..
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}

	//K8s clinet
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}

	//Docker client
	dockCli, err := client.NewEnvClient()
	if err != nil {
		return "", err
	}

	//Pull the docker image
	if err := pullImage(ctx, dockCli, dockJob.Image); err != nil {
		return "", err
	}

	//Define the configiration of the job object
	jobsClient := clientset.BatchV1().Jobs("agcicd")
	kubeJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "webhook-job" + id,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: int32Ptr(8),
			//Parallelism:  int32Ptr(1), //TODO starting with 1 pod, def
			//Completions:  int32Ptr(1), //TODO  starting with 1 pod, def
			//ttlSecondsAfterFinished: 30
			//activeDeadlineSeconds: ?
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:    "webhook-job" + id,
							Image:   dockJob.Image,
							Command: []string{"/bin/sh", "-c", strings.Join(dockJob.Commands, "\n")},
						},
					},
					RestartPolicy: apiv1.RestartPolicyOnFailure,
				},
			},
		},
	}
	_, err = jobsClient.Create(kubeJob)
	if err != nil {
		return "", err
	}

	logs := ""
	pods, err := clientset.CoreV1().Pods("agcicd").List(metav1.ListOptions{FieldSelector: ("metadata.name=webhook-job" + id)})
	for _, pod := range pods.Items {
		logs += k.PodLogs(pod, clientset)
	}

	return logs, nil
}

//PodLogs returns the result of recently push that are executed on the nodes ?
func (k *K8s) PodLogs(pod apiv1.Pod, clientset *kubernetes.Clientset) string {
	//delete ?
	podLogOpts := apiv1.PodLogOptions{}

	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream()
	if err != nil {
		panic(err)
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		panic(err)
	}
	str := buf.String()
	return str
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
