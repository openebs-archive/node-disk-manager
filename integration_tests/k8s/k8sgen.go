package k8s

import (
	"github.com/ghodss/yaml"
	"github.com/openebs/node-disk-manager/integration_tests/utils"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

type NDMYaml string

const (
	ConfigMapYAML                NDMYaml = "../yamls/configmap.yaml"
	ServiceAccountYAML           NDMYaml = "../yamls/serviceaccount.yaml"
	ClusterRoleYAML              NDMYaml = "../yamls/clusterrole.yaml"
	ClusterRoleBindingYAML       NDMYaml = "../yamls/clusterrolebinding.yaml"
	CustomResourceDefinitionYAML NDMYaml = "../yamls/diskCR.yaml"
	DaemonSetYAML                NDMYaml = "../yamls/daemonset.yaml"
)

// Generate the ConfigMap object for NDM from the yaml file
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

// Generate the ServiceAccount object from the yaml file
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

// Generate the ClusterRole object from the yaml file
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

// Generate the ClusterRoleBinding object from the yaml file
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

// Generate the CustomResourceDefinition object for disk CR from the yaml file
func GetCustomResourceDefinition() (apiextensionsv1beta1.CustomResourceDefinition, error) {
	var customResourceDefinition apiextensionsv1beta1.CustomResourceDefinition
	yamlString, err := utils.GetYAMLString(string(CustomResourceDefinitionYAML))
	if err != nil {
		return customResourceDefinition, err
	}
	err = yaml.Unmarshal([]byte(yamlString), &customResourceDefinition)
	if err != nil {
		return customResourceDefinition, err
	}
	return customResourceDefinition, err
}

// Generate the NDM DaemonSet object from the yaml file
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
