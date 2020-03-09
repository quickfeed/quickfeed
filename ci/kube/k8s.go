package kube

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/resource"

	metrics "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/autograde/aguis/ci"
)

// K8s is an implementation of the CI interface using K8s.
type K8s struct {
	//Endpoint string
	//clientset clientset.Interface
	isActive   int
	isLogReady int
	podLog     string
}

var (
	isJobActive sync.Cond = *sync.NewCond(&sync.Mutex{})
	isPodDone   sync.Cond = *sync.NewCond(&sync.Mutex{})
)

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }

//RunKubeJob runs the rescieved push from repository on the podes in our 3 nodes.
//dockJob is the container that will be creted using the base client docker image and commands that will run.
//id is a unique string for each job object
//TODO: kubeconfig param has to be deleted?
func (k *K8s) RunKubeJob(ctx context.Context, dockJob *ci.Job, courseName string, id string, kubeconfig *string) (string, error) {
	// use the current context in kubeconfig TODO: this has to change
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return "", err
	}
	//K8s clinet
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}
	//TODO: need to change clientset ?
	//clientset := K8s{}

	//for now genrate a random secret and put it in the path root/work/volums
	//TODO: pass data that need to be in secret! and check for RC?
	err = jobsecrets(id, courseName, *clientset, kubeRandomSecret())
	if err != nil {
		return "", err
	}
	//Define the configiration of the job object
	jobsClient := clientset.BatchV1().Jobs(courseName)
	confJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cijob" + id,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            int32Ptr(8),
			Parallelism:             int32Ptr(1), //1 is default, change this if the k8s struggling running the scripts
			Completions:             int32Ptr(1), //1 is default, change this if the k8s struggling running the scripts
			TTLSecondsAfterFinished: int32Ptr(60),
			ActiveDeadlineSeconds:   int64Ptr(1200), // This depends on how big the tasks are.
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            "cijob" + id,
							Image:           dockJob.Image,
							Command:         []string{"/bin/sh", "-c", strings.Join(dockJob.Commands, "\n")},
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									"cpu":    resource.MustParse("200m"), //TODO: test by changing this to "2" and run 8 in parallell
									"memory": resource.MustParse("200Mi"),
								},
								Requests: apiv1.ResourceList{
									"cpu":    resource.MustParse("100m"), //TODO: test by changing this to "2" and run 8 in parallell
									"memory": resource.MustParse("100Mi"),
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
		return err
	}
	return nil
}

//jobsecrets create a secrets.. TODO
func jobsecrets(secretName string, namespace string, clientset kubernetes.Clientset, pass string) error {
	newSec := &apiv1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			secretName: []byte(pass),
		},
		Type: "Opaque",
	}
	_, err := clientset.CoreV1().Secrets(namespace).Create(newSec)
	if err != nil {
		return err
	}
	return nil
}

func kubeRandomSecret() string {
	randomness := make([]byte, 10)
	_, err := rand.Read(randomness)
	if err != nil {
		log.Fatal("couldn't generate randomness")
	}
	return fmt.Sprintf("%x", sha1.Sum(randomness))
}

func monitorResourceUsage() error {
	var kubeconfig string
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	cli, err := metrics.NewForConfig(config)

	podMetrics, err := cli.MetricsV1beta1().PodMetricses(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	for _, podMetric := range podMetrics.Items {
		podContainers := podMetric.Containers
		for _, container := range podContainers {
			cpuQuantity, ok := container.Usage.Cpu().AsInt64()
			memQuantity, ok := container.Usage.Memory().AsInt64()
			if !ok {
				log.Fatal("couldn'tfind the quntities!")
			}
			msg := fmt.Sprintf("Container Name: %s \n CPU usage: %d \n Memory usage: %d", container.Name, cpuQuantity, memQuantity)
			fmt.Println(msg)
		}

	}
	return nil
}
