package k8sutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/golang/glog"
	. "github.com/openebs/node-disk-manager/integration_test/common"
	core_v1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// GetAllNamespacesCoreV1NamespaceArray returns V1NamespaceList of all the namespaces.
// :return kubernetes.client.models.v1_namespace_list.V1NamespaceList: list of namespaces.
func GetAllNamespacesCoreV1NamespaceArray() ([]core_v1.Namespace, error) {
	api, err := GetCoreV1API()
	if err != nil {
		return nil, err
	}
	nsList, err := api.Namespaces().List(meta_v1.ListOptions{})
	return nsList.Items, err
}

// GetAllNamespacesMap returns list of the names of all the namespaces.
// :return: map[string]core_v1.Namespace: map of namespaces where key is namespace name (str)
// and value is corresponding k8s.io/api/core/v1.Namespace object.
func GetAllNamespacesMap() (map[string]core_v1.Namespace, error) {
	namespacesList, err := GetAllNamespacesCoreV1NamespaceArray()
	if err != nil {
		return nil, err
	}

	namespaces := map[string]core_v1.Namespace{}
	for _, ns := range namespacesList {
		namespaces[ns.Name] = ns
	}
	return namespaces, nil
}

// GetPod returns first Pod object which has a prefix specified in its name in the given namespace.
// :return: kubernetes.client.models.v1_pod.V1Pod: Pod object.
func GetPod(namespace, podNamePrefix string) (core_v1.Pod, error) {
	api, err := GetCoreV1API()
	if err != nil {
		return core_v1.Pod{}, err
	}

	// Try to get the pod for 10 times as sometime code reaches
	// when pod is not even in ContainerCreating state
	i := 0
	var thePod core_v1.Pod
	for reflect.DeepEqual(thePod, core_v1.Pod{}) && i < 10 {
		time.Sleep(2 * time.Second)

		// List pods
		pods, err := api.Pods(namespace).List(meta_v1.ListOptions{})
		if err != nil {
			fmt.Printf("Error occured: %+v\n", err)
		}

		// Find the Pod
		if Debug {
			fmt.Println(strings.Repeat("*", 80))
			fmt.Printf("Current pods in %q namespace are:\n", namespace)
		}
		for _, pod := range pods.Items {
			if Debug {
				fmt.Println("Complete Pod name is:", pod.Name)
			}
			if strings.HasPrefix(pod.Name, podNamePrefix) {
				thePod = pod
				break
			}
		}
		if Debug {
			fmt.Println(strings.Repeat("*", 80))
		}
		i++
	}
	if reflect.DeepEqual(thePod, core_v1.Pod{}) {
		glog.Fatal("Failed getting NDM-Pod in given time.")
	}

	return thePod, nil
}

// ReloadPod reloads the state of the pod supplied and return the recent one
func ReloadPod(pod core_v1.Pod) (core_v1.Pod, error) {
	return GetPod(pod.Namespace, pod.Name)
}

// GetPodPhase returns phase of the pod passed as an k8s.io/api/core/v1.PodPhase object.
//		:param k8s.io/api/core/v1.Pod pod: pod object for which you want to get phase.
//		:return: k8s.io/api/core/v1.PodPhase: phase of the pod.
func GetPodPhase(pod core_v1.Pod) core_v1.PodPhase {
	return pod.Status.Phase
}

// GetPodPhaseStr returns phase of the pod passed in string format.
//		:param k8s.io/api/core/v1.Pod pod: pod object for which you want to get phase.
//		:return: str: phase of the pod.
func GetPodPhaseStr(pod core_v1.Pod) string {
	return string(GetPodPhase(pod))
}

