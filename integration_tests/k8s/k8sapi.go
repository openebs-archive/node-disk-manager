package k8s

import (
	"context"
	"fmt"
	"github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// The wait time for all k8s API related operations
const k8sWaitTime = 30 * time.Second

// GetPods returns the list of all pods in the given namespace along
// with their status
func GetPods(clientset *kubernetes.Clientset) (map[string]string, error) {
	pods := make(map[string]string)
	podList := &v1.PodList{}
	podList, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, pod := range podList.Items {
		pods[pod.Name] = string(pod.Status.Phase)
	}
	return pods, nil
}

// GetNodes returns list of all nodes(node name) in the cluster along with
// their status
func GetNodes(clientSet *kubernetes.Clientset) (map[string]string, error) {
	nodes := make(map[string]string)
	nodeList := &v1.NodeList{}
	nodeList, err := clientSet.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, node := range nodeList.Items {
		nodes[node.Name] = string(node.Status.Phase)
	}
	return nodes, nil
}

// GetDiskList returns list of DiskCR in the cluster
func GetDiskList(clientSet k8sClient) (*v1alpha1.DiskList, error) {
	diskList := &v1alpha1.DiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Disk",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	var err error
	err = clientSet.RunTimeClient.List(context.TODO(), &client.ListOptions{}, diskList)
	if err != nil {
		return nil, fmt.Errorf("cannot list disks. Error :%v", err)
	}
	return diskList, nil
}

// CreateNDMYAML creates all the objects specified in the NDM operator YAML.
// Each resource object is generated from yaml file in ../yamls/ and parsed into
// the required type. ConfigMap, ServiceAccount, ClusterRole, ClusterRoleBinding,
// CustomResourceDefinition and DaemonSet are created
func CreateNDMYAML(clientset k8sClient) error {
	var err error

	// creating NDM ConfigMap
	configmap, err := GetConfigMap()
	if err != nil {
		return err
	}
	err = CreateConfigMap(clientset.RunTimeClient, configmap)
	if err != nil {
		return err
	}

	// creating NDM serviceAccount
	serviceaccount, err := GetServiceAccount()
	if err != nil {
		return err
	}
	err = CreateServiceAccount(clientset.RunTimeClient, serviceaccount)
	if err != nil {
		return err
	}

	// creating NDM cluster role
	clusterrole, err := GetClusterRole()
	if err != nil {
		return err
	}
	err = CreateClusterRole(clientset.RunTimeClient, clusterrole)
	if err != nil {
		return err
	}

	// creating NDM clusterrole binding
	clusterrolebinding, err := GetClusterRoleBinding()
	if err != nil {
		return err
	}
	err = CreateClusterRoleBinding(clientset.RunTimeClient, clusterrolebinding)
	if err != nil {
		return err
	}

	// creating disk crd
	diskcrd, err := GetCustomResourceDefinition(DiskCRDYAML)
	if err != nil {
		return err
	}
	err = CreateCustomResourceDefinition(clientset.APIextClient, diskcrd)
	if err != nil {
		return err
	}

	// creating device crd
	devicecrd, err := GetCustomResourceDefinition(DeviceCRDYAML)
	if err != nil {
		return err
	}
	err = CreateCustomResourceDefinition(clientset.APIextClient, devicecrd)
	if err != nil {
		return err
	}

	// creating device request crd
	deviceRequestcrd, err := GetCustomResourceDefinition(DeviceRequestCRDYAML)
	if err != nil {
		return err
	}
	err = CreateCustomResourceDefinition(clientset.APIextClient, deviceRequestcrd)
	if err != nil {
		return err
	}

	// creating NDM daemonset
	daemonset, err := GetDaemonSet()
	if err != nil {
		return err
	}
	err = CreateDaemonSet(clientset.RunTimeClient, daemonset)
	if err != nil {
		return err
	}
	return nil
}

// DeleteNDMYAML deletes all the objects specified in the NDM operator YAML.
// ConfigMap, ServiceAccount, ClusterRole, ClusterRoleBinding,
// CustomResourceDefinition and DaemonSet are deleted
func DeleteNDMYAML(clientset k8sClient) error {
	var err error

	// deleting NDM ConfigMap
	configmap, err := GetConfigMap()
	if err != nil {
		return err
	}
	err = DeleteConfigMap(clientset.RunTimeClient, configmap)
	if err != nil {
		return err
	}

	// deleting NDM serviceAccount
	serviceaccount, err := GetServiceAccount()
	if err != nil {
		return err
	}
	err = DeleteServiceAccount(clientset.RunTimeClient, serviceaccount)
	if err != nil {
		return err
	}

	// deleting NDM cluster role
	clusterrole, err := GetClusterRole()
	if err != nil {
		return err
	}
	err = DeleteClusterRole(clientset.RunTimeClient, clusterrole)
	if err != nil {
		return err
	}

	// deleting NDM clusterrole binding
	clusterrolebinding, err := GetClusterRoleBinding()
	if err != nil {
		return err
	}
	err = DeleteClusterRoleBinding(clientset.RunTimeClient, clusterrolebinding)
	if err != nil {
		return err
	}

	// deleting disk crd
	diskcrd, err := GetCustomResourceDefinition(DiskCRDYAML)
	if err != nil {
		return err
	}
	err = DeleteCustomResourceDefinition(clientset.APIextClient, diskcrd)
	if err != nil {
		return err
	}
	// deleting device crd
	devicecrd, err := GetCustomResourceDefinition(DeviceCRDYAML)
	if err != nil {
		return err
	}
	err = DeleteCustomResourceDefinition(clientset.APIextClient, devicecrd)
	if err != nil {
		return err
	}
	// deleting device request crd
	devicerequestcrd, err := GetCustomResourceDefinition(DeviceRequestCRDYAML)
	if err != nil {
		return err
	}
	err = DeleteCustomResourceDefinition(clientset.APIextClient, devicerequestcrd)
	if err != nil {
		return err
	}

	// deleting NDM daemonset
	daemonset, err := GetDaemonSet()
	if err != nil {
		return err
	}
	err = DeleteDaemonSet(clientset.RunTimeClient, daemonset)
	if err != nil {
		return err
	}

	return nil
}

// CreateConfigMap creates the NDM config map
func CreateConfigMap(clientset client.Client, configMap v1.ConfigMap) error {
	err := clientset.Create(context.Background(), &configMap)
	return err
}

// DeleteConfigMap deletes the NDM config map
func DeleteConfigMap(clientset client.Client, configMap v1.ConfigMap) error {
	err := clientset.Delete(context.Background(), &configMap)
	return err
}

// CreateServiceAccount creates the service account required for NDM
func CreateServiceAccount(clientset client.Client, serviceAccount v1.ServiceAccount) error {
	err := clientset.Create(context.Background(), &serviceAccount)
	return err
}

// DeleteServiceAccount deletes the service account used by NDM
func DeleteServiceAccount(clientset client.Client, serviceAccount v1.ServiceAccount) error {
	err := clientset.Delete(context.Background(), &serviceAccount)
	return err
}

// CreateClusterRole creates the required cluster role
func CreateClusterRole(clientset client.Client, clusterRole rbacv1beta1.ClusterRole) error {
	err := clientset.Create(context.Background(), &clusterRole)
	return err
}

// DeleteClusterRole deletes the cluster role used by NDM
func DeleteClusterRole(clientset client.Client, clusterRole rbacv1beta1.ClusterRole) error {
	err := clientset.Delete(context.Background(), &clusterRole)
	return err
}

// CreateClusterRoleBinding creates the rolebinding
func CreateClusterRoleBinding(clientset client.Client, clusterRoleBinding rbacv1beta1.ClusterRoleBinding) error {
	err := clientset.Create(context.Background(), &clusterRoleBinding)
	return err
}

// DeleteClusterRoleBinding deletes the role binding
func DeleteClusterRoleBinding(clientset client.Client, clusterrolebinding rbacv1beta1.ClusterRoleBinding) error {
	err := clientset.Delete(context.Background(), &clusterrolebinding)
	return err
}

// CreateCustomResourceDefinition creates the CRD (currently CRD for disk)
func CreateCustomResourceDefinition(clientset *apiextensionsclient.Clientset, customResourceDefinition apiextensionsv1beta1.CustomResourceDefinition) error {
	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(&customResourceDefinition)
	return err
}

// DeleteCustomResourceDefinition deletes the CRD
func DeleteCustomResourceDefinition(clientset *apiextensionsclient.Clientset, customResourceDefinition apiextensionsv1beta1.CustomResourceDefinition) error {
	err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(customResourceDefinition.Name, &metav1.DeleteOptions{})
	return err
}

// CreateDaemonSet creates the NDM Daemonset
func CreateDaemonSet(clientset client.Client, daemonset v1beta1.DaemonSet) error {
	daemonset.Namespace = namespace
	err := clientset.Create(context.Background(), &daemonset)
	return err
}

// DeleteDaemonSet deletes the NDM Daemonset from the cluster
func DeleteDaemonSet(clientset client.Client, daemonset v1beta1.DaemonSet) error {
	daemonset.Namespace = namespace
	err := clientset.Delete(context.Background(), &daemonset, client.PropagationPolicy(metav1.DeletePropagationForeground))
	return err
}
