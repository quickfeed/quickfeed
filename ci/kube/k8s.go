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
	go jobsecrets(id, courseName, *clientset, kubeRandomSecret())

	//Define the configiration of the job object
	jobsClient := clientset.BatchV1().Jobs(courseName)
	confJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cijob" + id,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            int32Ptr(8),
			TTLSecondsAfterFinished: int32Ptr(5), //TODO: add ActiveDeadlineSeconds: int64Ptr(1000), terminate after 1000 sec ?
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
									"cpu":    resource.MustParse("100m"), //TODO: test by changing this to "2" and run 8 in parallell
									"memory": resource.MustParse("100Mi"),
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

	k.jobEvents(*createdJob, clientset, courseName, int32(1), createdJob.Name)
	k.jobWaitToActive()

	pods, err := clientset.CoreV1().Pods(courseName).List(metav1.ListOptions{
		LabelSelector: "job-name=" + createdJob.Name,
	})
	if err != nil {
		return "could not list the pods!", nil
	}

	for range pods.Items {
		k.podEvents(clientset, courseName, createdJob.Name)
		k.podWaitToSucc()
	}

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

//jobWaitToActive waits for the pod to success. The signal will be send from the k.podEvents function
func (k *K8s) jobWaitToActive() {
	isJobActive.L.Lock()
	for k.isActive == 0 {
		isJobActive.Wait()
	}
	k.isActive--
	isJobActive.L.Unlock()
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

//jobEvents watch the events of the jobs. If job is active, then sends signal to the jobWaitToActive() method.
func (k *K8s) jobEvents(job batchv1.Job, clientset *kubernetes.Clientset, namespace string, nrOfPods int32, kubejobname string) {
	watch, err := clientset.BatchV1().Jobs(namespace).Watch(metav1.ListOptions{
		LabelSelector: "job-name=" + kubejobname,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	go func() {
		for event := range watch.ResultChan() {
			j, ok := event.Object.(*batchv1.Job)
			if !ok {
				log.Fatal("unexpected type")
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

	str := buf.String()
	return str, nil
}

//podEvents watch the pods events, and send signal to the podWaitToSucc() method if pod successed.
func (k *K8s) podEvents(clientset *kubernetes.Clientset, namespace string, jobname string) {
	watch, err := clientset.CoreV1().Pods(namespace).Watch(metav1.ListOptions{
		LabelSelector: "job-name=" + jobname,
	})
	if err != nil {
		panic(err)
	}
	go func() {
		for event := range watch.ResultChan() {
			pod, ok := event.Object.(*apiv1.Pod)
			if !ok {
				panic("unexpected type")
			}
			if pod.Status.Phase == "Succeeded" {
				isPodDone.L.Lock()
				k.podLog, err = podLogs(*pod, clientset, namespace)
				if err != nil {
					panic(err)
				}
				k.isLogReady++
				isPodDone.Signal()
				isPodDone.L.Unlock()
			}
			if pod.Status.Phase == apiv1.PodFailed {
				fmt.Println("POD FAILED manual print") //TODO: What to do? delete and run the job again?
			}
		}
	}()
}

//jobsecrets create a secrets.. TODO
func jobsecrets(secretName string, namespace string, clientset kubernetes.Clientset, pass string) {
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
		panic(err)
	}
}

func kubeRandomSecret() string {
	randomness := make([]byte, 10)
	_, err := rand.Read(randomness)
	if err != nil {
		panic("couldn't generate randomness")
	}
	return fmt.Sprintf("%x", sha1.Sum(randomness))
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