// GetContainerStateInPod returns the state of the container of supplied index of the supplied Pod.
//    :param containerIndex: index of the container for which you want state.
//    :param timeout: maximum time duration to get the container's state.
//                       This method does not very strictly obey this param.
//    :return: k8s.io/api/core/v1.ContainerState: state of the container.
func GetContainerStateInPod(pod core_v1.Pod, containerIndex int, timeout time.Duration) (core_v1.ContainerState, error) {
	var err error
	startTime := time.Now()
	for reflect.DeepEqual(pod.Status.ContainerStatuses, []core_v1.ContainerStatus(nil)) && time.Since(startTime) < timeout {
		time.Sleep(time.Second)
		pod, err = ReloadPod(pod)
		if err != nil {
			return core_v1.ContainerState{}, err
		}
	}
	if time.Since(startTime) >= timeout {
		return core_v1.ContainerState{}, fmt.Errorf("Pod %q of namespace %q had no container till %v", pod.Name, pod.Namespace, timeout)
	}

	for len(pod.Status.ContainerStatuses) <= containerIndex && time.Since(startTime) < timeout {
		time.Sleep(time.Second)
		pod, err = ReloadPod(pod)
		if err != nil {
			return core_v1.ContainerState{}, err
		}
	}
	if time.Since(startTime) >= timeout {
		return core_v1.ContainerState{}, fmt.Errorf("Pod did not had %d containers till %v", containerIndex+1, timeout)
	}

	return pod.Status.ContainerStatuses[containerIndex].State, nil
}

// GetNodes returns a list of all the nodes.
//    :return: slice: list of nodes (slice of k8s.io/api/core/v1.Node array).
func GetNodes() (nodeNames []core_v1.Node, err error) {
	nodeNames = []core_v1.Node{}

	api, err := GetCoreV1API()
	if err != nil {
		return
	}

	// To handle latency it tries 10 times each after 1 second of wait
	waited := 0
	for waited < 10 {
		nodeList, err := api.Nodes().List(meta_v1.ListOptions{})
		if err != nil {
			break
		} else if len(nodeList.Items) == 0 {
			time.Sleep(time.Second)
			waited++
			continue
		}
		nodeNames = nodeList.Items
		break
	}

	return
}

// GetNodeNames returns a list of the name of all the nodes.
//    :return: slice: list of node names (slice of string array).
func GetNodeNames() (nodeNames []string, err error) {
	nodeNames = []string{}

	nodes, err := GetNodes()
	if err != nil {
		return
	}
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.Name)
	}

	return
}

// TODO: Write a function to label the node
// LabelNode label the node with the given key and value.
//    :param string node_name: Name of the node.
//    :param string key: Key of the label.
//    :param string value: Value of the label.
//    :return: error: if any error occured or nil otherwise.
// func LabelNode(nodeName, key, value string) error { return fmt.Errorf("Not Implemented") }

// GetDaemonset returns the k8s.io/api/extensions/v1beta1.DaemonSet for the name supplied.
func GetDaemonset(daemonsetName, daemonsetNamespace string) (v1beta1.DaemonSet, error) {
	clientset, err := GetClientset()

	daemonsetClient := clientset.ExtensionsV1beta1().DaemonSets(daemonsetNamespace)
	ds, err := daemonsetClient.Get(daemonsetName, meta_v1.GetOptions{})
	if err != nil {
		return v1beta1.DaemonSet{}, err
	}
	return *ds, nil
}

// ApplyDSFromManifestStruct Creates a Daemonset from the manifest supplied
func ApplyDSFromManifestStruct(manifest v1beta1.DaemonSet) (v1beta1.DaemonSet, error) {
	clientset, err := GetClientset()

	if manifest.Namespace == "" {
		manifest.Namespace = core_v1.NamespaceDefault
	}
	daemonsetClient := clientset.ExtensionsV1beta1().DaemonSets(manifest.Namespace)
	ds, err := daemonsetClient.Create(&manifest)
	if err != nil {
		return v1beta1.DaemonSet{}, err
	}
	return *ds, nil
}

// GetDaemonsetStructFromYamlBytes returns k8s.io/api/extensions/v1beta1.DaemonSet
// for the yaml supplied
func GetDaemonsetStructFromYamlBytes(yamlBytes []byte) (v1beta1.DaemonSet, error) {
	ds := v1beta1.DaemonSet{}

	jsonBytes, err := ConvertYAMLtoJSON(yamlBytes)
	if err != nil {
		return ds, fmt.Errorf("Error while Converting yaml string into Daemonset Structure. Error: %+v", err)
	}

	err = json.Unmarshal(jsonBytes, &ds)
	if err != nil {
		return ds, fmt.Errorf("Error occured while marshaling into Daemonset struct. Error: %+v", err)
	}

	return ds, nil
}

