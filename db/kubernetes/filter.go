package kubernetes

import "github.com/openebs/node-disk-manager/blockdevice"

const (
	kubernetesLabelPrefix   = "kubernetes.io/"
	KubernetesHostNameLabel = kubernetesLabelPrefix + blockdevice.HostName
)

func GenerateFilter(key, value string) string {
	var filterKey string
	if key == blockdevice.HostName {
		filterKey = KubernetesHostNameLabel
	}

	filterString := filterKey + "=" + value
	return filterString
}
