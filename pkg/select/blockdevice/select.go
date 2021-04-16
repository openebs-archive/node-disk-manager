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
	"fmt"
	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
)

// Filter selects a single block device from a list of block devices
func (c *Config) Filter(bdList *apis.BlockDeviceList) (*apis.BlockDevice, error) {
	if len(bdList.Items) == 0 {
		return nil, fmt.Errorf("no blockdevices found")
	}

	candidateDevices, err := c.getCandidateDevices(bdList)
	if err != nil {
		return nil, err
	}
	selectedDevice, err := c.getSelectedDevice(candidateDevices)
	if err != nil {
		return nil, err
	}
	return selectedDevice, nil
}

// getCandidateDevices selects a list of blockdevices from a given block device
// list based on criteria specified in the claim spec
func (c *Config) getCandidateDevices(bdList *apis.BlockDeviceList) (*apis.BlockDeviceList, error) {

	// filterKeys to be used for filtering, by default active and unclaimed filter is present
	filterKeys := []string{FilterActive,
		FilterUnclaimed,
		// do not consider any devices with legacy annotation for claiming
		FilterOutLegacyAnnotation,
		// remove block devices which do not have the blockdevice tag
		// if selector is present on the BDC, select only those devices
		// this applies to both manual and auto claiming.
		FilterBlockDeviceTag,
	}

	if c.ManualSelection {
		filterKeys = append(filterKeys,
			FilterBlockDeviceName,
		)
	} else {
		filterKeys = append(filterKeys,
			// Sparse BDs can be claimed only by manual selection. Therefore, all
			// sparse BDs will be filtered out in auto mode
			FilterOutSparseBlockDevices,
			FilterDeviceType,
			FilterVolumeMode,
			FilterNodeName,
		)
	}

	candidateBD := c.ApplyFilters(bdList, filterKeys...)

	if len(candidateBD.Items) == 0 {
		return nil, fmt.Errorf("no devices found matching the criteria")
	}

	return candidateBD, nil
}

// getSelectedDevice selects a single a block device based on the resource requirements
// requested by the claim
func (c *Config) getSelectedDevice(bdList *apis.BlockDeviceList) (*apis.BlockDevice, error) {
	if c.ManualSelection {
		return &bdList.Items[0], nil
	}

	// filterKeys for filtering based on resource requirements
	filterKeys := []string{FilterResourceStorage}

	selectedDevices := c.ApplyFilters(bdList, filterKeys...)

	if len(selectedDevices.Items) == 0 {
		return nil, fmt.Errorf("could not find a device with matching resource requirements")
	}

	// will use the first available block device
	return &selectedDevices.Items[0], nil
}
