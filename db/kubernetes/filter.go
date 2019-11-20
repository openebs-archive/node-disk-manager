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

package kubernetes

import "github.com/openebs/node-disk-manager/blockdevice"

const (
	// kubernetesLabelPrefix is the label prefix for kubernetes
	kubernetesLabelPrefix = "kubernetes.io/"

	// KubernetesHostNameLabel is the kubernetes hostname label
	KubernetesHostNameLabel = kubernetesLabelPrefix + blockdevice.HostName
)

// GenerateLabelFilter is used to generate a label filter that can be used
// while listing resources
func GenerateLabelFilter(key, value string) string {
	var filterKey string

	// if key or value is empty, filter will be empty string
	if len(key) == 0 ||
		len(value) == 0 {
		return ""
	}

	// depending on the key, the filter key will be different
	switch key {
	case blockdevice.HostName:
		filterKey = KubernetesHostNameLabel
	default:
		filterKey = key
	}

	filterString := filterKey + "=" + value
	return filterString
}
