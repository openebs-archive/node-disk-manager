package k8s

import (
	"github.com/ghodss/yaml"
	"github.com/openebs/node-disk-manager/integration_tests/utils"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// NDMYaml is a string type that stores the path to the YAML files
// which are required to deploy NDM
type NDMYaml string

// Path to various YAMLs used for integration testing
const (
	ConfigMapYAML           NDMYaml = "../yamls/configmap.yaml"
	ServiceAccountYAML      NDMYaml = "../yamls/serviceaccount.yaml"
	ClusterRoleYAML         NDMYaml = "../yamls/clusterrole.yaml"
	ClusterRoleBindingYAML  NDMYaml = "../yamls/clusterrolebinding.yaml"
	DiskCRDYAML             NDMYaml = "../yamls/diskCR.yaml"
	BlockDeviceCRDYAML      NDMYaml = "../yamls/blockDeviceCR.yaml"
	BlockDeviceClaimCRDYAML NDMYaml = "../yamls/blockDeviceClaimCR.yaml"
	DaemonSetYAML           NDMYaml = "../yamls/daemonset.yaml"
	DeploymentYAML          NDMYaml = "../yamls/deployment.yaml"
)

// GetConfigMap generates the ConfigMap object for NDM from the yaml file
func GetConfigMap() (v1.ConfigMap, error) {
	var configMap v1.ConfigMap
	yamlstring, err := utils.GetYAMLString(string(ConfigMapYAML))
	if err != nil {
		return configMap, err
	}
	err = yaml.Unmarshal([]byte(yamlstring), &configMap)
	if err != nil {
		return configMap, err
	}
	return configMap, nil
}

// GetServiceAccount generates the ServiceAccount object from the yaml file
func GetServiceAccount() (v1.ServiceAccount, error) {
	var serviceAccount v1.ServiceAccount
	yamlstring, err := utils.GetYAMLString(string(ServiceAccountYAML))
	if err != nil {
		return serviceAccount, err
	}
	err = yaml.Unmarshal([]byte(yamlstring), &serviceAccount)
	if err != nil {
		return serviceAccount, err
	}
	return serviceAccount, nil
}

// GetClusterRole generates the ClusterRole object from the yaml file
func GetClusterRole() (rbacv1beta1.ClusterRole, error) {
	var clusterRole rbacv1beta1.ClusterRole
	yamlstring, err := utils.GetYAMLString(string(ClusterRoleYAML))
	if err != nil {
		return clusterRole, err
	}
	err = yaml.Unmarshal([]byte(yamlstring), &clusterRole)
	if err != nil {
		return clusterRole, err
	}
	return clusterRole, nil
}

// GetClusterRoleBinding generates the ClusterRoleBinding object from the yaml file
func GetClusterRoleBinding() (rbacv1beta1.ClusterRoleBinding, error) {
	var clusterRoleBinding rbacv1beta1.ClusterRoleBinding
	yamlstring, err := utils.GetYAMLString(string(ClusterRoleBindingYAML))
	if err != nil {
		return clusterRoleBinding, err
	}
	err = yaml.Unmarshal([]byte(yamlstring), &clusterRoleBinding)
	if err != nil {
		return clusterRoleBinding, err
	}
	return clusterRoleBinding, err
}

// GetCustomResourceDefinition generates the CustomResourceDefinition object from the specified
// YAML file
func GetCustomResourceDefinition(crdyaml NDMYaml) (apiextensionsv1beta1.CustomResourceDefinition, error) {
	var customResourceDefinition apiextensionsv1beta1.CustomResourceDefinition
	yamlString, err := utils.GetYAMLString(string(crdyaml))
	if err != nil {
		return customResourceDefinition, err
	}
	err = yaml.Unmarshal([]byte(yamlString), &customResourceDefinition)
	if err != nil {
		return customResourceDefinition, err
	}
	return customResourceDefinition, err
}

// GetDaemonSet generates the NDM DaemonSet object from the yaml file
func GetDaemonSet() (v1beta1.DaemonSet, error) {
	var daemonSet v1beta1.DaemonSet
	yamlstring, err := utils.GetYAMLString(string(DaemonSetYAML))
	if err != nil {
		return daemonSet, err
	}
	err = yaml.Unmarshal([]byte(yamlstring), &daemonSet)
	if err != nil {
		return daemonSet, err
	}
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
	if err != nil {
		return deployment, err
	}
	return deployment, err
}
