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
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/features"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"k8s.io/klog"
)

// EventAction action type for disk events like attach or detach events
type EventAction string

const (
	// AttachEA is attach disk event name
	AttachEA EventAction = libudevwrapper.UDEV_ACTION_ADD
	// DetachEA is detach disk event name
	DetachEA EventAction = libudevwrapper.UDEV_ACTION_REMOVE
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

	isErrorDuringUpdate := false
	// iterate through each block device and perform the add/update operation
	for _, device := range msg.Devices {
		klog.Infof("Processing details for %s", device.DevPath)
		pe.Controller.FillBlockDeviceDetails(device)
		// if ApplyFilter returns true then we process the event further
		if !pe.Controller.ApplyFilter(device) {
			continue
		}
		klog.Infof("Processed details for %s", device.DevPath)

		if isGPTBasedUUIDEnabled {
			err := pe.addBlockDevice(*device, bdAPIList)
			if err != nil {
				isErrorDuringUpdate = true
				klog.Error(err)
				// if error occurs we should start the scan again
				break
			}
		} else {
			// if GPTBasedUUID is disabled and the device type is partition,
			// the event can be skipped.
			if device.DeviceAttributes.DeviceType == libudevwrapper.UDEV_PARTITION {
				klog.Info("GPTBasedUUID disabled. skip creating block device resource for partition.")
				continue
			}
			deviceInfo := pe.Controller.NewDeviceInfoFromBlockDevice(device)

			existingBlockDeviceResource := pe.Controller.GetExistingBlockDeviceResource(bdAPIList, deviceInfo.UUID)
			err := pe.Controller.PushBlockDeviceResource(existingBlockDeviceResource, deviceInfo)
			if err != nil {
				isErrorDuringUpdate = true
				klog.Error(err)
			}
		}
	}

	if isErrorDuringUpdate {
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
