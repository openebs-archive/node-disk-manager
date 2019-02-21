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
	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
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

// addEvent fill disk/device details from different probes and push it to etcd
func (pe *ProbeEvent) addEvent(msg controller.EventMessage) {
	diskList, err := pe.Controller.ListDiskResource()
	if err != nil {
		glog.Error(err)
		go pe.initOrErrorEvent()
		return
	}
	for _, diskDetails := range msg.Devices {
		glog.Info("Processing details for ", diskDetails.ProbeIdentifiers.Uuid)
		pe.Controller.FillDiskDetails(diskDetails)

		// if ApplyFilter returns true then we process the event further
		glog.Info("applying filters to ", diskDetails.ProbeIdentifiers.Uuid)
		if !pe.Controller.ApplyFilter(diskDetails) {
			continue
		}

		glog.Info("processed details for ", diskDetails.ProbeIdentifiers.Uuid)
		oldDr := pe.Controller.GetExistingDiskResource(diskList, diskDetails.ProbeIdentifiers.Uuid)

		if diskDetails.DiskType == libudevwrapper.UDEV_DISK {
			// update etcd with the new disk details
			pe.Controller.PushDiskResource(oldDr, diskDetails)
		} else if diskDetails.DiskType == libudevwrapper.UDEV_PARTITION {
			// append the partition data to disk details and push to etcd
			glog.Info("Appending partition data to ", diskDetails.ProbeIdentifiers.Uuid)
			newDrCopy := oldDr.DeepCopy()
			newDrCopy.Spec.PartitionDetails = append(newDrCopy.Spec.PartitionDetails, diskDetails.ToPartition()...)
			pe.Controller.UpdateDisk(*newDrCopy, oldDr)
		}

		if diskDetails.DiskType == libudevwrapper.UDEV_DISK {
			/*
			 * There will be one device CR for each physical Disk.
			 * For network devices like LUN there will be a device
			 * CR but no disk CR. 1 to N mapping would be valid case
			 * where disk have more than one partitions.
			 * TODO: Need to check if udev event is for physical disk
			 * and based on that need to create disk CR or device CR
			 * or both.
			 */
			deviceDetails := pe.Controller.NewDeviceInfoFromDiskInfo(diskDetails)
			if deviceDetails != nil {
				glog.Infof("DeviceDetails:%#v", deviceDetails)
				deviceList, err := pe.Controller.ListDeviceResource()
				if err != nil {
					glog.Error(err)
					go pe.initOrErrorEvent()
					return
				}
				oldDvr := pe.Controller.GetExistingDeviceResource(deviceList, deviceDetails.Uuid)
				pe.Controller.PushDeviceResource(oldDvr, deviceDetails)
			}
		}

		diskList, err = pe.Controller.ListDiskResource()
		if err != nil {
			glog.Error(err)
			go pe.initOrErrorEvent()
			return
		}
	}
}

// deleteEvent deactivates disk resource or remove device details
// using uuid from etcd
func (pe *ProbeEvent) deleteEvent(msg controller.EventMessage) {
	diskList, err := pe.Controller.ListDiskResource()
	if err != nil {
		glog.Error(err)
		go pe.initOrErrorEvent()
		return
	}
	mismatch := false
	// set mismatch = true when one disk is removed from node and
	// entry related that disk not present in etcd in that case it
	// again rescan full system and update etcd accordingly.
	for _, diskDetails := range msg.Devices {
		oldDr := pe.Controller.GetExistingDiskResource(diskList, diskDetails.ProbeIdentifiers.Uuid)
		if oldDr == nil {
			mismatch = true
			continue
		}
		if diskDetails.DiskType == libudevwrapper.UDEV_DISK {
			pe.Controller.DeactivateDisk(*oldDr)
		} else if diskDetails.DiskType == libudevwrapper.UDEV_PARTITION {
			glog.Info("Removing Partition Data from ", diskDetails.ProbeIdentifiers.Uuid)
			newDrCopy := oldDr.DeepCopy()
			// TODO : Delete the specific partition, instead of the last partition
			if len(newDrCopy.Spec.PartitionDetails) > 0 {
				// Currently the details of last partition will be removed
				newDrCopy.Spec.PartitionDetails = newDrCopy.Spec.PartitionDetails[:len(newDrCopy.Spec.PartitionDetails)-1]
			}
			pe.Controller.UpdateDisk(*newDrCopy, oldDr)
		}
	}
	if mismatch {
		go pe.initOrErrorEvent()
	}
}

// initOrErrorEvent rescan system and update disk resource this is
// used for initial setup and when any uid mismatch or error occurred.
func (pe *ProbeEvent) initOrErrorEvent() {
	udevProbe := newUdevProbe(pe.Controller)
	defer udevProbe.free()
	err := udevProbe.scan()
	if err != nil {
		glog.Error(err)
	}
}
