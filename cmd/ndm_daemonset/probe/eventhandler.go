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

// addDiskEvent fill disk details from different probes and push it to etcd
func (pe *ProbeEvent) addDiskEvent(msg controller.EventMessage) {
	diskList, err := pe.Controller.ListDiskResource()
	if err != nil {
		klog.Error(err)
		go pe.initOrErrorEvent()
		return
	}
	isErrorDuringUpdate := false
	for _, diskDetails := range msg.Devices {
		klog.Info("Processing details for ", diskDetails.ProbeIdentifiers.Uuid)
		pe.Controller.FillDiskDetails(diskDetails)
		// if ApplyFilter returns true then we process the event further
		if !pe.Controller.ApplyFilter(diskDetails) {
			continue
		}
		klog.Info("Processed details for ", diskDetails.ProbeIdentifiers.Uuid)
		oldDr := pe.Controller.GetExistingDiskResource(diskList, diskDetails.ProbeIdentifiers.Uuid)
		// if old DiskCR doesn't exist and partition is found, it is ignored since we don't need info
		// of partition if disk as a whole is ignored
		if oldDr == nil && len(diskDetails.PartitionData) != 0 {
			klog.Info("Skipping partition of already excluded disk ", diskDetails.ProbeIdentifiers.Uuid)
			continue
		}
		// if diskCR is already present, and udev event is generated for partition, append the partition info
		// to the diskCR
		if oldDr != nil && len(diskDetails.PartitionData) != 0 {
			newDrCopy := oldDr.DeepCopy()
			klog.Info("Appending partition data to ", diskDetails.ProbeIdentifiers.Uuid)
			newDrCopy.Spec.PartitionDetails = append(newDrCopy.Spec.PartitionDetails, diskDetails.ToPartition()...)
			pe.Controller.UpdateDisk(*newDrCopy, oldDr)
		} else {
			pe.Controller.PushDiskResource(oldDr, diskDetails)

			/*
			 * There will be one blockdevice CR for each physical Disk.
			 * For network devices like LUN there will be a blockdevice
			 * CR but no disk CR. 1 to N mapping would be valid case
			 * where disk have more than one partitions.
			 * TODO: Need to check if udev event is for physical disk
			 * and based on that need to create disk CR or blockdevice CR
			 * or both.
			 */
			deviceDetails := pe.Controller.NewDeviceInfoFromDiskInfo(diskDetails)
			if deviceDetails != nil {
				klog.Infof("DeviceDetails:%#v", deviceDetails)
				deviceList, err := pe.Controller.ListBlockDeviceResource()
				if err != nil {
					klog.Error(err)
					go pe.initOrErrorEvent()
					return
				}
				oldDvr := pe.Controller.GetExistingBlockDeviceResource(deviceList, deviceDetails.UUID)
				err = pe.Controller.PushBlockDeviceResource(oldDvr, deviceDetails)
				if err != nil {
					klog.Errorf("error pushing block device resource to etcd: %v", err)
					isErrorDuringUpdate = true
				}
			}
		}
		/// update the list of DiskCRs
		diskList, err = pe.Controller.ListDiskResource()
		// if the listing errored out, or there was an error during pushing the resource
		// to etcd, we do a rescan.
		if err != nil || isErrorDuringUpdate {
			klog.Error(err)
			go pe.initOrErrorEvent()
			return
		}
	}
}

// deleteEvent deactivate disk/blockdevice resource using uuid from etcd
func (pe *ProbeEvent) deleteEvent(msg controller.EventMessage) {
	diskOk := pe.deleteDisk(msg)
	blockDeviceOk := pe.deleteBlockDevice(msg)

	// when one disk is removed from node and entry related to
	// that disk is not present in etcd,  in that case it
	// again rescan full system and update etcd accordingly.
	if !diskOk || !blockDeviceOk {
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
		bdUUID := pe.Controller.DiskToDeviceUUID(diskDetails.ProbeIdentifiers.Uuid)
		oldBDResource := pe.Controller.GetExistingBlockDeviceResource(bdList, bdUUID)
		if oldBDResource == nil {
			ok = false
			continue
		}
		pe.Controller.DeactivateBlockDevice(*oldBDResource)
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
		oldDiskResource := pe.Controller.GetExistingDiskResource(diskList, diskDetails.ProbeIdentifiers.Uuid)
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
