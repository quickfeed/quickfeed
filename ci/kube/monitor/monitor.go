package main

import (
	"fmt"
	"time"
	"flag"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	home       = homeDir()
	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "")
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
			return h
		}
	return os.Getenv("USERPROFILE") // windows
}

func main() {
	//nodeName1 := "kubernetes-slave01"
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	//var ctx context.Context
	mc, err := metrics.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	//nodeMc, err := mc.MetricsV1beta1().NodeMetricses().Get(nodeName1, metav1.GetOptions{})
	listNodeMc, err := mc.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{})
	if err != nil {
		 panic(err)
	 }
	//listPodMc, err := mc.MetricsV1beta1().PodMetricses(metav1.NamespaceAll).List(metav1.ListOptions{})
	ticker := time.NewTicker(time.Second * 30)
	fmt.Println("Inside main..")
	fmt.Println(len(listNodeMc.Items))

	for range ticker.C {
	fmt.Println("inside ticking..")
		for _, nodeMc := range listNodeMc.Items {
			fmt.Println("inside list node..")
			//ns := nodeMc.
			//for _, nm := range nodeMc {
			cpuQuantity, ok := nodeMc.Usage.Cpu().AsInt64()
			memQuantity, ok := nodeMc.Usage.Memory().AsInt64()
			if !ok {
				return
			}
			msg := fmt.Sprintf("Container Name: %s \n CPU usage: %d \n Memory usage: %d", nodeMc.Name, cpuQuantity, memQuantity)
			fmt.Println(msg)
		}
	}
	ticker.Stop()
}
