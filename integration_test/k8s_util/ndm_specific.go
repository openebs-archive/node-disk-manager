package k8sutil

import (
	"time"

	core_v1 "k8s.io/api/core/v1"
)

const (
	// NdmNamespace is the namespace of the node-disk-manager
	NdmNamespace = core_v1.NamespaceDefault // Redeclared in ndm_util/ndm_util.go
)

// GetNdmPod returns Pod object of node-disk-manager.
// :return: kubernetes.client.models.v1_pod.V1Pod: node-disk-manager Pod object.
func GetNdmPod() (core_v1.Pod, error) {
	// Assumption: node-disk-manager pods runs under default namespace (k8s.io/api/core/v1.NamespaceDefault).
	// Assumption: Pod name starts with string "node-disk-manager".
	// Assumption: There is only one node-disk-manager pod (which is true for minikube).
	return GetPod(NdmNamespace, "node-disk-manager")
}

// GetContainerStateInNdmPod returns the state of the first container of the supplied index.
//    :param waitTimeUnit: maximum time duration to get the container's state.
//                       This method does not very strictly obey this param.
//    :return: k8s.io/api/core/v1.ContainerState: state of the container.
func GetContainerStateInNdmPod(waitTimeUnit time.Duration) (core_v1.ContainerState, error) {
	ndmPod, err := GetNdmPod()
	if err != nil {
		return core_v1.ContainerState{}, err
	}
	// Assumption: There is only one container in the node-disk-manager pod.
	return GetContainerStateInPod(ndmPod, 0, waitTimeUnit)
}
