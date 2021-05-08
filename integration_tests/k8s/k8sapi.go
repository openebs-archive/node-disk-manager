/*
Copyright 2019 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	apiextensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// The wait time for all k8s API related operations
const k8sWaitTime = 40 * time.Second

// The wait time for reconciliation loop to run
const k8sReconcileTime = 10 * time.Second

// ListPodStatus returns the list of all pods in the given Namespace along
// with their status
func (c K8sClient) ListPodStatus() (map[string]string, error) {
	pods := make(map[string]string)
	podList, err := c.ClientSet.CoreV1().Pods(Namespace).List(context.TODO(), metav1.ListOptions{})
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
func (c K8sClient) ListNodeStatus() (map[string]string, error) {
	nodes := make(map[string]string)
	nodeList, err := c.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, node := range nodeList.Items {
		nodes[node.Name] = string(node.Status.Phase)
	}
	return nodes, nil
}

// ListBlockDevices returns list of BlockDeviceCR in the cluster
func (c K8sClient) ListBlockDevices() (*apis.BlockDeviceList, error) {
	bdList := &apis.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	err := c.RunTimeClient.List(context.TODO(), bdList)
	if err != nil {
		return nil, fmt.Errorf("cannot list blockdevices. Error :%v", err)
	}
	return bdList, nil
}

// GetBlockDeviceClaim from Namespace and name
func (c K8sClient) GetBlockDeviceClaim(NameSpace, Name string) (*apis.BlockDeviceClaim, error) {
	bdc := &apis.BlockDeviceClaim{}
	err := c.RunTimeClient.Get(context.TODO(), client.ObjectKey{
		Name:      Name,
		Namespace: NameSpace,
	}, bdc)
	if err != nil {
		return nil, err
	}
	return bdc, nil
}

// ListBlockDeviceClaims returns list of BlockDeviceClaims in the cluster
func (c K8sClient) ListBlockDeviceClaims() (*apis.BlockDeviceClaimList, error) {
	bdcList := &apis.BlockDeviceClaimList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDeviceClaim",
			APIVersion: "openebs.io/v1alpha1",
		},
	}
	err := c.RunTimeClient.List(context.TODO(), bdcList)
	if err != nil {
		return nil, fmt.Errorf("cannot list block device claims. Error :%v", err)
	}
	return bdcList, nil
}

// RestartPod the given pod
func (c K8sClient) RestartPod(name string) error {
	pods, err := c.ListPodStatus()
	if err != nil {
		return nil
	}
	for pod := range pods {
		if strings.Contains(pod, name) {
			return c.ClientSet.CoreV1().Pods(Namespace).Delete(context.Background(), pod, metav1.DeleteOptions{})
		}
	}
	return fmt.Errorf("could not find given pod")
}

// NewBDC creates a sample device claim which can be used for
// claiming a block device.
func NewBDC(bdcName string) *apis.BlockDeviceClaim {
	bdcResources := apis.DeviceClaimResources{
		Requests: make(map[corev1.ResourceName]resource.Quantity),
	}
	bdcSpec := apis.DeviceClaimSpec{
		Resources: bdcResources,
	}
	bdc := &apis.BlockDeviceClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDeviceClaim",
			APIVersion: "openebs.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: make(map[string]string),
			Name:   bdcName,
		},
		Spec: bdcSpec,
	}
	return bdc
}

// CreateNamespace creates a namespace
func (c K8sClient) CreateNamespace(namespace corev1.Namespace) error {
	err := c.RunTimeClient.Create(context.Background(), &namespace)
	return err
}

// DeleteNamespace deletes a namespace
func (c K8sClient) DeleteNamespace(namespace corev1.Namespace) error {
	err := c.RunTimeClient.Delete(context.Background(), &namespace)
	return err
}

// CreateConfigMap creates a config map
func (c K8sClient) CreateConfigMap(configMap corev1.ConfigMap) error {
	err := c.RunTimeClient.Create(context.Background(), &configMap)
	return err
}

// DeleteConfigMap deletes the config map
func (c K8sClient) DeleteConfigMap(configMap corev1.ConfigMap) error {
	err := c.RunTimeClient.Delete(context.Background(), &configMap)
	return err
}

// CreateServiceAccount creates a service account
func (c K8sClient) CreateServiceAccount(serviceAccount corev1.ServiceAccount) error {
	err := c.RunTimeClient.Create(context.Background(), &serviceAccount)
	return err
}

// DeleteServiceAccount deletes the service account
func (c K8sClient) DeleteServiceAccount(serviceAccount corev1.ServiceAccount) error {
	err := c.RunTimeClient.Delete(context.Background(), &serviceAccount)
	return err
}

// CreateClusterRole creates a cluster role
func (c K8sClient) CreateClusterRole(clusterRole rbacv1beta1.ClusterRole) error {
	err := c.RunTimeClient.Create(context.Background(), &clusterRole)
	return err
}

// DeleteClusterRole deletes the cluster role
func (c K8sClient) DeleteClusterRole(clusterRole rbacv1beta1.ClusterRole) error {
	err := c.RunTimeClient.Delete(context.Background(), &clusterRole)
	return err
}

// CreateClusterRoleBinding creates a rolebinding
func (c K8sClient) CreateClusterRoleBinding(clusterRoleBinding rbacv1beta1.ClusterRoleBinding) error {
	err := c.RunTimeClient.Create(context.Background(), &clusterRoleBinding)
	return err
}

// DeleteClusterRoleBinding deletes the role binding
func (c K8sClient) DeleteClusterRoleBinding(clusterrolebinding rbacv1beta1.ClusterRoleBinding) error {
	err := c.RunTimeClient.Delete(context.Background(), &clusterrolebinding)
	return err
}

// CreateCustomResourceDefinition creates a CRD
func (c K8sClient) CreateCustomResourceDefinition(customResourceDefinition apiextensionsV1.CustomResourceDefinition) error {
	_, err := c.APIextClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.Background(), &customResourceDefinition, metav1.CreateOptions{})
	return err
}

// DeleteCustomResourceDefinition deletes the CRD
func (c K8sClient) DeleteCustomResourceDefinition(customResourceDefinition apiextensionsV1.CustomResourceDefinition) error {
	err := c.APIextClient.ApiextensionsV1().CustomResourceDefinitions().Delete(context.Background(), customResourceDefinition.Name, metav1.DeleteOptions{})
	return err
}

// CreateDaemonSet creates a Daemonset
func (c K8sClient) CreateDaemonSet(daemonset appsv1.DaemonSet) error {
	daemonset.Namespace = Namespace
	err := c.RunTimeClient.Create(context.Background(), &daemonset)
	return err
}

// DeleteDaemonSet deletes the Daemonset
func (c K8sClient) DeleteDaemonSet(daemonset appsv1.DaemonSet) error {
	daemonset.Namespace = Namespace
	err := c.RunTimeClient.Delete(context.Background(), &daemonset, client.PropagationPolicy(metav1.DeletePropagationForeground))
	return err
}

// CreateDeployment creates a deployment
func (c K8sClient) CreateDeployment(deployment appsv1.Deployment) error {
	deployment.Namespace = Namespace
	err := c.RunTimeClient.Create(context.Background(), &deployment)
	return err
}

// DeleteDeployment deletes a deployment
func (c K8sClient) DeleteDeployment(deployment appsv1.Deployment) error {
	deployment.Namespace = Namespace
	err := c.RunTimeClient.Delete(context.Background(), &deployment)
	return err
}

// CreateBlockDeviceClaim creates a BDC
func (c K8sClient) CreateBlockDeviceClaim(claim *apis.BlockDeviceClaim) error {
	err := c.RunTimeClient.Create(context.Background(), claim)
	return err
}

// UpdateBlockDeviceClaim updates the BDC
func (c K8sClient) UpdateBlockDeviceClaim(claim *apis.BlockDeviceClaim) error {
	err := c.RunTimeClient.Update(context.Background(), claim)
	return err
}

// DeleteBlockDeviceClaim deletes a BDC
func (c K8sClient) DeleteBlockDeviceClaim(claim *apis.BlockDeviceClaim) error {
	err := c.RunTimeClient.Delete(context.Background(), claim)
	return err
}
