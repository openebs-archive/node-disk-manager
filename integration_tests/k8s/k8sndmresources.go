package k8s

import (
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateNDMYAML creates all the objects specified in the NDM operator YAML.
// Each resource object is generated from yaml file in ../yamls/ and parsed into
// the required type. ConfigMap, ServiceAccount, ClusterRole, ClusterRoleBinding,
// CustomResourceDefinition and DaemonSet are created
func CreateNDMYAML(clientset k8sClient) error {
	var err error

	// creating NDM ConfigMap
	err = CreateNDMConfigMap(clientset.RunTimeClient)
	if err != nil {
		return err
	}

	// creating NDM serviceAccount
	err = CreateNDMServiceAccount(clientset.RunTimeClient)
	if err != nil {
		return err
	}

	// creating NDM cluster role
	err = CreateNDMClusterRole(clientset.RunTimeClient)
	if err != nil {
		return err
	}

	// creating NDM clusterrole binding
	err = CreateNDMClusterRoleBinding(clientset.RunTimeClient)
	if err != nil {
		return err
	}

	// creating disk crd
	err = CreateNDMDiskCRD(clientset.APIextClient)
	if err != nil {
		return err
	}

	// creating device crd
	err = CreateNDMDeviceCRD(clientset.APIextClient)
	if err != nil {
		return err
	}

	// creating device request crd
	err = CreateNDMDeviceRequestCRD(clientset.APIextClient)
	if err != nil {
		return err
	}

	// creating NDM daemonset
	return CreateNDMDaemonSet(clientset.RunTimeClient)
}

// DeleteNDMYAML deletes all the objects specified in the NDM operator YAML.
// ConfigMap, ServiceAccount, ClusterRole, ClusterRoleBinding,
// CustomResourceDefinition and DaemonSet are deleted
func DeleteNDMYAML(clientset k8sClient) error {
	var err error

	// deleting NDM ConfigMap
	err = DeleteNDMConfigMap(clientset.RunTimeClient)
	if err != nil {
		return err
	}

	// deleting NDM serviceAccount
	err = DeleteNDMServiceAccount(clientset.RunTimeClient)
	if err != nil {
		return err
	}

	// deleting NDM cluster role
	err = DeleteNDMClusterRole(clientset.RunTimeClient)
	if err != nil {
		return err
	}

	// deleting NDM clusterrole binding
	err = DeleteNDMClusterRoleBinding(clientset.RunTimeClient)
	if err != nil {
		return err
	}

	// deleting disk crd
	err = DeleteNDMDiskCRD(clientset.APIextClient)
	if err != nil {
		return err
	}
	// deleting device crd
	err = DeleteNDMDeviceCRD(clientset.APIextClient)
	if err != nil {
		return err
	}
	// deleting device request crd
	err = DeleteNDMDeviceRequestCRD(clientset.APIextClient)
	if err != nil {
		return err
	}

	// deleting NDM daemonset
	return DeleteNDMDaemonSet(clientset.RunTimeClient)
}

// CreateNDMConfigMap creates the ConfigMap required for NDM
func CreateNDMConfigMap(clientSet client.Client) error {
	configmap, err := GetConfigMap()
	if err != nil {
		return err
	}
	return CreateConfigMap(clientSet, configmap)
}

// CreateNDMServiceAccount creates the ServiceAccount required for NDM
func CreateNDMServiceAccount(clientSet client.Client) error {
	serviceaccount, err := GetServiceAccount()
	if err != nil {
		return err
	}
	return CreateServiceAccount(clientSet, serviceaccount)
}

// CreateNDMClusterRole creates the ClusterRole required for NDM
func CreateNDMClusterRole(clientSet client.Client) error {
	clusterrole, err := GetClusterRole()
	if err != nil {
		return err
	}
	return CreateClusterRole(clientSet, clusterrole)
}

// CreateNDMClusterRoleBinding creates the role binding required by NDM
func CreateNDMClusterRoleBinding(clientSet client.Client) error {
	clusterrolebinding, err := GetClusterRoleBinding()
	if err != nil {
		return err
	}
	return CreateClusterRoleBinding(clientSet, clusterrolebinding)
}

// CreateNDMDiskCRD creates the CustomResourceDefinition for a Disk type
func CreateNDMDiskCRD(clientset *apiextensionsclient.Clientset) error {
	diskcrd, err := GetCustomResourceDefinition(DiskCRDYAML)
	if err != nil {
		return err
	}
	return CreateCustomResourceDefinition(clientset, diskcrd)
}

// CreateNDMDeviceCRD creates the CustomResourceDefinition for a Device type
func CreateNDMDeviceCRD(clientset *apiextensionsclient.Clientset) error {
	devicecrd, err := GetCustomResourceDefinition(DeviceCRDYAML)
	if err != nil {
		return err
	}
	return CreateCustomResourceDefinition(clientset, devicecrd)
}

// CreateNDMDeviceRequestCRD creates the CustomResourceDefinition for a DeviceRequest type
func CreateNDMDeviceRequestCRD(clientset *apiextensionsclient.Clientset) error {
	deviceRequestcrd, err := GetCustomResourceDefinition(DeviceRequestCRDYAML)
	if err != nil {
		return err
	}
	return CreateCustomResourceDefinition(clientset, deviceRequestcrd)
}

// CreateNDMDaemonSet creates the NDM Daemonset
func CreateNDMDaemonSet(clientset client.Client) error {
	daemonset, err := GetDaemonSet()
	if err != nil {
		return err
	}
	return CreateDaemonSet(clientset, daemonset)
}

// DeleteNDMConfigMap deletes the ConfigMap required for NDM
func DeleteNDMConfigMap(clientset client.Client) error {
	configmap, err := GetConfigMap()
	if err != nil {
		return err
	}
	return DeleteConfigMap(clientset, configmap)
}

// DeleteNDMServiceAccount deletes the ServiceAccount required for NDM
func DeleteNDMServiceAccount(clientset client.Client) error {
	serviceaccount, err := GetServiceAccount()
	if err != nil {
		return err
	}
	return DeleteServiceAccount(clientset, serviceaccount)
}

// DeleteNDMClusterRole deletes the ClusterRole required for NDM
func DeleteNDMClusterRole(clientset client.Client) error {
	clusterrole, err := GetClusterRole()
	if err != nil {
		return err
	}
	return DeleteClusterRole(clientset, clusterrole)
}

// DeleteNDMClusterRoleBinding deletes the role binding required by NDM
func DeleteNDMClusterRoleBinding(clientset client.Client) error {
	clusterrolebinding, err := GetClusterRoleBinding()
	if err != nil {
		return err
	}
	return DeleteClusterRoleBinding(clientset, clusterrolebinding)
}

// DeleteNDMDiskCRD deletes the CustomResourceDefinition for a Disk type
func DeleteNDMDiskCRD(clientset *apiextensionsclient.Clientset) error {
	diskcrd, err := GetCustomResourceDefinition(DiskCRDYAML)
	if err != nil {
		return err
	}
	return DeleteCustomResourceDefinition(clientset, diskcrd)
}

// DeleteNDMDeviceCRD deletes the CustomResourceDefinition for a Device type
func DeleteNDMDeviceCRD(clientset *apiextensionsclient.Clientset) error {
	devicecrd, err := GetCustomResourceDefinition(DeviceCRDYAML)
	if err != nil {
		return err
	}
	return DeleteCustomResourceDefinition(clientset, devicecrd)
}

// DeleteNDMDeviceRequestCRD deletes the CustomResourceDefinition for a DeviceRequest type
func DeleteNDMDeviceRequestCRD(clientset *apiextensionsclient.Clientset) error {
	devicerequestcrd, err := GetCustomResourceDefinition(DeviceRequestCRDYAML)
	if err != nil {
		return err
	}
	return DeleteCustomResourceDefinition(clientset, devicerequestcrd)
}

// DeleteNDMDaemonSet deletes the NDM Daemonset
func DeleteNDMDaemonSet(clientset client.Client) error {
	daemonset, err := GetDaemonSet()
	if err != nil {
		return err
	}
	return DeleteDaemonSet(clientset, daemonset)
}
