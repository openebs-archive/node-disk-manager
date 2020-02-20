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
	bdList, err := pe.Controller.ListBlockDeviceResource()
	if err != nil {
		klog.Error(err)
		go pe.initOrErrorEvent()
		return
	}
	// iterate through each block device and perform the add/update operation
	for _, device := range msg.Devices {
		klog.Infof("Processing details for %s", device.UUID)
		pe.Controller.FillBlockDeviceDetails(device)
		// if ApplyFilter returns true then we process the event further
		if !pe.Controller.ApplyFilter(device) {
			continue
		}
		klog.Infof("Processed details for %s : %s", device.DevPath, device.UUID)
		deviceInfo := pe.Controller.NewDeviceInfoFromBlockDevice(device)

		existingBlockDeviceResource := pe.Controller.GetExistingBlockDeviceResource(bdList, deviceInfo.UUID)
		pe.Controller.PushBlockDeviceResource(existingBlockDeviceResource, deviceInfo)
	}
}

// deleteBlockDeviceEvent deactivate blockdevice resource using uuid from etcd
func (pe *ProbeEvent) deleteBlockDeviceEvent(msg controller.EventMessage) {
	blockDeviceOk := pe.deleteBlockDevice(msg)

	// when one disk is removed from node and entry related to
	// that disk is not present in etcd,  in that case it
	// again rescan full system and update etcd accordingly.
	if !blockDeviceOk {
		go pe.initOrErrorEvent()
	}
}

func (pe *ProbeEvent) deleteBlockDevice(msg controller.EventMessage) bool {
	bdList, err := pe.Controller.ListBlockDeviceResource()
	if err != nil {
		klog.Error(err)
		return false
	}
	ok := true
	for _, diskDetails := range msg.Devices {
		existingBlockDeviceResource := pe.Controller.GetExistingBlockDeviceResource(bdList, diskDetails.UUID)
		if existingBlockDeviceResource == nil {
			ok = false
			continue
		}
		pe.Controller.DeactivateBlockDevice(*existingBlockDeviceResource)
	}
	return ok
}

func (pe *ProbeEvent) deleteDisk(msg controller.EventMessage) bool {
	diskList, err := pe.Controller.ListDiskResource()
	if err != nil {
		klog.Error(err)
		return false
	}
	ok := true
	for _, diskDetails := range msg.Devices {
		oldDiskResource := pe.Controller.GetExistingDiskResource(diskList, diskDetails.UUID)
		if oldDiskResource == nil {
			ok = false
			continue
		}
		pe.Controller.DeactivateDisk(*oldDiskResource)
	}
	return ok
}

// initOrErrorEvent rescan system and update disk resource this is
// used for initial setup and when any uid mismatch or error occurred.
func (pe *ProbeEvent) initOrErrorEvent() {
	udevProbe := newUdevProbe(pe.Controller)
	defer udevProbe.free()
	err := udevProbe.scan()
	if err != nil {
		klog.Error(err)
	}
}
