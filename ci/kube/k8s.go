package kube

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

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
}

var (
	podLog  string
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

	jobsecrets(courseName, *clientset, "client token id or this kinds of things ?")

	//Define the configiration of the job object
	jobsClient := clientset.BatchV1().Jobs(courseName)
	confJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cijob" + id,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: int32Ptr(8),
			//Parallelism:  int32Ptr(4), //TODO starting with 1 pod, def
			//Completions:             int32Ptr(10), //TODO  starting with 1 pod, def
			TTLSecondsAfterFinished: int32Ptr(20),
			//ActiveDeadlineSeconds:   int64Ptr(1000), // terminate after 1000 sec ?
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
									"cpu":    resource.MustParse("100m"),
									"memory": resource.MustParse("100Mi"),
								},
								Requests: apiv1.ResourceList{
									"cpu":    resource.MustParse("100m"),
									"memory": resource.MustParse("100Mi"),
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "secreting",
									MountPath: "/root/work/volums",
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
									SecretName: "job-secret",
								},
							},
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
	if !jobEvents(*creteadJob, clientset, courseName, int32(1), creteadJob.Name) {
		waiting.Wait()
	}

	pods, err := clientset.CoreV1().Pods(courseName).List(metav1.ListOptions{
		LabelSelector: "job-name=" + creteadJob.Name,
	})
	if err != nil {
		return "could not list the pods!", nil
	}
	jobLock.Unlock()

	//log := make(chan bool)

	for range pods.Items {
		//channel or condition variables ?
		podLock.Lock()
		if !podEvents(clientset, courseName, creteadJob.Name) {
			condPod.Wait()
		}
		podLock.Unlock()
		// go podEvents(clientset, courseName, creteadJob.Name, log)
		//<-log
	}
	return podLog, nil
}

//DeleteObject deleting ..
func (k *K8s) DeleteObject(clientset kubernetes.Clientset, namespace string) error {
	ticker := time.NewTicker(24 * time.Hour)
	for {
		<-ticker.C
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
		} else {
			return nil
		}
	}
}

//PodLogs returns ...
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

	str := buf.String()
	return str, nil
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
			//fmt.Printf("Type: %v\n", event.Type)
			pod, ok := event.Object.(*apiv1.Pod)

			if !ok {
				panic("unexpected type")
			}
			if pod.Status.Phase == "Succeeded" {
				podLock.Lock()
				podLog, err = podLogs(*pod, clientset, namespace)
				if err != nil {
					panic(err)
				}
				st = true
				condPod.Signal()
				podLock.Unlock()
				/* podLog, err = podLogs(*pod, clientset,namespace)
				if err != nil {
					panic(err)
				}
				log <- true */
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
			//fmt.Printf("Type: %v\n", event.Type)
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

//jobsecrets create a secrets.. TODO
func jobsecrets(namespace string, clientset kubernetes.Clientset, pass string) {
	newSec := &apiv1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job-secret",
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"job-secret": []byte(pass),
		},
		Type: "Opaque",
	}
	_, err := clientset.CoreV1().Secrets(namespace).Create(newSec)
	if err != nil {
		panic(err)
	}
}

func resourceUsage() {
	var kubeconfig, master string //empty, assuming inClusterConfig
	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		panic(err)
	}

	cli, err := metrics.NewForConfig(config)

	podMetrics, err := cli.MetricsV1beta1().PodMetricses(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, podMetric := range podMetrics.Items {
		podContainers := podMetric.Containers
		for _, container := range podContainers {
			cpuQuantity, ok := container.Usage.Cpu().AsInt64()
			memQuantity, ok := container.Usage.Memory().AsInt64()
			if !ok {
				return
			}
			msg := fmt.Sprintf("Container Name: %s \n CPU usage: %d \n Memory usage: %d", container.Name, cpuQuantity, memQuantity)
			fmt.Println(msg)
		}

	}
}
