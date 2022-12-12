/*
Copyright 2020 The OpenEBS Authors

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

package filter

import (
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/util"

	"k8s.io/klog/v2"
)

// NOTE: This is an internal filter used by NDM to validate the block devices.
// Devices having invalid values in various fields(Path, Capacity etc) will be
// filtered out.
//
// The filter cannot be configured or disabled by user.

var (
	deviceValidityFilterName  = "device validity filter" // filter valid devices
	deviceValidityFilterState = defaultEnabled           // filter state
)

// deviceValidityFilterRegister contains registration process of DeviceValidityFilter
var deviceValidityFilterRegister = func() {
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		return
	}

	var fi controller.FilterInterface = newDeviceValidityFilter(ctrl)
	newRegisterFilter := &registerFilter{
		name:       deviceValidityFilterName,
		state:      deviceValidityFilterState,
		fi:         fi,
		controller: ctrl,
	}
	newRegisterFilter.register()
}

// deviceValidityFilter contains controller and validator functions
type deviceValidityFilter struct {
	controller             *controller.Controller
	excludeValidationFuncs []validationFunc
}

// validationFunc is a function type for the validations to be performed on the
// block device
type validationFunc func(*blockdevice.BlockDevice) bool

// newDeviceValidityFilter returns new pointer to a deviceValidityFilter
func newDeviceValidityFilter(ctrl *controller.Controller) *deviceValidityFilter {
	return &deviceValidityFilter{
		controller: ctrl,
	}
}

// Start sets the validator functions to be used
func (dvf *deviceValidityFilter) Start() {
	dvf.excludeValidationFuncs = make([]validationFunc, 0)
	dvf.excludeValidationFuncs = append(dvf.excludeValidationFuncs,
		isValidDevPath,
		isValidCapacity,
		isValidDMDevice,
		isValidPartition,
	)
}

// Include returns true because no specific internal validations are done
// for a device.
func (dvf *deviceValidityFilter) Include(blockDevice *blockdevice.BlockDevice) bool {
	return true
}

// Exclude returns true if all the exclude validation function passes.
// i.e The given block device is a valid entry.
func (dvf *deviceValidityFilter) Exclude(blockDevice *blockdevice.BlockDevice) bool {
	for _, vf := range dvf.excludeValidationFuncs {
		if !vf(blockDevice) {
			return false
		}
	}
	return true
}

// isValidDevPath checks if the path is not empty
func isValidDevPath(bd *blockdevice.BlockDevice) bool {
	if len(bd.DevPath) == 0 {
		klog.V(4).Infof("device has an invalid dev path")
		return false
	}
	return true
}

// isValidCapacity checks if the device has a valid capacity
func isValidCapacity(bd *blockdevice.BlockDevice) bool {
	if bd.Capacity.Storage == 0 {
		klog.V(4).Infof("device: %s has invalid capacity", bd.DevPath)
		return false
	}
	return true
}

// isValidPartition checks if the blockdevice for the partition has a valid partition UUID
func isValidPartition(bd *blockdevice.BlockDevice) bool {
	if bd.DeviceAttributes.DeviceType == blockdevice.BlockDeviceTypePartition &&
		len(bd.PartitionInfo.PartitionEntryUUID) == 0 {
		klog.V(4).Infof("device: %s of device-type partition has invalid partition UUID", bd.DevPath)
		return false
	}
	return true
}

// isValidDMDevice checks if the blockdevice is a valid dm device
func isValidDMDevice(bd *blockdevice.BlockDevice) bool {
	if util.Contains(blockdevice.DeviceMapperDeviceTypes, bd.DeviceAttributes.DeviceType) &&
		len(bd.DMInfo.DMUUID) == 0 {
		klog.V(4).Infof("device: %s of device mapper type has invalid DM_UUID", bd.DevPath)
		return false
	}
	return true
}
