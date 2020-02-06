//Kuberntes interface
package kube


import (

	"fmt"



	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/client-go/util/retry"

	"github.com/autograde/aguis/ci"


	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"


	//corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	v1 "k8s.io/api/core/v1"

)

// K8s is an implementation of the CI interface using K8s.
type K8s struct {
	queue PriorityQueue
}



//RunToNodes runs the rescieved push from repository on the podes in our 3 nodes.
func (k *K8s) RunToNodes(d ci.Docker){
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	
	//deploymentsClient := clientset.AppsV1().Deployments("agcicd")

	

	//when a new jobs come run in it in a new pod?
	//when done delete the pod? result?
	pod, err := clientset.CoreV1().Pods("default").Create(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myPod",
			Namespace: v1.NamespaceDefault,
		},
	})
	
}


//Result returns the result of recently push that are executed on the nodes
func (k *K8s) Result() string{ 
	result:= ""
	return result
}

//CleanUp cleans up the pods that are done with running of the recently push
func (k *K8s) CleanUp(){}


func int32Ptr(i int32) *int32 { return &i }




//UpdateDeployment updates the deployment if some changes accours
func (k *K8s) UpdateDeployment(lastImage string) {

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	
	deploymentsClient := clientset.AppsV1().Deployments("agcicd")

	// Update Deployment
	fmt.Println("Updating deployment...")
	//    Two options to Update() this Deployment:
	//
	//    1. Modify the "deployment" variable and call: Update(deployment).
	//       This works like the "kubectl replace" command and it overwrites/loses changes
	//       made by other clients between you Create() and Update() the object.
	//    2. Modify the "result" returned by Get() and retry Update(result) until
	//       you no longer get a conflict error. This way, you can preserve changes made
	//       by other clients between Create() and Update(). This is implemented below
	//			 using the retry utility package included with client-go. (RECOMMENDED)
	//
	// More Info:
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
				
		result, getErr := deploymentsClient.Get("agcicd", metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get latest version of Deployment: %v", getErr))
		}
		result.Spec.Replicas = int32Ptr(1)                                       // reduce replica count
		result.Spec.Template.Spec.Containers[0].Image = lastImage // change image version
		_, updateErr := deploymentsClient.Update(result)
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}

}

//Deploy deploys... 
//TODO what we are using this for?
func (k *K8s) Deploy(lastImage string) *appsv1.Deployment{
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-dep",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: lastImage,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 8080,
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}