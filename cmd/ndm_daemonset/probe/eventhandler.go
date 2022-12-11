/*
Copyright 2018 OpenEBS Authors.

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

package probe

import (
	"errors"

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/features"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/openebs/node-disk-manager/pkg/util"
	"k8s.io/klog/v2"
)

// EventAction action type for disk events like attach or detach events
type EventAction string

const (
	// AttachEA is attach disk event name
	AttachEA EventAction = libudevwrapper.UDEV_ACTION_ADD
	// DetachEA is detach disk event name
	DetachEA EventAction = libudevwrapper.UDEV_ACTION_REMOVE
	// ChangeEA event is generated whenever some property of the blockdevice
	// changes in udev
	ChangeEA EventAction = libudevwrapper.UDEV_ACTION_CHANGE
)

var (
	ErrNeedRescan = errors.New("need rescan")
)

// ProbeEvent struct contain a copy of controller it will update disk resources
type ProbeEvent struct {
	Controller *controller.Controller
}

// addBlockDeviceEvent fill block device details from different probes and push it to etcd
func (pe *ProbeEvent) addBlockDeviceEvent(msg controller.EventMessage) {
	// bdAPIList is the list of all the BlockDevice resources in the cluster
	bdAPIList, err := pe.Controller.ListBlockDeviceResource(true)
	if err != nil {
		klog.Error(err)
		go Rescan(pe.Controller)
		return
	}

	isGPTBasedUUIDEnabled := features.FeatureGates.IsEnabled(features.GPTBasedUUID)

	isNeedRescan := false
	erroredDevices := make([]string, 0)

	// iterate through each block device and perform the add/update operation
	for _, device := range msg.Devices {
		klog.Infof("Processing details for %s", device.DevPath)
		pe.Controller.FillBlockDeviceDetails(device, msg.RequestedProbes...)

		// add all devices to the hierarchy cache, irrespective of whether they will be
		// filtered at a later stage. This is done so that a complete disk hierarchy is available
		// at all times by NDM. It also helps in device processing when complex filter configurations
		// are provided. Ref: https://github.com/openebs/openebs/issues/3321
		pe.addBlockDeviceToHierarchyCache(*device)

		// if ApplyFilter returns true then we process the event further
		if !pe.Controller.ApplyFilter(device) {
			continue
		}
		klog.Infof("Processed details for %s", device.DevPath)

		if isGPTBasedUUIDEnabled {
			if isParentOrSlaveDevice(*device, erroredDevices) {
				klog.Warningf("device: %s skipped, because the parent / slave device has errored", device.DevPath)
				continue
			}
			err := pe.addBlockDevice(*device, bdAPIList)
			if err != nil {
				isNeedRescan = true
				if !errors.Is(err, ErrNeedRescan) {
					erroredDevices = append(erroredDevices, device.DevPath)
					klog.Error(err)
				}
			}
		} else {
			// if GPTBasedUUID is disabled and the device type is partition,
			// the event can be skipped.
			if device.DeviceAttributes.DeviceType == blockdevice.BlockDeviceTypePartition {
				klog.Info("GPTBasedUUID disabled. skip creating block device resource for partition.")
				continue
			}
			deviceInfo := pe.Controller.NewDeviceInfoFromBlockDevice(device)

			existingBlockDeviceResource := pe.Controller.GetExistingBlockDeviceResource(bdAPIList, deviceInfo.UUID)
			err := pe.Controller.PushBlockDeviceResource(existingBlockDeviceResource, deviceInfo)
			if err != nil {
				isNeedRescan = true
				klog.Error(err)
			}
		}
	}

	if isNeedRescan {
		go Rescan(pe.Controller)
	}
}

// deleteBlockDeviceEvent deactivate blockdevice resource using uuid from etcd
func (pe *ProbeEvent) deleteBlockDeviceEvent(msg controller.EventMessage) {
	bdAPIList, err := pe.Controller.ListBlockDeviceResource(false)
	if err != nil {
		klog.Error(err)
	}

	isDeactivated := true
	isGPTBasedUUIDEnabled := features.FeatureGates.IsEnabled(features.GPTBasedUUID)

	for _, device := range msg.Devices {
		if isGPTBasedUUIDEnabled {
			_ = pe.deleteBlockDevice(*device, bdAPIList)
		} else {
			if device.DeviceAttributes.DeviceType == blockdevice.BlockDeviceTypePartition {
				klog.Info("GPTBasedUUID disabled. skip delete block device resource for partition.")
				continue
			}
			existingBlockDeviceResource := pe.Controller.GetExistingBlockDeviceResource(bdAPIList, device.UUID)
			if existingBlockDeviceResource == nil {
				// do nothing, may be the disk was filtered, or it was not created
				isDeactivated = false
				continue
			}
			pe.Controller.DeactivateBlockDevice(*existingBlockDeviceResource)
		}
	}

	// rescan only if GPT based UUID is disabled.
	if !isDeactivated && !isGPTBasedUUIDEnabled {
		go Rescan(pe.Controller)
	}
}

func (pe *ProbeEvent) changeBlockDeviceEvent(msg controller.EventMessage) {
	var err error

	if msg.AllBlockDevices {
		for _, bd := range pe.Controller.BDHierarchy {
			klog.Infof("Processing changes for %s", bd.DevPath)
			err = pe.changeBlockDevice(&bd, msg.RequestedProbes...)
			if err != nil {
				klog.Errorf("failed to update blockdevice: %v", err)
			}
		}
		return
	}

	for _, bd := range msg.Devices {
		// The bd in `msg.Devices` mostly doesn't contain any information other than the
		// DevPath. Get corresponding bd from cache since cache will have latest info
		// for the bd.
		cacheBD, ok := pe.Controller.BDHierarchy[bd.DevPath]
		klog.Infof("Processing changes for %s", cacheBD.DevPath)
		if ok {
			err = pe.changeBlockDevice(&cacheBD, msg.RequestedProbes...)
		} else {
			// Device not in heiracrhy cache. this shouldn't happen, but to recover
			// We use the mostly empty bd and run it through all probes
			klog.V(4).Info("could not find bd in heirarchy cache. ",
				"This shouldn't happen. Abandoning change.")
			err = nil
		}
		if err != nil {
			klog.Errorf("failed to update blockdevice: %v", err)
		}
	}

}

// isParentOrSlaveDevice check if any of the errored device is a parent / slave to the
// given blockdevice
func isParentOrSlaveDevice(bd blockdevice.BlockDevice, erroredDevices []string) bool {
	for _, erroredDevice := range erroredDevices {
		if bd.DependentDevices.Parent == erroredDevice {
			return true
		}
		if util.Contains(bd.DependentDevices.Slaves, erroredDevice) {
			return true
		}
	}
	return false
}
