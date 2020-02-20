package kube

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/autograde/aguis/ci"
)

// K8s is an implementation of the CI interface using K8s.
type K8s struct {
	Endpoint string
}

var (
	podLog  string
	stat    bool
	jobLock sync.Mutex
	podLock sync.Mutex
	waiting sync.Cond = *sync.NewCond(&jobLock)
	condPod sync.Cond = *sync.NewCond(&podLock)
)

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }

//RunKubeJob runs the rescieved push from repository on the podes in our 3 nodes.
//dockJob is the container that will be creted using the base client docker image and commands that will run.
//id is a unique string for each job object
func (k *K8s) RunKubeJob(ctx context.Context, dockJob *ci.Job, id string, kubeconfig *string) (string, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return "", err
	}
	//K8s clinet
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}

	//Define the configiration of the job object
	jobsClient := clientset.BatchV1().Jobs("agcicd")
	confJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cijob" + id,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: int32Ptr(8),
			//Parallelism:  int32Ptr(3), //TODO starting with 1 pod, def
			//Completions:             int32Ptr(10), //TODO  starting with 1 pod, def
			TTLSecondsAfterFinished: int32Ptr(20),
			ActiveDeadlineSeconds:   int64Ptr(1000), // terminate after 1000 sec ?
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

	creteadJob, err := jobsClient.Create(confJob)
	if err != nil {
		return "", err
	}

	jobLock.Lock()
	if !jobEvents(*creteadJob, clientset, "agcicd", int32(1), creteadJob.Name) {
		waiting.Wait()
	}
	jobLock.Unlock()

	pods, err := clientset.CoreV1().Pods("agcicd").List(metav1.ListOptions{
		LabelSelector: "job-name=" + creteadJob.Name,
	})
	if err != nil {
		return "could not list the pods!", nil
	}

	for range pods.Items {
		podLock.Lock()
		if !podEvents(clientset, "agcicd", creteadJob.Name) {
			condPod.Wait()
		}
		podLock.Unlock()
	}
	return podLog, nil
}

//DeleteObject deleting ..
func (k *K8s) DeleteObject(pod apiv1.Pod, clientset kubernetes.Clientset, namespace string) error {
	jobs, err := clientset.BatchV1().Jobs(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	if len(jobs.Items) > 0 {
		for _, job := range jobs.Items {
			err := clientset.BatchV1().Jobs(namespace).Delete(job.Name, &metav1.DeleteOptions{
				GracePeriodSeconds: int64Ptr(30),
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//PodLogs returns ...
func podLogs(pod apiv1.Pod, clientset *kubernetes.Clientset, namespace string) string {
	podLogOpts := apiv1.PodLogOptions{}

	req := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &podLogOpts)
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

//PodEvents is ...
func podEvents(clientset *kubernetes.Clientset, namespace string, jobname string) bool {
	st := false
	watch, err := clientset.CoreV1().Pods(namespace).Watch(metav1.ListOptions{
		LabelSelector: "job-name=" + jobname,
	})
	if err != nil {
		panic(err)
	}
	go func(st bool) {
		for event := range watch.ResultChan() {
			fmt.Printf("Type: %v\n", event.Type)
			pod, ok := event.Object.(*v1.Pod)

			if !ok {
				panic("unexpected type")
			}
			if pod.Status.Phase == "Succeeded" {
				podLock.Lock()
				podLog = podLogs(*pod, clientset, "agcicd")
				st = true
				condPod.Signal()
				podLock.Unlock()
			}
			if pod.Status.Phase == apiv1.PodFailed {
				fmt.Println("POD FAILED manual print") //TODO: What to do? delete and run the job again?
			}
		}
	}(st)
	return st
}

func jobEvents(job batchv1.Job, clientset *kubernetes.Clientset, namespace string, nrOfPods int32, kubejobname string) bool {
	st := false
	watch, err := clientset.BatchV1().Jobs(namespace).Watch(metav1.ListOptions{
		LabelSelector: "job-name=" + kubejobname,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	go func(st bool) {
		for event := range watch.ResultChan() {
			fmt.Printf("Type: %v\n", event.Type)
			j, ok := event.Object.(*batchv1.Job)
			if !ok {
				log.Fatal("unexpected type")
			}
			//if j.Status.StartTime != nil {
			if j.Status.Active == nrOfPods {
				jobLock.Lock()
				st = true
				waiting.Signal()
				jobLock.Unlock()
			}
		}
	}(st)
	return st
}
