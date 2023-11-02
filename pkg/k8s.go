package pkg

import (
	"context"
	"errors"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log/slog"
)

var client = *NewK8SClient()
var kubeconfig = "config"

// NewK8SClient  This will create a new client for us to connect with k8s
// Be sure to add service account to your pod
func NewK8SClient() *kubernetes.Clientset {
	// creates the in-cluster configs
	config, err := rest.InClusterConfig()
	if err != nil {
		slog.Error("Error in creating in-cluster config,will try to read kubeconfig", "error", err)
		// read kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			slog.Error("Error in reading kubeconfig", "error", err)
		}
	}

	slog.Info("Successfully created cluster config")
	// creates the client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		slog.Error("Error in creating client", "error", err)
		panic(err.Error())
	}
	slog.Info("Successfully created client")
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

// GetJobLog get job log
func GetJobLog(jobName string, namespace string, limit *int64) (io.ReadCloser, error) {

	// find pod with label job-name=jobName
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "job-name=" + jobName,
	})
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 0 {
		return nil, errors.New("no pod found")
	}
	return GetPodLog(pods.Items[0].Name, namespace, limit)
}
