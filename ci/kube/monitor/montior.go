package main

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

func main() {
	//nodeName1 := "kubernetes-slave01"
	var kubeconfig, master string //empty, assuming inClusterConfig
	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
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
	//listPodMc, err := mc.MetricsV1beta1().PodMetricses(metav1.NamespaceAll).List(metav1.ListOptions{})
	ticker := time.NewTicker(time.Second * 30)
	for range ticker.C {
		for _, nodeMc := range listNodeMc.Items {
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
