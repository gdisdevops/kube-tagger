package util

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewKubernetesClient(kubeconfig string) (kubernetes.Interface, error) {
	config, err := getClientConfig(kubeconfig)
	if err != nil {
		log.Error("Failed to configure k8s client")
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error("Failed to create k8s clientset", err)
		return nil, err
	}

	return clientset, nil
}

func getClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		log.Debug("Use kubeconfig provided by commandline flag")
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	log.Debug("Use in-cluster k8s configuration")
	return rest.InClusterConfig()
}
