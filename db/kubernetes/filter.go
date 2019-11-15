package kubernetes

import "github.com/openebs/node-disk-manager/blockdevice"

const (
	kubernetesLabelPrefix   = "kubernetes.io/"
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
