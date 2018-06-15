package common

import (
	"fmt"
	"os"
	"path/filepath"

	api_core_v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	typed_core_v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	// Install special auth plugins like GCP Plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
)

// Different phases of Namespace

// NsGoodPhases is an array of phases of the Namespace which are considered to be good
var NsGoodPhases = []api_core_v1.NamespacePhase{"Active"}

// Different phases of Pod

// PodWaitPhases is an array of phases of the Pod which are considered to be waiting
var PodWaitPhases = []string{"Pending"}

// PodGoodPhases is an array of phases of the Pod which are considered to be good
var PodGoodPhases = []string{"Running"}

// PodBadPhases is an array of phases of the Pod which are considered to be bad
var PodBadPhases = []string{"Error"}

// Different states of Pod

// PodWaitStates is an array of the states of the Pod which are considered to be waiting
var PodWaitStates = []string{"ContainerCreating", "Pending"}

// PodGoodStates is an array of the states of the Pod which are considered to be good
var PodGoodStates = []string{"Running"}

// PodBadStates is an array of the states of the Pod which are considered to be bad
var PodBadStates = []string{"CrashLoopBackOff", "ImagePullBackOff", "RunContainerError", "ContainerCannotRun"}

// IsNSinGoodPhase checks if supplied namespace is in good phase or not
// by matching phase of supplied namespace with pre-identified Good phase list (NsGoodPhases)
func IsNSinGoodPhase(namespace api_core_v1.Namespace) bool {
	for _, phase := range NsGoodPhases {
		if phase == namespace.Status.Phase {
			return true
		}
	}

	return false
}

// IsPodStateWait checks if supplied pod state is wait state or not
// by matching state of supplied pod state with pre-identified Wait states list (PodWaitStates)
func IsPodStateWait(podState string) bool {
	for _, state := range PodWaitStates {
		if state == podState {
			return true
		}
	}

	return false
}

// IsPodStateGood checks if supplied pod state is good or not
// by matching state of supplied pod state with pre-identified Good states list (PodGoodStates)
func IsPodStateGood(podState string) bool {
	for _, state := range PodGoodStates {
		if state == podState {
			return true
		}
	}

	return false
}

// GetClientConfig first tries to get a config object which uses the service account kubernetes gives to pods,
// if it is called from a process running in a kubernetes environment.
// Otherwise, it tries to build config from a default kubeconfig filepath if it fails, it fallback to the default config.
// Once it get the config, it returns the same.
func GetClientConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		if Debug {
			fmt.Printf("Unable to create config. Error: %+v\n", err)
		}
		err1 := err
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			err = fmt.Errorf("InClusterConfig as well as BuildConfigFromFlags Failed. Error in InClusterConfig: %+v\nError in BuildConfigFromFlags: %+v", err1, err)
			return nil, err
		}
	}

	return config, nil
}

// GetClientset first tries to get a config object which uses the service account kubernetes gives to pods,
// if it is called from a process running in a kubernetes environment.
// Otherwise, it tries to build config from a default kubeconfig filepath if it fails, it fallback to the default config.
// Once it get the config, it creates a new Clientset for the given config and returns the clientset.
func GetClientset() (*kubernetes.Clientset, error) {
	config, err := GetClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		err = fmt.Errorf("Failed creating clientset. Error: %+v", err)
		return nil, err
	}

	return clientset, nil
}

// GetRESTClient first tries to get a config object which uses the service account kubernetes gives to pods,
// if it is called from a process running in a kubernetes environment.
// Otherwise, it tries to build config from a default kubeconfig filepath if it fails, it fallback to the default config.
// Once it get the config, it
func GetRESTClient() (*rest.RESTClient, error) {
	config, err := GetClientConfig()
	if err != nil {
		return &rest.RESTClient{}, err
	}

	return rest.RESTClientFor(config)
}

// GetCoreV1API returns a k8s.io/client-go/kubernetes/typed/core/v1.CoreV1Interface object
// and nil as error in case of no error occures when it tries to make a clientset
// otherwise it returns nil for k8s.io/client-go/kubernetes/typed/core/v1.CoreV1Interface
// and the error occured
func GetCoreV1API() (typed_core_v1.CoreV1Interface, error) {
	clientset, err := GetClientset()
	if err != nil {
		return nil, err
	}
	return clientset.CoreV1(), nil
}
