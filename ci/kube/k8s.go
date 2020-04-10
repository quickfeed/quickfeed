package kube

import (
	"bytes"
	"context"
	"io"
	"log"
	"strings"
	"sync"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// K8s is an implementation of the Kube interface using K8s.
type K8s struct {
	isActive   int
	isLogReady int
	result     string
}

var (
	jobActivated sync.Cond = *sync.NewCond(&sync.Mutex{})
	podExecuted  sync.Cond = *sync.NewCond(&sync.Mutex{})
	nrOfPods               = int32Ptr(1)
	//cli = getClient()
)

// KRun implements the Kube Interface.
// This method creates a K8s Job and Secret Object.
// The Job object takes the task from the Autograder, and run it in a Docker Container.
// The Job Object will be created in the spesified NameSpace by the Parameter "courseName".
// Tasks can run in parallel. The method will not return before the Job is on completed status.
func (k *K8s) KRun(ctx context.Context, task *Container, id string, courseName string, secretAg string) (string, error) {

	clientset, err := getClient()
	if err != nil {
		return "", err
	}

	err = makeSecret(clientset, id, courseName, secretAg)
	if err != nil {
		return "", err
	}

	jobClient := clientset.BatchV1().Jobs(courseName)
	// Dynamically define the configuration of the job object.
	jobConfig := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "job" + id,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            int32Ptr(8),
			Parallelism:             nrOfPods,
			Completions:             nrOfPods,
			TTLSecondsAfterFinished: int32Ptr(60),
			ActiveDeadlineSeconds:   int64Ptr(3600),
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            "container" + id,
							Image:           task.Image,
							Command:         []string{"/bin/sh", "-c", strings.Join(task.Commands, "\n")},
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
							// The secret is created will be Mounted be the Continer
							// TODO: Vera Yaseneva - command out line 43 - 46, 80 - 98, and 123, if K8s secret not used.
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

	jobObject, err := jobClient.Create(jobConfig)
	if err != nil {
		return "", err
	}
	// Watch the Job events until it is activated.
	err = k.jobEvents(*jobObject, clientset, courseName, jobObject.Name)
	if err != nil {
		return "", err
	}
	k.waitForJob()
	// Watch the Pod events until it succeed.
	err = k.podEvents(clientset, courseName, jobObject.Name)
	if err != nil {
		return "", err
	}
	k.waitForLogs()

	removeSecret(clientset, id, courseName)
	return k.result, nil
}

// waitForJob waits for a signal from the jobEvents() that indicates that the job status is set on active.
func (k *K8s) waitForJob() {
	jobActivated.L.Lock()
	for k.isActive == 0 {
		jobActivated.Wait()
	}
	k.isActive--
	jobActivated.L.Unlock()
}

// waitForLogs waits for a signal from the podEvents() that indicates that the Pod status is set on Succeeded.
func (k *K8s) waitForLogs() {
	podExecuted.L.Lock()
	for k.isLogReady == 0 {
		podExecuted.Wait()
	}
	k.isLogReady--
	podExecuted.L.Unlock()
}

// jobEvents watches a job events. When the job status is active, a signal will be sent to the waitForJob() method.
func (k *K8s) jobEvents(job batchv1.Job, clientset *kubernetes.Clientset, namespace string, jobname string) error {
	watch, err := clientset.BatchV1().Jobs(namespace).Watch(metav1.ListOptions{
		LabelSelector: "job-name=" + jobname,
	})
	if err != nil {
		return err
	}
	go func() {
		for event := range watch.ResultChan() {
			j, ok := event.Object.(*batchv1.Job)
			if !ok {
				log.Fatal("unexpected type")
			}
			if j.Status.Active == *nrOfPods {
				jobActivated.L.Lock()
				k.isActive++
				jobActivated.Signal()
				jobActivated.L.Unlock()
			}
		}
	}()
	return nil
}

// podEvents watches a pods events. When the Pod status is Succeeded, a signal will be sent to the waitForLogs() method.
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
				podExecuted.L.Lock()
				k.result, err = podLogs(*pod, clientset, namespace)
				if err != nil {
					logError = err
				}
				k.isLogReady++
				podExecuted.Signal()
				podExecuted.L.Unlock()
			}
			/* 			if pod.Status.Phase == apiv1.PodFailed {
				logError = err
			} */
		}
	}(logError)
	if logError != nil {
		return logError
	}
	return nil
}

// podLogs read the logs of the Succeeded pod(s).
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
