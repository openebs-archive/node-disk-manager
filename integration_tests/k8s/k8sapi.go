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

// Get all the pods in the given namespace along with their status
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

// Get all the nodes in the cluster along with their status
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

// Get the list of DiskCR in the cluster
func GetDiskList(clientSet k8sClient) (*v1alpha1.DiskList, error) {
	diskList := &v1alpha1.DiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Disk",
			APIVersion: "openebs.io/v1alpha1",
		},
	}
	i := 3

	var err error
	for i > 0 {
		i--
		err = clientSet.RunTimeClient.List(context.TODO(), &client.ListOptions{}, diskList)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot list disks. Error :%v", err)
	}
	return diskList, nil
}

// Create all the objects specified in the NDM operator YAML
// ConfigMap, ServiceAccount, ClusterRole, ClusterRoleBinding,
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

	// creating crd
	crd, err := GetCustomResourceDefinition()
	if err != nil {
		return err
	}
	err = CreateCustomResourceDefinition(clientset.APIextClient, crd)
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

// Deletes all the objects specified in the NDM operator YAML
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

	// deleting crd
	crd, err := GetCustomResourceDefinition()
	if err != nil {
		return err
	}
	err = DeleteCustomResourceDefinition(clientset.APIextClient, crd)
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

// Create the ConfigMap for NDM
func CreateConfigMap(clientset client.Client, configMap v1.ConfigMap) error {
	//_, err := clientset.CoreV1().ConfigMaps(ns).Create(&configMap)
	err := clientset.Create(context.Background(), &configMap)
	return err
}

// Delete the ConfigMap for NDM
func DeleteConfigMap(clientset client.Client, configMap v1.ConfigMap) error {
	err := clientset.Delete(context.Background(), &configMap)
	return err
}

// Create ServiceAccount required for NDM
func CreateServiceAccount(clientset client.Client, serviceAccount v1.ServiceAccount) error {
	//_, err := clientset.CoreV1().ServiceAccounts(ns).Create(&serviceAccount)
	err := clientset.Create(context.Background(), &serviceAccount)
	return err
}

// Delete ServiceAccount required for NDM
func DeleteServiceAccount(clientset client.Client, serviceAccount v1.ServiceAccount) error {
	err := clientset.Delete(context.Background(), &serviceAccount)
	return err
}

// Create the ClusterRole required for NDM
func CreateClusterRole(clientset client.Client, clusterRole rbacv1beta1.ClusterRole) error {
	//_, err := clientset.RbacV1beta1().ClusterRoles().Create(&clusterrole)
	err := clientset.Create(context.Background(), &clusterRole)
	return err
}

// Delete the ClusterRole required for NDM
func DeleteClusterRole(clientset client.Client, clusterRole rbacv1beta1.ClusterRole) error {
	err := clientset.Delete(context.Background(), &clusterRole)
	return err
}

// Create the ClusterRoleBinding required for NDM
func CreateClusterRoleBinding(clientset client.Client, clusterRoleBinding rbacv1beta1.ClusterRoleBinding) error {
	//_, err := clientset.RbacV1beta1().ClusterRoleBindings().Create(&clusterrolebinding)
	err := clientset.Create(context.Background(), &clusterRoleBinding)
	return err
}

// Delete the ClusterRoleBinding required for NDM
func DeleteClusterRoleBinding(clientset client.Client, clusterrolebinding rbacv1beta1.ClusterRoleBinding) error {
	err := clientset.Delete(context.Background(), &clusterrolebinding)
	return err
}

// Create CRD
func CreateCustomResourceDefinition(clientset *apiextensionsclient.Clientset, customResourceDefinition apiextensionsv1beta1.CustomResourceDefinition) error {
	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(&customResourceDefinition)
	//err := clientset.Create(context.Background(), &customresourcedefinition)
	return err
}

// Delete CRD
func DeleteCustomResourceDefinition(clientset *apiextensionsclient.Clientset, customResourceDefinition apiextensionsv1beta1.CustomResourceDefinition) error {
	err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(customResourceDefinition.Name, &metav1.DeleteOptions{})
	//err := clientset.Delete(context.Background(), &customresourcedefinition)
	return err
}

// Create the NDM DaemonSet
func CreateDaemonSet(clientset client.Client, daemonset v1beta1.DaemonSet) error {
	//_, err := clientset.ExtensionsV1beta1().DaemonSets(ns).Create(&daemonset)
	daemonset.Namespace = namespace
	err := clientset.Create(context.Background(), &daemonset)
	return err
}

// Delete NDM DaemonSet
func DeleteDaemonSet(clientset client.Client, daemonset v1beta1.DaemonSet) error {
	/*deletePropogation := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePropogation,
	}*/
	daemonset.Namespace = namespace
	err := clientset.Delete(context.Background(), &daemonset, client.PropagationPolicy(metav1.DeletePropagationForeground))
	return err
}
