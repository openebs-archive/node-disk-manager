package k8s

import (
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

// The wait time for all k8s API related operations
const k8sWaitTime = time.Minute

// Get all the pods in the given namespace along with their status
func GetPods(clientSet *kubernetes.Clientset, ns string) (map[string]string, error) {
	pods := make(map[string]string)
	podList, err := clientSet.CoreV1().Pods(ns).List(metav1.ListOptions{})
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
	nodeList, err := clientSet.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, node := range nodeList.Items {
		nodes[node.Name] = string(node.Status.Phase)
	}
	return nodes, nil
}

// Create all the objects specified in the NDM operator YAML
// ConfigMap, ServiceAccount, ClusterRole, ClusterRoleBinding,
// CustomResourceDefinition and DaemonSet are created
func CreateNDMYAML(clientset *kubernetes.Clientset, ns string) error {
	var err error

	// creating NDM ConfigMap
	configmap, err := GetConfigMap()
	if err != nil {
		return err
	}
	err = CreateConfigMap(clientset, ns, configmap)
	if err != nil {
		return err
	}

	// creating NDM serviceAccount
	serviceaccount, err := GetServiceAccount()
	if err != nil {
		return err
	}
	err = CreateServiceAccount(clientset, ns, serviceaccount)
	if err != nil {
		return err
	}

	// creating NDM cluster role
	clusterrole, err := GetClusterRole()
	if err != nil {
		return err
	}
	err = CreateClusterRole(clientset, ns, clusterrole)
	if err != nil {
		return err
	}

	// creating NDM clusterrole binding
	clusterrolebinding, err := GetClusterRoleBinding()
	if err != nil {
		return err
	}
	err = CreateClusterRoleBinding(clientset, ns, clusterrolebinding)
	if err != nil {
		return err
	}

	// creating NDM daemonset
	daemonset, err := GetDaemonSet()
	if err != nil {
		return err
	}
	err = CreateDaemonSet(clientset, ns, daemonset)
	if err != nil {
		return err
	}
	return nil
}

// Deletes all the objects specified in the NDM operator YAML
// ConfigMap, ServiceAccount, ClusterRole, ClusterRoleBinding,
// CustomResourceDefinition and DaemonSet are deleted
func DeleteNDMYAML(clientset *kubernetes.Clientset, ns string) error {
	var err error

	// deleting NDM ConfigMap
	configmap, err := GetConfigMap()
	if err != nil {
		return err
	}
	err = DeleteConfigMap(clientset, ns, configmap)
	if err != nil {
		return err
	}

	// deleting NDM serviceAccount
	serviceaccount, err := GetServiceAccount()
	if err != nil {
		return err
	}
	err = DeleteServiceAccount(clientset, ns, serviceaccount)
	if err != nil {
		return err
	}

	// deleting NDM cluster role
	clusterrole, err := GetClusterRole()
	if err != nil {
		return err
	}
	err = DeleteClusterRole(clientset, ns, clusterrole)
	if err != nil {
		return err
	}

	// deleting NDM clusterrole binding
	clusterrolebinding, err := GetClusterRoleBinding()
	if err != nil {
		return err
	}
	err = DeleteClusterRoleBinding(clientset, ns, clusterrolebinding)
	if err != nil {
		return err
	}

	// deleting NDM daemonset
	daemonset, err := GetDaemonSet()
	if err != nil {
		return err
	}
	err = DeleteDaemonSet(clientset, ns, daemonset)
	if err != nil {
		return err
	}

	return nil
}

// Create the ConfigMap for NDM
func CreateConfigMap(clientset *kubernetes.Clientset, ns string, configMap v1.ConfigMap) error {
	_, err := clientset.CoreV1().ConfigMaps(ns).Create(&configMap)
	return err
}

// Delete the ConfigMap for NDM
func DeleteConfigMap(clientset *kubernetes.Clientset, ns string, configMap v1.ConfigMap) error {
	err := clientset.CoreV1().ConfigMaps(ns).Delete(configMap.Name, &metav1.DeleteOptions{})
	return err
}

// Create ServiceAccount required for NDM
func CreateServiceAccount(clientset *kubernetes.Clientset, ns string, serviceAccount v1.ServiceAccount) error {
	_, err := clientset.CoreV1().ServiceAccounts(ns).Create(&serviceAccount)
	return err
}

// Delete ServiceAccount required for NDM
func DeleteServiceAccount(clientset *kubernetes.Clientset, ns string, serviceAccount v1.ServiceAccount) error {
	err := clientset.CoreV1().ServiceAccounts(ns).Delete(serviceAccount.Name, &metav1.DeleteOptions{})
	return err
}

// Create the ClusterRole required for NDM
func CreateClusterRole(clientset *kubernetes.Clientset, ns string, clusterrole rbacv1beta1.ClusterRole) error {
	_, err := clientset.RbacV1beta1().ClusterRoles().Create(&clusterrole)
	return err
}

// Delete the ClusterRole required for NDM
func DeleteClusterRole(clientset *kubernetes.Clientset, ns string, clusterrole rbacv1beta1.ClusterRole) error {
	err := clientset.RbacV1beta1().ClusterRoles().Delete(clusterrole.Name, &metav1.DeleteOptions{})
	return err
}

// Create the ClusterRoleBinding required for NDM
func CreateClusterRoleBinding(clientset *kubernetes.Clientset, ns string, clusterrolebinding rbacv1beta1.ClusterRoleBinding) error {
	_, err := clientset.RbacV1beta1().ClusterRoleBindings().Create(&clusterrolebinding)
	return err
}

// Delete the ClusterRoleBinding required for NDM
func DeleteClusterRoleBinding(clientset *kubernetes.Clientset, ns string, clusterrolebinding rbacv1beta1.ClusterRoleBinding) error {
	err := clientset.RbacV1beta1().ClusterRoleBindings().Delete(clusterrolebinding.Name, &metav1.DeleteOptions{})
	return err
}

/*// TODO Need to find package of CRD
func CreateCustomResourceDescription(clientset *kubernetes.Clientset, ns string, customresourcedefinition apiextensionsv1beta1.CustomResourceDefinition) error {
	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(&customresourcedefinition)
	return err
}*/

// Create the NDM DaemonSet
func CreateDaemonSet(clientset *kubernetes.Clientset, ns string, daemonset v1beta1.DaemonSet) error {
	_, err := clientset.ExtensionsV1beta1().DaemonSets(ns).Create(&daemonset)
	return err
}

// Delete NDM DaemonSet
func DeleteDaemonSet(clientset *kubernetes.Clientset, ns string, daemonset v1beta1.DaemonSet) error {
	deletePropogation := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePropogation,
	}
	err := clientset.ExtensionsV1beta1().DaemonSets(ns).Delete(daemonset.Name, &deleteOptions)
	return err
}
