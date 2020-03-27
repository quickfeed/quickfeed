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

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/resource"
)

// K8s is an implementation of the CI interface using K8s.
type K8s struct {
	isActive   int
	isLogReady int
	podLog     string
}

var (
	isJobActive sync.Cond = *sync.NewCond(&sync.Mutex{})
	isPodDone   sync.Cond = *sync.NewCond(&sync.Mutex{})
)

//KRun runs the rescieved push from repository on the podes in our 3 nodes.
//dockJob is the container that will be creted using the base client docker image and commands that will run.
//id is a unique string for each job object
func (k *K8s) KRun(ctx context.Context, podRun *Container, id string, courseName string /* , secretAg string */) (string, error) {

	clientset, err := getClient()
	if err != nil {
		return "", err
	}

	/* err = Jobsecrets(id, courseName, secretAg)
	if err != nil {
		return "", err
	} */

	//Define the configiration of the job object
	jobsClient := clientset.BatchV1().Jobs(courseName)
	confJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cijob" + id,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            int32Ptr(8),
			Parallelism:             int32Ptr(1),
			Completions:             int32Ptr(1),
			TTLSecondsAfterFinished: int32Ptr(60),
			ActiveDeadlineSeconds:   int64Ptr(120000), // change this if jobs can take longer time.
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            "cijob" + id,
							Image:           podRun.Image,
							Command:         []string{"/bin/sh", "-c", strings.Join(podRun.Commands, "\n")},
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									"cpu":    resource.MustParse("700m"),
									"memory": resource.MustParse("1Gi"),
								},
								Requests: apiv1.ResourceList{
									"cpu":    resource.MustParse("700m"),
									"memory": resource.MustParse("1Gi"),
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "secreting",
									MountPath: "/root/work/secreting",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "secreting",
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName: id,
								},
							},
						},
					},
					RestartPolicy: apiv1.RestartPolicyOnFailure,
				},
			},
		},
	}

	createdJob, err := jobsClient.Create(confJob)
	if err != nil {
		return "", err
	}
	err = k.jobEvents(*createdJob, clientset, courseName, int32(1), createdJob.Name)
	if err != nil {
		return "", err
	}
	k.jobWaitToActive()
	fmt.Println("after jobevent")

	err = k.podEvents(clientset, courseName, createdJob.Name)
	if err != nil {
		return "", err
	}
	k.podWaitToSucc()

	return k.podLog, nil
}

//podWaitToSucc waits for the pod to success. The signal will be send from the k.podEvents function
func (k *K8s) podWaitToSucc() {
	isPodDone.L.Lock()
	for k.isLogReady == 0 {
		isPodDone.Wait()
	}
	k.isLogReady--
	isPodDone.L.Unlock()
}

//jobWaitToActive waits for the pod to success. The signal will be send from the k.jobEvents function
func (k *K8s) jobWaitToActive() {
	isJobActive.L.Lock()
	for k.isActive == 0 {
		isJobActive.Wait()
	}
	k.isActive--
	isJobActive.L.Unlock()
}

//jobEvents watch the events of the jobs. If job is active, then sends signal to the jobWaitToActive() method.
func (k *K8s) jobEvents(job batchv1.Job, clientset *kubernetes.Clientset, namespace string, nrOfPods int32, kubejobname string) error {
	watch, err := clientset.BatchV1().Jobs(namespace).Watch(metav1.ListOptions{
		LabelSelector: "job-name=" + kubejobname,
	})
	if err != nil {
		return err
	}
	go func() {
		for event := range watch.ResultChan() {
			j, ok := event.Object.(*batchv1.Job)
			if !ok {
				log.Fatal("unexpected type") //TODO need to handle this?
			}
			//if j.Status.StartTime != nil {
			if j.Status.Active == nrOfPods {
				isJobActive.L.Lock()
				k.isActive++
				isJobActive.Signal()
				isJobActive.L.Unlock()
			}
		}
	}()
	return nil
}

//podEvents watch the pods events, and send signal to the podWaitToSucc() method if pod successed.
func (k *K8s) podEvents(clientset *kubernetes.Clientset, namespace string, jobname string) error {
	watch, err := clientset.CoreV1().Pods(namespace).Watch(metav1.ListOptions{
		LabelSelector: "job-name=" + jobname,
	})
	if err != nil {
		return err
	}
	var logError error
	go func(logError error) {
		for event := range watch.ResultChan() {
			pod, ok := event.Object.(*apiv1.Pod)
			if !ok {
				log.Fatal("unexpected type")
			}
			if pod.Status.Phase == "Succeeded" {
				isPodDone.L.Lock()
				k.podLog, err = podLogs(*pod, clientset, namespace)
				if err != nil {
					logError = err
				}
				k.isLogReady++
				isPodDone.Signal()
				isPodDone.L.Unlock()
			}
			if pod.Status.Phase == apiv1.PodFailed {
				logError = err
			}
		}
	}(logError)
	if logError != nil {
		return logError
	}
	return nil
}

//podLogs read the logs of the cuurently running pods.
func podLogs(pod apiv1.Pod, clientset *kubernetes.Clientset, namespace string) (string, error) {
	podLogOpts := apiv1.PodLogOptions{}

	req := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream()
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}

	logsStr := buf.String()
	return logsStr, nil
}
