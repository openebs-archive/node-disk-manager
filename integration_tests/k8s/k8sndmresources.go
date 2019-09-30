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

/*// CreateNDMYAML creates all the objects specified in the NDM operator YAML.
// Each resource object is generated from yaml file in ../yamls/ and parsed into
// the required type. ConfigMap, ServiceAccount, ClusterRole, ClusterRoleBinding,
// CustomResourceDefinition and DaemonSet are created
func (c K8sClient) CreateNDMYAML() error {
	var err error

	// creating NDM ConfigMap
	err = c.CreateNDMConfigMap()
	if err != nil {
		return err
	}

	// creating NDM serviceAccount
	err = c.CreateNDMServiceAccount()
	if err != nil {
		return err
	}

	// creating NDM cluster role
	err = c.CreateNDMClusterRole()
	if err != nil {
		return err
	}

	// creating NDM clusterrole binding
	err = c.CreateNDMClusterRoleBinding()
	if err != nil {
		return err
	}

	// creating ndm daemon-set
	err = c.CreateNDMDaemonSet()
	if err != nil {
		return err
	}

	// creating NDO Deployment
	return c.CreateNDMOperatorDeployment()
}

// DeleteNDMYAML deletes all the objects specified in the NDM operator YAML.
// ConfigMap, ServiceAccount, ClusterRole, ClusterRoleBinding,
// CustomResourceDefinition and DaemonSet are deleted
func (c K8sClient) DeleteNDMYAML() error {
	var err error

	// deleting NDM ConfigMap
	err = c.DeleteNDMConfigMap()
	if err != nil {
		return err
	}

	// deleting NDM serviceAccount
	err = c.DeleteNDMServiceAccount()
	if err != nil {
		return err
	}

	// deleting NDM cluster role
	err = c.DeleteNDMClusterRole()
	if err != nil {
		return err
	}

	// deleting NDM clusterrole binding
	err = c.DeleteNDMClusterRoleBinding()
	if err != nil {
		return err
	}

	// deleting ndm daemon-set
	err = c.DeleteNDMDaemonSet()
	if err != nil {
		return err
	}

	// deleting NDO Deployment
	return c.DeleteNDMOperatorDeployment()
}
*/
// CreateNDMConfigMap creates the ConfigMap required for NDM
func (c K8sClient) CreateNDMConfigMap() error {
	configmap, err := GetConfigMap()
	if err != nil {
		return err
	}
	return c.CreateConfigMap(configmap)
}

// CreateNDMServiceAccount creates the ServiceAccount required for NDM
func (c K8sClient) CreateNDMServiceAccount() error {
	serviceaccount, err := GetServiceAccount()
	if err != nil {
		return err
	}
	return c.CreateServiceAccount(serviceaccount)
}

// CreateNDMClusterRole creates the ClusterRole required for NDM
func (c K8sClient) CreateNDMClusterRole() error {
	clusterrole, err := GetClusterRole()
	if err != nil {
		return err
	}
	return c.CreateClusterRole(clusterrole)
}

// CreateNDMClusterRoleBinding creates the role binding required by NDM
func (c K8sClient) CreateNDMClusterRoleBinding() error {
	clusterrolebinding, err := GetClusterRoleBinding()
	if err != nil {
		return err
	}
	return c.CreateClusterRoleBinding(clusterrolebinding)
}

// CreateNDMCRDs creates the Disk, BlockDevice and BlockDeviceClaim CRDs
func (c K8sClient) CreateNDMCRDs() error {
	var err error
	err = c.CreateNDMDiskCRD()
	if err != nil {
		return err
	}
	err = c.CreateNDMBlockDeviceCRD()
	if err != nil {
		return err
	}
	err = c.CreateNDMBlockDeviceClaimCRD()
	if err != nil {
		return err
	}
	return nil
}

// CreateNDMDiskCRD creates the CustomResourceDefinition for a Disk type
func (c K8sClient) CreateNDMDiskCRD() error {
	diskcrd, err := GetCustomResourceDefinition(DiskCRDYAML)
	if err != nil {
		return err
	}
	return c.CreateCustomResourceDefinition(diskcrd)
}

// CreateNDMBlockDeviceCRD creates the CustomResourceDefinition for a Device type
func (c K8sClient) CreateNDMBlockDeviceCRD() error {
	devicecrd, err := GetCustomResourceDefinition(BlockDeviceCRDYAML)
	if err != nil {
		return err
	}
	return c.CreateCustomResourceDefinition(devicecrd)
}

