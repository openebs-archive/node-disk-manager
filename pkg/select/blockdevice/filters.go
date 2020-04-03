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

package blockdevice

import (
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/select/verify"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// FilterActive is the filter for getting active BDs
	FilterActive = "filterActive"
	// FilterUnclaimed is the filter for getting unclaimed BDs
	FilterUnclaimed = "filterUnclaimed"
	// FilterDeviceType is the filter for matching DeviceType like SSD, HDD, sparse
	FilterDeviceType = "filterDeviceType"
	// FilterVolumeMode is the  filter for filtering based on Volume Mode required
	FilterVolumeMode = "filterVolumeMode"
	// FilterBlockDeviceName is the filter for getting a BD based on a name
	FilterBlockDeviceName = "filterBlockDeviceName"
	// FilterResourceStorage is the filter for matching resource storage
	FilterResourceStorage = "filterResourceStorage"
	// FilterOutSparseBlockDevices is used to filter out sparse BDs
	FilterOutSparseBlockDevices = "filterSparseBlockDevice"
	// FilterNodeName is used to filter based on nodename
	FilterNodeName = "filterNodeName"
)

// filterFunc is the func type for the filter functions
type filterFunc func(original *apis.BlockDeviceList, spec *apis.DeviceClaimSpec) *apis.BlockDeviceList

var filterFuncMap = map[string]filterFunc{
	FilterActive:                filterActive,
	FilterUnclaimed:             filterUnclaimed,
	FilterDeviceType:            filterDeviceType,
	FilterVolumeMode:            filterVolumeMode,
	FilterBlockDeviceName:       filterBlockDeviceName,
	FilterResourceStorage:       filterResourceStorage,
	FilterOutSparseBlockDevices: filterOutSparseBlockDevice,
	FilterNodeName:              filterNodeName,
}

// ApplyFilters apply the filter specified in the filterkeys on the given BD List,
func (c *Config) ApplyFilters(bdList *apis.BlockDeviceList, filterKeys ...string) *apis.BlockDeviceList {
	filteredList := bdList
	for _, key := range filterKeys {
		filteredList = filterFuncMap[key](filteredList, c.ClaimSpec)
	}
	return filteredList
}

// filterActive filters out active Blockdevices from the BDList
func filterActive(originalBD *apis.BlockDeviceList, spec *apis.DeviceClaimSpec) *apis.BlockDeviceList {
	filteredBDList := &apis.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	for _, bd := range originalBD.Items {
		if bd.Status.State == controller.NDMActive {
			filteredBDList.Items = append(filteredBDList.Items, bd)
		}
	}
	return filteredBDList
}

// filterUnclaimed returns only unclaimed devices
func filterUnclaimed(originalBD *apis.BlockDeviceList, spec *apis.DeviceClaimSpec) *apis.BlockDeviceList {
	filteredBDList := &apis.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	for _, bd := range originalBD.Items {
		if bd.Status.ClaimState == apis.BlockDeviceUnclaimed {
			filteredBDList.Items = append(filteredBDList.Items, bd)
		}
	}
	return filteredBDList
}

// filterDeviceType returns only BDs which match the device type
func filterDeviceType(originalBD *apis.BlockDeviceList, spec *apis.DeviceClaimSpec) *apis.BlockDeviceList {

	// if no device type is specified, filter will not be effective
	if spec.DeviceType == "" {
		return originalBD
	}

	filteredBDList := &apis.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	for _, bd := range originalBD.Items {
		if bd.Spec.Details.DeviceType == spec.DeviceType {
			filteredBDList.Items = append(filteredBDList.Items, bd)
		}
	}
	return filteredBDList
}

// filterVolumeMode returns only BDs which matches the specified volume mode
func filterVolumeMode(originalBD *apis.BlockDeviceList, spec *apis.DeviceClaimSpec) *apis.BlockDeviceList {

	volumeMode := spec.Details.BlockVolumeMode

	// if volume mode is not specified in claim spec, this filter will not be effective
	if volumeMode == "" {
		return originalBD
	}

	filteredBDList := &apis.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	for _, bd := range originalBD.Items {
		switch volumeMode {
		case apis.VolumeModeBlock:
			// if blockvolume mode, FS and Mountpoint should be empty
			if bd.Spec.FileSystem.Type != "" || bd.Spec.FileSystem.Mountpoint != "" {
				continue
			}

		case apis.VolumeModeFileSystem:
			// in FSVolume Mode,
			// In BD: FS and Mountpoint should not be empty. If empty the BD
			// is removed by filter
			if bd.Spec.FileSystem.Type == "" || bd.Spec.FileSystem.Mountpoint == "" {
				continue
			}
			// In BDC: If DeviceFormat is specified, then it should match with BD,
			// else do not compare FSType. If the check fails, the BD is removed by the filter.
			if spec.Details.DeviceFormat != "" && bd.Spec.FileSystem.Type != spec.Details.DeviceFormat {
				continue
			}
		}
		filteredBDList.Items = append(filteredBDList.Items, bd)
	}
	return filteredBDList
}

// filterBlockDeviceName returns a single BD in the list, which matches the BDName
func filterBlockDeviceName(originalBD *apis.BlockDeviceList, spec *apis.DeviceClaimSpec) *apis.BlockDeviceList {
	filteredBDList := &apis.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	for _, bd := range originalBD.Items {
		if bd.Name == spec.BlockDeviceName {
			filteredBDList.Items = append(filteredBDList.Items, bd)
			break
		}
	}
	return filteredBDList
}

// filterResourceStorage gets the devices which match the storage resource requirement
func filterResourceStorage(originalBD *apis.BlockDeviceList, spec *apis.DeviceClaimSpec) *apis.BlockDeviceList {
	filteredBDList := &apis.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	capacity, _ := verify.GetRequestedCapacity(spec.Resources.Requests)

	for _, bd := range originalBD.Items {
		if bd.Spec.Capacity.Storage >= uint64(capacity) {
			filteredBDList.Items = append(filteredBDList.Items, bd)
			break
		}
	}
	return filteredBDList
}

// filterOutSparseBlockDevice returns only non sparse devices
func filterOutSparseBlockDevice(originalBD *apis.BlockDeviceList, spec *apis.DeviceClaimSpec) *apis.BlockDeviceList {
	filteredBDList := &apis.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	for _, bd := range originalBD.Items {
		if bd.Spec.Details.DeviceType != blockdevice.SparseBlockDeviceType {
			filteredBDList.Items = append(filteredBDList.Items, bd)
		}
	}
	return filteredBDList
}

func filterNodeName(originalBD *apis.BlockDeviceList, spec *apis.DeviceClaimSpec) *apis.BlockDeviceList {

	// if node name is not given in BDC, this filter will not work
	if len(spec.BlockDeviceNodeAttributes.NodeName) == 0 {
		return originalBD
	}

	filteredBDList := &apis.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	for _, bd := range originalBD.Items {
		if bd.Spec.NodeAttributes.NodeName == spec.BlockDeviceNodeAttributes.NodeName {
			filteredBDList.Items = append(filteredBDList.Items, bd)
		}
	}
	return filteredBDList
}
