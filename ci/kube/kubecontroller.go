package ci

import (
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// uses the current context in kubeconfig
	// path-to-kubeconfig -- for example, /root/.kube/config
	config, _ := clientcmd.BuildConfigFromFlags("", "$HOME/.kube/config")
	// creates the clientset
	clientset, _ := kubernetes.NewForConfig(config)

	//acess the API to list the nodes
	/* 	serv, err := clientset.CoreV1().Nodes().List(v1.ListOptions{})
	if err !=nil {
		fmt.Println("Could'nt find any nodes!")
	}
	fmt.Printf("There are %d nodes in the cluster\n", len(serv )) */

	// access the API to list pods
	pods, err := clientset.CoreV1().Pods("").List(v1.ListOptions{})
	if err != nil {
		fmt.Println("Could'nt find any pods!")
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
}
