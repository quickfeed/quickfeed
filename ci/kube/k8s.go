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
	logs    string
	m       sync.Mutex
	waiting sync.Cond = *sync.NewCond(&m)
	active  chan bool
)

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }

//CreateJob runs the rescieved push from repository on the podes in our 3 nodes.
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

	kj, err := jobsClient.Create(confJob)
	if err != nil {
		return "", err
	}

	m.Lock()
	jobEvents(*kj, clientset, "agcicd", int32(1))
	if <-active != true {
		waiting.Wait()
	}
	m.Unlock()

	pods, err := clientset.CoreV1().Pods("agcicd").List(metav1.ListOptions{
		LabelSelector: "job-name=" + ("cijob" + id),
	})

	fmt.Println(len(pods.Items))

	for _, pod := range pods.Items {
		k.PodEvents(pod, clientset, "agcicd")
		logs = k.PodLogs(pod, clientset, "agcicd")
	}

	return logs, nil
}

//DeleteObject deleting ..
func (k *K8s) DeleteObject(pod apiv1.Pod, clientset kubernetes.Clientset, namespace string, kubeJob string) error {
	/* err := clientset.CoreV1().Pods("agcicd").Delete(pod.Name, &metav1.DeleteOptions{GracePeriodSeconds: int64Ptr(40)})
	if err != nil {
		panic(err)
	} */
	err := clientset.BatchV1().Jobs(namespace).Delete(kubeJob, &metav1.DeleteOptions{GracePeriodSeconds: int64Ptr(30)})
	if err != nil {
		return err
	}
	return nil
}

//PodLogs returns ...
func (k *K8s) PodLogs(pod apiv1.Pod, clientset *kubernetes.Clientset, namespace string) string {
	podLogOpts := apiv1.PodLogOptions{}

	req := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &podLogOpts)
	//TODO
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
func (k *K8s) PodEvents(pod apiv1.Pod, clientset *kubernetes.Clientset, namespace string) {
	//name := pod.GetName()
	watch, err := clientset.CoreV1().Pods(namespace).Watch(metav1.ListOptions{})
	if err != nil {
		fmt.Println("nothing to watch")
		log.Fatal(err.Error())
	}
	//go func() {
	//fmt.Println("INSIDE Goroutine")
	for event := range watch.ResultChan() {
		fmt.Printf("Type: %v\n", event.Type)
		p, ok := event.Object.(*v1.Pod)

		if !ok {
			log.Fatal("unexpected type")
		}
		if p.Status.Phase == apiv1.PodSucceeded {
			fmt.Println("notify succ..")
			break
		}

		if p.Status.Phase == apiv1.PodFailed {
			fmt.Println("POD FAILED manual print") //TODO: What to do? delete and run the job again?
			break
		}
	}
	//}()
}

func jobEvents(job batchv1.Job, clientset *kubernetes.Clientset, namespace string, nrOfPods int32) {
	active = make(chan bool)
	watch, err := clientset.BatchV1().Jobs(namespace).Watch(metav1.ListOptions{})
	if err != nil {
		fmt.Println("nothing to watch")
		log.Fatal(err.Error())
	}
	go func() {
		for event := range watch.ResultChan() {
			fmt.Printf("Type: %v\n", event.Type)
			j, ok := event.Object.(*batchv1.Job)
			if !ok {
				log.Fatal("unexpected type")
			}
			if j.Status.Active == nrOfPods {
				active <- true
				m.Lock()
				waiting.Signal()
				m.Unlock()
				fmt.Println("job active")
				break
			}
		}
	}()
}
