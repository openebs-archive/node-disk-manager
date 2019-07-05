package setup

import (
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/crds"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

func buildDiskCRD() (*apiext.CustomResourceDefinition, error) {
	crdBuilder := crds.NewBuilder()
	crdBuilder.WithName(apis.DiskResourceName).
		WithGroup(apis.GroupName).
		WithVersion(apis.APIVersion).
		WithScope(apiext.ClusterScoped).
		WithKind(apis.DiskResourceKind).
		WithListKind(apis.DiskResourceListKind).
		WithPlural(apis.DiskResourcePlural).
		WithShortNames([]string{apis.DiskResourceShort}).
		WithPrinterColumns("Size", "string", ".spec.capacity.storage").
		WithPrinterColumns("State", "string", ".status.state").
		WithPrinterColumns("Age", "date", ".metadata.creationTimestamp")
	return crdBuilder.Build()
}

func buildBlockDeviceCRD() (*apiext.CustomResourceDefinition, error) {
	crdBuilder := crds.NewBuilder()
	crdBuilder.WithName(apis.BlockDeviceResourceName).
		WithGroup(apis.GroupName).
		WithVersion(apis.APIVersion).
		WithScope(apiext.NamespaceScoped).
		WithKind(apis.BlockDeviceResourceKind).
		WithListKind(apis.BlockDeviceResourceListKind).
		WithPlural(apis.BlockDeviceResourcePlural).
		WithShortNames([]string{apis.BlockDeviceResourceShort}).
		WithPrinterColumns("Size", "string", ".spec.capacity.storage").
		WithPrinterColumns("ClaimState", "string", ".status.claimState").
		WithPrinterColumns("Status", "string", ".status.state").
		WithPrinterColumns("Age", "date", ".metadata.creationTimestamp")
	return crdBuilder.Build()
}

func buildBlockDeviceClaimCRD() (*apiext.CustomResourceDefinition, error) {
	crdBuilder := crds.NewBuilder()
	crdBuilder.WithName(apis.BlockDeviceClaimResourceName).
		WithGroup(apis.GroupName).
		WithVersion(apis.APIVersion).
		WithScope(apiext.NamespaceScoped).
		WithKind(apis.BlockDeviceClaimResourceKind).
		WithListKind(apis.BlockDeviceClaimResourceListKind).
		WithPlural(apis.BlockDeviceClaimResourcePlural).
		WithShortNames([]string{apis.BlockDeviceClaimResourceShort}).
		WithPrinterColumns("BlockDeviceName", "string", ".spec.blockDeviceName").
		WithPrinterColumns("Phase", "string", ".status.phase").
		WithPrinterColumns("Age", "date", ".metadata.creationTimestamp")
	return crdBuilder.Build()
}
