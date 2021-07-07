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

package v1alpha1

var (
	// BlockDeviceResourceKind is the kind of block device CRD
	BlockDeviceResourceKind = "BlockDevice"
	// BlockDeviceResourceListKind is the list kind for block device
	BlockDeviceResourceListKind = "BlockDeviceList"
	// BlockDeviceResourcePlural is the plural form used for block device
	BlockDeviceResourcePlural = "blockdevices"
	// BlockDeviceResourceShort is the short name used for block device CRD
	BlockDeviceResourceShort = "bd"
	// BlockDeviceResourceName is the name of the block device resource
	BlockDeviceResourceName = BlockDeviceResourcePlural + "." + GroupVersion.Group

	// BlockDeviceClaimResourceKind is the kind of block device claim CRD
	BlockDeviceClaimResourceKind = "BlockDeviceClaim"
	// BlockDeviceClaimResourceListKind is the list kind for block device claim
	BlockDeviceClaimResourceListKind = "BlockDeviceClaimList"
	// BlockDeviceClaimResourcePlural is the plural form used for block device claim
	BlockDeviceClaimResourcePlural = "blockdeviceclaims"
	// BlockDeviceClaimResourceShort is the short name used for block device claim CRD
	BlockDeviceClaimResourceShort = "bdc"
	// BlockDeviceClaimResourceName is the name of the block device claim resource
	BlockDeviceClaimResourceName = BlockDeviceClaimResourcePlural + "." + GroupVersion.Group
)
