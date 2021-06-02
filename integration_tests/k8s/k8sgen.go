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
	"github.com/ghodss/yaml"
	"github.com/openebs/node-disk-manager/integration_tests/utils"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// NDMYaml is a string type that stores the path to the YAML files
// which are required to deploy NDM
type NDMYaml string

// Path to various YAMLs used for integration testing
const (
	ConfigMapYAML           NDMYaml = "../yamls/node-disk-manager-config.yaml"
	ServiceAccountYAML      NDMYaml = "../../deploy/yamls/serviceaccount.yaml"
	ClusterRoleYAML         NDMYaml = "../../deploy/yamls/clusterrole.yaml"
	ClusterRoleBindingYAML  NDMYaml = "../../deploy/yamls/clusterrolebinding.yaml"
	BlockDeviceCRDYAML      NDMYaml = "../../deploy/crds/openebs.io_blockdevices.yaml"
	BlockDeviceClaimCRDYAML NDMYaml = "../../deploy/crds/openebs.io_blockdeviceclaims.yaml"
	DaemonSetYAML           NDMYaml = "../yamls/node-disk-manager.yaml"
	DeploymentYAML          NDMYaml = "../../deploy/yamls/node-disk-operator.yaml"
	OpenEBSNamespaceYAML    NDMYaml = "../../deploy/yamls/namespace.yaml"
)

// GetNamespace generates the openebs namespace object from yaml file
func GetNamespace() (v1.Namespace, error) {
	var ns v1.Namespace
	yamlstring, err := utils.GetYAMLString(string(OpenEBSNamespaceYAML))
	if err != nil {
		return ns, err
	}
	err = yaml.Unmarshal([]byte(yamlstring), &ns)
	return ns, err
}

// GetConfigMap generates the ConfigMap object for NDM from the yaml file
func GetConfigMap() (v1.ConfigMap, error) {
	var configMap v1.ConfigMap
	yamlstring, err := utils.GetYAMLString(string(ConfigMapYAML))
	if err != nil {
		return configMap, err
	}
	err = yaml.Unmarshal([]byte(yamlstring), &configMap)
	return configMap, err
}

// GetServiceAccount generates the ServiceAccount object from the yaml file
func GetServiceAccount() (v1.ServiceAccount, error) {
	var serviceAccount v1.ServiceAccount
	yamlstring, err := utils.GetYAMLString(string(ServiceAccountYAML))
	if err != nil {
		return serviceAccount, err
	}
	err = yaml.Unmarshal([]byte(yamlstring), &serviceAccount)
	return serviceAccount, err
}

// GetClusterRole generates the ClusterRole object from the yaml file
func GetClusterRole() (rbacv1.ClusterRole, error) {
	var clusterRole rbacv1.ClusterRole
	yamlstring, err := utils.GetYAMLString(string(ClusterRoleYAML))
	if err != nil {
		return clusterRole, err
	}
	err = yaml.Unmarshal([]byte(yamlstring), &clusterRole)
	return clusterRole, err
}

// GetClusterRoleBinding generates the ClusterRoleBinding object from the yaml file
func GetClusterRoleBinding() (rbacv1.ClusterRoleBinding, error) {
	var clusterRoleBinding rbacv1.ClusterRoleBinding
	yamlstring, err := utils.GetYAMLString(string(ClusterRoleBindingYAML))
	if err != nil {
		return clusterRoleBinding, err
	}
	err = yaml.Unmarshal([]byte(yamlstring), &clusterRoleBinding)
	return clusterRoleBinding, err
}

// GetCustomResourceDefinition generates the CustomResourceDefinition object from the specified
// YAML file
func GetCustomResourceDefinition(crdyaml NDMYaml) (apiextensionsV1.CustomResourceDefinition, error) {
	var customResourceDefinition apiextensionsV1.CustomResourceDefinition
	yamlString, err := utils.GetYAMLString(string(crdyaml))
	if err != nil {
		return customResourceDefinition, err
	}
	err = yaml.Unmarshal([]byte(yamlString), &customResourceDefinition)
	return customResourceDefinition, err
}

// GetDaemonSet generates the NDM DaemonSet object from the yaml file
func GetDaemonSet() (appsv1.DaemonSet, error) {
	var daemonSet appsv1.DaemonSet
	yamlstring, err := utils.GetYAMLString(string(DaemonSetYAML))
	if err != nil {
		return daemonSet, err
	}
	err = yaml.Unmarshal([]byte(yamlstring), &daemonSet)
	return daemonSet, err
}

// GetDeployment generates the NDO Deployment object from the yaml file
func GetDeployment() (appsv1.Deployment, error) {
	var deployment appsv1.Deployment
	yamlstring, err := utils.GetYAMLString(string(DeploymentYAML))
	if err != nil {
		return deployment, err
	}
	err = yaml.Unmarshal([]byte(yamlstring), &deployment)
	return deployment, err
}
