package k8s

import (
	"context"
	"fmt"
	"github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// The wait time for all k8s API related operations
const k8sWaitTime = 30 * time.Second

// ListPodStatus returns the list of all pods in the given namespace along
// with their status
func (c k8sClient) ListPodStatus() (map[string]string, error) {
	pods := make(map[string]string)
	podList := &v1.PodList{}
	podList, err := c.ClientSet.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, pod := range podList.Items {
		pods[pod.Name] = string(pod.Status.Phase)
	}
	return pods, nil
}

// ListNodeStatus returns list of all nodes(node name) in the cluster along with
// their status
func (c k8sClient) ListNodeStatus() (map[string]string, error) {
	nodes := make(map[string]string)
	nodeList := &v1.NodeList{}
	nodeList, err := c.ClientSet.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, node := range nodeList.Items {
		nodes[node.Name] = string(node.Status.Phase)
	}
	return nodes, nil
}

// ListDisk returns list of DiskCR in the cluster
func (c k8sClient) ListDisk() (*v1alpha1.DiskList, error) {
	diskList := &v1alpha1.DiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Disk",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	var err error
	err = c.RunTimeClient.List(context.TODO(), &client.ListOptions{}, diskList)
	if err != nil {
		return nil, fmt.Errorf("cannot list disks. Error :%v", err)
	}
	return diskList, nil
}

// ListBlockDevices returns list of BlockDeviceCR in the cluster
func (c k8sClient) ListBlockDevices() (*v1alpha1.BlockDeviceList, error) {
	bdList := &v1alpha1.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	var err error
	err = c.RunTimeClient.List(context.TODO(), &client.ListOptions{}, bdList)
	if err != nil {
		return nil, fmt.Errorf("cannot list disks. Error :%v", err)
	}
	return bdList, nil
}

// CreateConfigMap creates a config map
func (c k8sClient) CreateConfigMap(configMap v1.ConfigMap) error {
	err := c.RunTimeClient.Create(context.Background(), &configMap)
	return err
}

// DeleteConfigMap deletes the config map
func (c k8sClient) DeleteConfigMap(configMap v1.ConfigMap) error {
	err := c.RunTimeClient.Delete(context.Background(), &configMap)
	return err
}

// CreateServiceAccount creates a service account
func (c k8sClient) CreateServiceAccount(serviceAccount v1.ServiceAccount) error {
	err := c.RunTimeClient.Create(context.Background(), &serviceAccount)
	return err
}

// DeleteServiceAccount deletes the service account
func (c k8sClient) DeleteServiceAccount(serviceAccount v1.ServiceAccount) error {
	err := c.RunTimeClient.Delete(context.Background(), &serviceAccount)
	return err
}

// CreateClusterRole creates a cluster role
func (c k8sClient) CreateClusterRole(clusterRole rbacv1beta1.ClusterRole) error {
	err := c.RunTimeClient.Create(context.Background(), &clusterRole)
	return err
}

// DeleteClusterRole deletes the cluster role
func (c k8sClient) DeleteClusterRole(clusterRole rbacv1beta1.ClusterRole) error {
	err := c.RunTimeClient.Delete(context.Background(), &clusterRole)
	return err
}

// CreateClusterRoleBinding creates a rolebinding
func (c k8sClient) CreateClusterRoleBinding(clusterRoleBinding rbacv1beta1.ClusterRoleBinding) error {
	err := c.RunTimeClient.Create(context.Background(), &clusterRoleBinding)
	return err
}

// DeleteClusterRoleBinding deletes the role binding
func (c k8sClient) DeleteClusterRoleBinding(clusterrolebinding rbacv1beta1.ClusterRoleBinding) error {
	err := c.RunTimeClient.Delete(context.Background(), &clusterrolebinding)
	return err
}

// CreateCustomResourceDefinition creates a CRD
func (c k8sClient) CreateCustomResourceDefinition(customResourceDefinition apiextensionsv1beta1.CustomResourceDefinition) error {
	_, err := c.APIextClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(&customResourceDefinition)
	return err
}

// DeleteCustomResourceDefinition deletes the CRD
func (c k8sClient) DeleteCustomResourceDefinition(customResourceDefinition apiextensionsv1beta1.CustomResourceDefinition) error {
	err := c.APIextClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(customResourceDefinition.Name, &metav1.DeleteOptions{})
	return err
}

// CreateDaemonSet creates a Daemonset
func (c k8sClient) CreateDaemonSet(daemonset v1beta1.DaemonSet) error {
	daemonset.Namespace = namespace
	err := c.RunTimeClient.Create(context.Background(), &daemonset)
	return err
}

// DeleteDaemonSet deletes the Daemonset
func (c k8sClient) DeleteDaemonSet(daemonset v1beta1.DaemonSet) error {
	daemonset.Namespace = namespace
	err := c.RunTimeClient.Delete(context.Background(), &daemonset, client.PropagationPolicy(metav1.DeletePropagationForeground))
	return err
}
