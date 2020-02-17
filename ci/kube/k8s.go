package kube

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/autograde/aguis/ci"
)

// K8s is an implementation of the CI interface using K8s.
type K8s struct {
	Endpoint string
}

func int32Ptr(i int32) *int32 { return &i }

//CreateJob runs the rescieved push from repository on the podes in our 3 nodes.
//dockJob is the container that will be creted using the base client docker image and commands that will run.
//id is a unique string for each job object
func (k *K8s) RunKubeJob(ctx context.Context, dockJob *ci.Job, id string) (string, error) {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	//K8s clinet
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	//Define the configiration of the job object
	jobsClient := clientset.BatchV1().Jobs("agcicd")
	kubeJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cijob" + id,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: int32Ptr(8),
			//Parallelism:  int32Ptr(1), //TODO starting with 1 pod, def
			//Completions:  int32Ptr(1), //TODO  starting with 1 pod, def
			//ttlSecondsAfterFinished: 30
			//activeDeadlineSeconds:
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            "cijob" + id,
							Image:           dockJob.Image,
							Command:         []string{"/bin/sh", "-c", strings.Join(dockJob.Commands, "\n")},
							ImagePullPolicy: apiv1.PullIfNotPresent,
						},
					},
					RestartPolicy: apiv1.RestartPolicyOnFailure,
				},
			},
		},
	}

	_, err = jobsClient.Create(kubeJob)
	if err != nil {
		fmt.Println("false 5")
		return "", err
	}

	logs := ""
	pods, err := clientset.CoreV1().Pods("agcicd").List(metav1.ListOptions{}) // TODO: does it find correct pod on correct job?
	fmt.Println(len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Println(pod.Status)
		//logs += k.PodLogs(pod, clientset)
		fmt.Println("logs : " + logs)
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

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
