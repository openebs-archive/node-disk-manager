package k8s

import (
	"github.com/openebs/node-disk-manager/integration_tests/utils"
	"github.com/openebs/node-disk-manager/pkg/apis"
	_ "github.com/openebs/node-disk-manager/pkg/apis"
	_ "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	//"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	namespace = "default"
)

type k8sClient struct {
	ClientSet     *kubernetes.Clientset
	APIextClient  *apiextensionsclient.Clientset
	RunTimeClient client.Client
}

// Generate the clientset to connect to the k8s cluster
// from the config file
func GetClientSet() (k8sClient, error) {
	clientSet := k8sClient{}
	kubeConfigPath, err := utils.GetConfigPath()
	if err != nil {
		return clientSet, err
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return clientSet, err
	}
	// clientSet-go clientSet
	clientSet.ClientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		return clientSet, err
	}

	// clientSet for creating CRDs
	clientSet.APIextClient, err = apiextensionsclient.NewForConfig(config)
	if err != nil {
		return clientSet, err
	}

	// controller-runtime clientSet
	mgr, err := manager.New(config, manager.Options{Namespace: namespace})
	if err != nil {
		return clientSet, err
	}

	// add to scheme
	scheme := mgr.GetScheme()
	if err = apis.AddToScheme(scheme); err != nil {
		return clientSet, err
	}

	clientSet.RunTimeClient, err = client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return clientSet, err
	}
	return clientSet, nil
}