// CreateNDMBlockDeviceClaimCRD creates the CustomResourceDefinition for a BlockDeviceClaim type
func (c K8sClient) CreateNDMBlockDeviceClaimCRD() error {
	deviceClaimcrd, err := GetCustomResourceDefinition(BlockDeviceClaimCRDYAML)
	if err != nil {
		return err
	}
	return c.CreateCustomResourceDefinition(deviceClaimcrd)
}

// CreateNDMDaemonSet creates the NDM Daemonset
func (c K8sClient) CreateNDMDaemonSet() error {
	daemonset, err := GetDaemonSet()
	if err != nil {
		return err
	}
	return c.CreateDaemonSet(daemonset)
}

// DeleteNDMConfigMap deletes the ConfigMap required for NDM
func (c K8sClient) DeleteNDMConfigMap() error {
	configmap, err := GetConfigMap()
	if err != nil {
		return err
	}
	return c.DeleteConfigMap(configmap)
}

// DeleteNDMServiceAccount deletes the ServiceAccount required for NDM
func (c K8sClient) DeleteNDMServiceAccount() error {
	serviceaccount, err := GetServiceAccount()
	if err != nil {
		return err
	}
	return c.DeleteServiceAccount(serviceaccount)
}

// DeleteNDMClusterRole deletes the ClusterRole required for NDM
func (c K8sClient) DeleteNDMClusterRole() error {
	clusterrole, err := GetClusterRole()
	if err != nil {
		return err
	}
	return c.DeleteClusterRole(clusterrole)
}

// DeleteNDMClusterRoleBinding deletes the role binding required by NDM
func (c K8sClient) DeleteNDMClusterRoleBinding() error {
	clusterrolebinding, err := GetClusterRoleBinding()
	if err != nil {
		return err
	}
	return c.DeleteClusterRoleBinding(clusterrolebinding)
}

// DeleteNDMCRDs deletes the disk, blockdevice and blockdevice claim CRDs
func (c K8sClient) DeleteNDMCRDs() error {
	var err error
	err = c.DeleteNDMDiskCRD()
	if err != nil {
		return err
	}
	err = c.DeleteNDMBlockDeviceCRD()
	if err != nil {
		return err
	}
	err = c.DeleteNDMBlockDeviceClaimCRD()
	if err != nil {
		return err
	}
	return nil
}

// DeleteNDMDiskCRD deletes the CustomResourceDefinition for a Disk type
func (c K8sClient) DeleteNDMDiskCRD() error {
	diskcrd, err := GetCustomResourceDefinition(DiskCRDYAML)
	if err != nil {
		return err
	}
	return c.DeleteCustomResourceDefinition(diskcrd)
}

// DeleteNDMBlockDeviceCRD deletes the CustomResourceDefinition for a Device type
func (c K8sClient) DeleteNDMBlockDeviceCRD() error {
	devicecrd, err := GetCustomResourceDefinition(BlockDeviceCRDYAML)
	if err != nil {
		return err
	}
	return c.DeleteCustomResourceDefinition(devicecrd)
}

// DeleteNDMBlockDeviceClaimCRD deletes the CustomResourceDefinition for a BlockDeviceClaim type
func (c K8sClient) DeleteNDMBlockDeviceClaimCRD() error {
	deviceClaimcrd, err := GetCustomResourceDefinition(BlockDeviceClaimCRDYAML)
	if err != nil {
		return err
	}
	return c.DeleteCustomResourceDefinition(deviceClaimcrd)
}

// DeleteNDMDaemonSet deletes the NDM Daemonset
func (c K8sClient) DeleteNDMDaemonSet() error {
	daemonset, err := GetDaemonSet()
	if err != nil {
		return err
	}
	return c.DeleteDaemonSet(daemonset)
}

// CreateNDMOperatorDeployment creates the NDM Operator
func (c K8sClient) CreateNDMOperatorDeployment() error {
	deployment, err := GetDeployment()
	if err != nil {
		return err
	}
	return c.CreateDeployment(deployment)
}

// DeleteNDMOperatorDeployment deletes the NDM operator
func (c K8sClient) DeleteNDMOperatorDeployment() error {
	deployment, err := GetDeployment()
	if err != nil {
		return err
	}
	return c.DeleteDeployment(deployment)
}
