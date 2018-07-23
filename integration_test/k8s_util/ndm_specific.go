package k8sutil

import (
	"errors"
	"time"

	core_v1 "k8s.io/api/core/v1"

	cr "github.com/openebs/node-disk-manager/integration_test/common_resource"
)

const (
	// NdmNamespace is the namespace of the node-disk-manager
	NdmNamespace = core_v1.NamespaceDefault // Redeclared in ndm_util/ndm_util.go
)

// GetNdmPods returns pods of node-disk-manager if it can get them within timeout of 2 minutes.
// :return: kubernetes.client.models.v1_pod.V1Pod: node-disk-manager Pod object.
func GetNdmPods() ([]core_v1.Pod, error) {
	// Assumption: node-disk-manager pods runs under default namespace (k8s.io/api/core/v1.NamespaceDefault).
	// Assumption: Pod name starts with string "node-disk-manager".
	// Assumption: There is only one node-disk-manager pod (which is true for minikube).
	return cr.CitfInstance.K8S.GetPodsOrTimeout(NdmNamespace, "node-disk-manager", 2*time.Minute)
}

// GetContainerStateInNdmPod returns the state of the first container of the supplied index.
//    :param waitTimeUnit: maximum time duration to get the container's state.
//                       This method does not very strictly obey this param.
//    :return: k8s.io/api/core/v1.ContainerState: state of the container.
func GetContainerStateInNdmPod(waitTimeUnit time.Duration) (core_v1.ContainerState, error) {
	ndmPods, err := GetNdmPods()
	if err != nil {
		return core_v1.ContainerState{}, err
	} else if len(ndmPods) <= 0 {
		return core_v1.ContainerState{}, errors.New("no node-disk-manager pod is there")
	}
	// Assumption: There is only one pod of node-disk-manager.
	// Assumption: There is only one container in the node-disk-manager pod.
	return cr.CitfInstance.K8S.GetContainerStateByIndexInPodWithTimeout(&ndmPods[0], 0, waitTimeUnit)
}
