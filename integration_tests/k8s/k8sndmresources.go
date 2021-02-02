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

// CreateOpenEBSNamespace creates the openebs namespace required for NDM installation
func (c K8sClient) CreateOpenEBSNamespace() error {
	ns, err := GetNamespace()
	if err != nil {
		return err
	}
	return c.CreateNamespace(ns)
}

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

// DeleteOpenEBSNamespace deletes the openebs namespace in which NDM was installed
func (c K8sClient) DeleteOpenEBSNamespace() error {
	ns, err := GetNamespace()
	if err != nil {
		return err
	}
	return c.DeleteNamespace(ns)
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
