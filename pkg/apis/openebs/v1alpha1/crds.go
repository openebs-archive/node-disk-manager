package v1alpha1

import (
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Disk Resource
	DiskResourceKind     = "Disk"
	DiskResourceListKind = "DiskList"
	DiskResourcePlural   = "disks"
	DiskResourceShort    = "disk"
	DiskResourceName     = DiskResourcePlural + "." + GroupName

	// BlockDevice Resource
	BlockDeviceResourceKind     = "BlockDevice"
	BlockDeviceResourceListKind = "BlockDeviceList"
	BlockDeviceResourcePlural   = "blockdevices"
	BlockDeviceResourceShort    = "bd"
	BlockDeviceResourceName     = BlockDeviceResourcePlural + "." + GroupName

	// BlockDevice Resource
	BlockDeviceClaimResourceKind     = "BlockDeviceClaim"
	BlockDeviceClaimResourceListKind = "BlockDeviceClaimList"
	BlockDeviceClaimResourcePlural   = "blockdeviceclaims"
	BlockDeviceClaimResourceShort    = "bdc"
	BlockDeviceClaimResourceName     = BlockDeviceClaimResourcePlural + "." + GroupName
)

func buildCRD(name, kind, listKind, plural, short string, scope apiext.ResourceScope) *apiext.CustomResourceDefinition {
	return &apiext.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiext.CustomResourceDefinitionSpec{
			Group:   SchemeGroupVersion.Group,
			Version: SchemeGroupVersion.Version,
			Scope:   scope,
			Names: apiext.CustomResourceDefinitionNames{
				Kind:       kind,
				ListKind:   listKind,
				Plural:     plural,
				ShortNames: []string{short},
			},
		},
	}
}

// DiskCRD returns a cluster-scoped disk CustomResourceDefinition
func DiskCRD() *apiext.CustomResourceDefinition {
	return buildCRD(DiskResourceName,
		DiskResourceKind,
		DiskResourceListKind,
		DiskResourcePlural,
		DiskResourceShort,
		apiext.ClusterScoped)
}

// BlockDeviceCRD returns a namespace-scoped blockdevice CustomResourceDefinition
func BlockDeviceCRD() *apiext.CustomResourceDefinition {
	return buildCRD(BlockDeviceResourceName,
		BlockDeviceResourceKind,
		BlockDeviceResourceListKind,
		BlockDeviceResourcePlural,
		BlockDeviceResourceShort,
		apiext.NamespaceScoped)
}

// BlockDeviceClaimCRD returns a namespace scoped blockdeviceclaim CustomResourceDefinition
func BlockDeviceClaimCRD() *apiext.CustomResourceDefinition {
	return buildCRD(BlockDeviceClaimResourceName,
		BlockDeviceClaimResourceKind,
		BlockDeviceClaimResourceListKind,
		BlockDeviceClaimResourcePlural,
		BlockDeviceClaimResourceShort,
		apiext.NamespaceScoped)
}
