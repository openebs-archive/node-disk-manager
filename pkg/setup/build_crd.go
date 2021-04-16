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

package setup

// import (
// 	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
// 	"github.com/openebs/node-disk-manager/pkg/crds"
// 	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
// )

// // buildBlockDeviceCRD is used to build the blockdevice CRD
// func buildBlockDeviceCRD() (*apiext.CustomResourceDefinition, error) {
// 	crdBuilder := crds.NewBuilder()
// 	crdBuilder.WithName(apis.BlockDeviceResourceName).
// 		WithGroup(apis.GroupName).
// 		WithVersion(apis.APIVersion).
// 		WithScope(apiext.NamespaceScoped).
// 		WithKind(apis.BlockDeviceResourceKind).
// 		WithListKind(apis.BlockDeviceResourceListKind).
// 		WithPlural(apis.BlockDeviceResourcePlural).
// 		WithShortNames([]string{apis.BlockDeviceResourceShort}).
// 		WithPrinterColumns("NodeName", "string", ".spec.nodeAttributes.nodeName").
// 		WithPriorityPrinterColumns("Path", "string", ".spec.path", 1).
// 		WithPriorityPrinterColumns("FSType", "string", ".spec.filesystem.fsType", 1).
// 		WithPrinterColumns("Size", "string", ".spec.capacity.storage").
// 		WithPrinterColumns("ClaimState", "string", ".status.claimState").
// 		WithPrinterColumns("Status", "string", ".status.state").
// 		WithPrinterColumns("Age", "date", ".metadata.creationTimestamp")
// 	return crdBuilder.Build()
// }

// // buildBlockDeviceClaimCRD is used to build the blockdevice claim CRD
// func buildBlockDeviceClaimCRD() (*apiext.CustomResourceDefinition, error) {
// 	crdBuilder := crds.NewBuilder()
// 	crdBuilder.WithName(apis.BlockDeviceClaimResourceName).
// 		WithGroup(apis.GroupName).
// 		WithVersion(apis.APIVersion).
// 		WithScope(apiext.NamespaceScoped).
// 		WithKind(apis.BlockDeviceClaimResourceKind).
// 		WithListKind(apis.BlockDeviceClaimResourceListKind).
// 		WithPlural(apis.BlockDeviceClaimResourcePlural).
// 		WithShortNames([]string{apis.BlockDeviceClaimResourceShort}).
// 		WithPrinterColumns("BlockDeviceName", "string", ".spec.blockDeviceName").
// 		WithPrinterColumns("Phase", "string", ".status.phase").
// 		WithPrinterColumns("Age", "date", ".metadata.creationTimestamp")
// 	return crdBuilder.Build()
// }
