package util

import (
	"context"
	"io"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

var client = *NewK8SClient()

// NewK8SClient  This will create a new client for us to connect with k8s
// Be sure to add service account to your pod
func NewK8SClient() *kubernetes.Clientset {
	// creates the in-cluster configs
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Println("Error in creating in-cluster config,will try to read kubeconfig")
		// read kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", "aiops.yaml")
		if err != nil {
			log.Println("Error in reading kubeconfig", err)
		}
	}

	log.Println("Successfully read kubeconfig")
	// creates the client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println("Error in creating client", err)
		panic(err.Error())
	}
	log.Println("Successfully created client")
	return client
}

// GetPodLog get pod log
func GetPodLog(podName string, namespace string, limit *int64) (io.ReadCloser, error) {
	logOptions := &v1.PodLogOptions{
		Follow:    true,
		TailLines: limit,
	}
	req := client.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
	return req.Stream(context.Background())
}