// TODO: Write a function to apply the YAML with the help of client-go
// YAMLApply apply the yaml specified by the argument.
//    :param str yamlPath: Path of the yaml file that is to be applied.
// func YAMLApplyAPI(yamlPath string) error { return fmt.Errorf("Not Implemented") }

// YAMLApply apply the yaml specified by the argument.
//    :param str yamlPath: Path of the yaml file that is to be applied.
func YAMLApply(yamlPath string) error {
	// TODO: Try using API call first. i.e. Using client-go

	err := RunCommand("kubectl apply -f " + yamlPath)
	if err != nil {
		glog.Errorf("Error occured while applying the %s. Error: %+v", yamlPath, err)
		return fmt.Errorf("Failed applying %s", yamlPath)
	}
	return nil
}

// ExecToPodThroughAPI uninterractively exec to the pod with the command specified.
// :param string command: list of the str which specify the command.
// :param string pod_name: Pod name
// :param string namespace: namespace of the Pod.
// :param io.Reader stdin: Standerd Input if necessary, otherwise `nil`
// :return: string: Output of the command. (STDOUT)
//          string: Errors. (STDERR)
//           error: If any error has occured otherwise `nil`
// TODO: Need to fix the error (in exec.Steam) (unable to upgrade connection: you must specify at least 1 of stdin, stdout, stderr)
func ExecToPodThroughAPI(command, podName, namespace string, stdin io.Reader) (string, string, error) {
	config, err := GetClientConfig()
	if err != nil {
		return "", "", err
	}
	clientset, err := GetClientset()
	if err != nil {
		return "", "", err
	}

	req := clientset.Core().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")
	scheme := runtime.NewScheme()
	parameterCodec := runtime.NewParameterCodec(scheme)
	req.VersionedParams(&core_v1.PodExecOptions{
		Command: strings.Fields(command),
		Stdin:   stdin != nil,
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
	}, parameterCodec)

	fmt.Println("Request URL: ", req.URL().String())

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("Error while creating Executor. Error: %+v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	return stdout.String(), stderr.String(), fmt.Errorf("Error in Stream. Error: %+v", err)
}

// ExecToPod uninterractively exec to the pod with the command specified
// first through API with `stdin` param as `nil`, if it fails then it uses `kubectl exec`
// :param string command: list of the str which specify the command.
// :param string pod_name: Pod name
// :param string namespace: namespace of the Pod.
// :return: string: Output of the command. (STDOUT)
//           error: If any error has occured otherwise `nil`
func ExecToPod(command, podName, namespace string) (string, error) {
	stdout, stderr, err := ExecToPodThroughAPI(command, podName, namespace, nil)
	if err == nil {
		return stdout, nil
	}

	// When Exec trhough API fails
	glog.Warningf("Error while exec into Pod through API. Stderr: %q. Error: %+v", stderr, err)
	return ExecCommand("kubectl -n " + namespace + " exec " + podName + " -- " + command)
}

// GetLog returns the log of the pod.
// :param string pod_name: Name of the pod. (required)
// :param string namespace: Namespace of the pod. (required)
// :return: string: Log of the pod specified.
//           error: If an error has occured, otherwise `nil`
// TODO: Fix in API call (Error: GroupVersion is required when initializing a RESTClient)
func GetLog(podName, namespace string) (string, error) {
	// We can't declare a variable somewhere which can be skipped by goto
	var req *rest.Request
	var readCloser io.ReadCloser
	var err error
	bytes := []byte{}
	restClient, err := GetRESTClient()
	if err != nil {
		goto use_kubectl
	}

	req = restClient.Get().Namespace(namespace).Name(podName).Resource("pods").SubResource("log")
	readCloser, err = req.Stream()
	if err != nil {
		goto use_kubectl
	}

	defer readCloser.Close()
	_, err = readCloser.Read(bytes)
	if err == nil {
		return string(bytes), nil
	}

use_kubectl:
	glog.Warningf("Error while getting log with API call. Error: %+v", err)

	return ExecCommand("kubectl -n " + namespace + " logs " + podName)
}
