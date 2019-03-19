package k8s

import (
	"github.com/openebs/node-disk-manager/integration_tests/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeConfigPath *string
	clientSet      *kubernetes.Clientset
)

// Generate the clientset to connect to the k8s cluster
// from the config file
func GetClientSet() (*kubernetes.Clientset, error) {
	kubeConfigPath, err := utils.GetConfigPath()
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
