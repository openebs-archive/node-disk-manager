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
	"fmt"
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/features"
	"github.com/openebs/node-disk-manager/pkg/partition"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/openebs/node-disk-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
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
		go pe.initOrErrorEvent()
		return
	}

	isGPTBasedUUIDEnabled := pe.Controller.FeatureGates.IsEnabled(features.GPTBasedUUID)

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
			err := pe.addBlockDevice(*device)
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
		go pe.initOrErrorEvent()
	}
}

// deleteBlockDeviceEvent deactivate blockdevice resource using uuid from etcd
func (pe *ProbeEvent) deleteBlockDeviceEvent(msg controller.EventMessage) {
	bdAPIList, err := pe.Controller.ListBlockDeviceResource(false)
	if err != nil {
		klog.Error(err)
	}

	isDeactivated := true
	isGPTBasedUUIDEnabled := pe.Controller.FeatureGates.IsEnabled(features.GPTBasedUUID)

	for _, device := range msg.Devices {
		// create the new UUID for removing the device
		if isGPTBasedUUIDEnabled {
			_, ok := pe.Controller.BDHierarchy[device.DevPath]
			if !ok {
				klog.Infof("Disk %s not in hierarchy", device.DevPath)
				// not in hierarchy continue
				continue
			}
			// remove from the hierarchy
			delete(pe.Controller.BDHierarchy, device.DevPath)

			uuid, ok := generateUUID(*device)
			// this can happen if the device didn't have any unique identifiers
			if !ok {
				klog.Info("could not recreate UUID while removing device")
				continue
			}
			// use the new UUID for deactivating the blockdevice
			device.UUID = uuid
		}

		existingBlockDeviceResource := pe.Controller.GetExistingBlockDeviceResource(bdAPIList, device.UUID)
		if existingBlockDeviceResource == nil {
			// do nothing, may be the disk was filtered, or it was not created
			isDeactivated = false
			continue
		}
		pe.Controller.DeactivateBlockDevice(*existingBlockDeviceResource)
	}

	// rescan only if GPT based UUID is disabled.
	if !isDeactivated && !isGPTBasedUUIDEnabled {
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
		klog.Error(err)
	}
}

// generateUUID creates a new UUID based on the algorithm proposed in
// https://github.com/openebs/openebs/pull/2666
func generateUUID(bd blockdevice.BlockDevice) (string, bool) {
	var ok bool
	var uuidField, uuid string

	// select the field which is to be used for generating UUID
	//
	// Serial number is not used directly for UUID generation. This is because serial number is not
	// unique in some cloud environments. For example, in GCP the serial number is
	// configurable by the --device-name flag while attaching the disk.
	// If this flag is not provided, GCP automatically assigns the serial number
	// which is unique only to the node. Therefore Serial number is used only in cases
	// where the disk has a WWN.
	//
	// If disk has WWN, a combination of WWN+Serial will be used. This is done because there are cases
	// where the disks has same WWN but different serial. It is seen in some storage arrays.
	// All the LUNs will have same WWN, but different serial.
	//
	// PartitionTableUUID is not used for UUID generation in NDM. The only case where the disk has a PartitionTable
	// and not partition is when, the user has manually created a partition table without writing any actual partitions.
	// This means NDM will have to give its consumers the entire disk, i.e consumers will have access to the sectors
	// where partition table is written. If consumers decide to reformat or erase the disk completely the partition
	// table UUID is also lost, making NDM unable to identify the disk. Hence, even if a partition table is present
	// NDM will rewrite it and create a new GPT table and a single partition. Thus consumers will have access only to
	// the partition and the unique data will be stored in sectors where consumers do not have access.

	switch {
	case bd.DeviceAttributes.DeviceType == libudevwrapper.UDEV_PARTITION:
		// The partition entry UUID is used when a partition (/dev/sda1) is processed. The partition UUID should be used
		// if available, other than the partition table UUID, because multiple partitions can have the same partition table
		// UUID, but each partition will have a different UUID.
		klog.Infof("device(%s) is a partition, using partition UUID: %s", bd.DevPath, bd.PartitionInfo.PartitionEntryUUID)
		uuidField = bd.PartitionInfo.PartitionEntryUUID
		ok = true
	case len(bd.DeviceAttributes.WWN) > 0:
		// if device has WWN, both WWN and Serial will be used for UUID generation.
		klog.Infof("device(%s) has a WWN, using WWN: %s and Serial: %s",
			bd.DevPath,
			bd.DeviceAttributes.WWN, bd.DeviceAttributes.Serial)
		uuidField = bd.DeviceAttributes.WWN +
			bd.DeviceAttributes.Serial
		ok = true
	case len(bd.FSInfo.FileSystemUUID) > 0:
		klog.Infof("device(%s) has a filesystem, using filesystem UUID: %s", bd.DevPath, bd.FSInfo.FileSystemUUID)
		uuidField = bd.FSInfo.FileSystemUUID
		ok = true
	}

	if ok {
		uuid = libudevwrapper.NDMBlockDevicePrefix + util.Hash(uuidField)
		klog.Infof("generated uuid: %s for device: %s", uuid, bd.DevPath)
	}

	return uuid, ok
}

func (pe *ProbeEvent) addBlockDevice(bd blockdevice.BlockDevice) error {

	bdAPIList, err := pe.Controller.ListBlockDeviceResource(true)
	if err != nil {
		klog.Error(err)
		return err
	}

	// check if the device already exists in the cache
	_, ok := pe.Controller.BDHierarchy[bd.DevPath]
	if ok {
		klog.V(4).Infof("device: %s already exists in cache, "+
			"the event was likely generated by a partition table re-read", bd.DevPath)
	}
	if !ok {
		klog.V(4).Infof("device: %s does not exist in cache, "+
			"the device is now connected to this node", bd.DevPath)
	}

	// in either case, whether is existed or not, we will update with the latest BD into the cache
	pe.Controller.BDHierarchy[bd.DevPath] = bd

	/*
		Cases when an add event is generated
		1. A new disk is added to the cluster to this node -  the disk is first time in this cluster
		2. A new disk is added to this node -  the disk was already present in the cluster and it was moved to this node
		3. A disk was detached and reconnected to this node
		4. An add event due to partition table reread . This may cause events to be generated without the disk
			being physically removed this node - (when a new partition is created on the device also, its the same case)
	*/

	// check if the disk can be uniquely identified. we try to generate the UUID for the device
	klog.V(4).Infof("checking if device: %s can be uniquely identified", bd.DevPath)
	uuid, ok := generateUUID(bd)
	// if UUID cannot be generated create a GPT partition on the device
	if !ok {
		klog.V(4).Infof("device: %s cannot be uniquely identified", bd.DevPath)
		if len(bd.DependentDevices.Partitions) > 0 ||
			len(bd.DependentDevices.Holders) > 0 {
			klog.V(4).Infof("device: %s has holders/partitions. %+v", bd.DevPath, bd.DependentDevices)
		} else {
			klog.Infof("starting to create partition on device: %s", bd.DevPath)
			d := partition.Disk{
				DevPath:          bd.DevPath,
				DiskSize:         bd.Capacity.Storage,
				LogicalBlockSize: uint64(bd.DeviceAttributes.LogicalBlockSize),
			}
			if err := d.CreateSinglePartition(); err != nil {
				klog.Errorf("error creating partition for %s, %v", bd.DevPath, err)
				return err
			}
			klog.Infof("created new partition in %s", bd.DevPath)
			return nil
		}
	} else {
		bd.UUID = uuid
		klog.V(4).Infof("uuid: %s has been generated for device: %s", uuid, bd.DevPath)
		bdAPI, err := pe.Controller.GetBlockDevice(uuid)

		if errors.IsNotFound(err) {
			klog.V(4).Infof("device: %s, uuid: %s not found in etcd", bd.DevPath, uuid)
			/*
				Cases when the BlockDevice is not found in etcd
				1. The device is appearing in this cluster for the first time
				2. The device had partitions and BlockDevice was not created
			*/

			if bd.DeviceAttributes.DeviceType == libudevwrapper.UDEV_PARTITION {
				klog.V(4).Infof("device: %s is partition", bd.DevPath)
				klog.V(4).Info("checking if device has a parent")
				// check if device has a parent that is claimed
				parentBD, ok := pe.Controller.BDHierarchy[bd.DependentDevices.Parent]
				if !ok {
					klog.V(4).Infof("unable to find parent device for device: %s", bd.DevPath)
					return fmt.Errorf("cannot get parent device for device: %s", bd.DevPath)
				}

				klog.V(4).Infof("parent device: %s found for device: %s", parentBD.DevPath, bd.DevPath)
				klog.V(4).Infof("checking if parent device can be uniquely identified")
				parentUUID, parentOK := generateUUID(parentBD)
				if !parentOK {
					klog.V(4).Infof("unable to generate UUID for parent device, may be a device without WWN")
					// cannot generate UUID for parent, may be a device without WWN
					// used the new algorithm to create partitions
					return pe.createBlockDeviceResourceIfNoHolders(bd, bdAPIList)
				}

				klog.V(4).Infof("uuid: %s generated for parent device: %s", parentUUID, parentBD.DevPath)

				parentBDAPI, err := pe.Controller.GetBlockDevice(parentUUID)

				if errors.IsNotFound(err) {
					// parent not present in etcd, may be device without wwn or had partitions/holders
					klog.V(4).Infof("parent device: %s, uuid: %s not found in etcd", parentBD.DevPath, parentUUID)
					return pe.createBlockDeviceResourceIfNoHolders(bd, bdAPIList)
				}

				if err != nil {
					klog.Error(err)
					return err
					// get call failed
				}

				if parentBDAPI.Status.ClaimState != apis.BlockDeviceUnclaimed {
					// device is in use, and the consumer is doing something
					// do nothing
					klog.V(4).Infof("parent device: %s is in use, device: %s can be ignored", parentBD.DevPath, bd.DevPath)
					return nil
				} else {
					// the consumer created some partitions on the disk.
					// So the parent BD need to be deactivated and partition BD need to be created.
					// 1. deactivate parent
					// 2. create resource for partition

					pe.Controller.DeactivateBlockDevice(*parentBDAPI)
					deviceInfo := pe.Controller.NewDeviceInfoFromBlockDevice(&bd)
					existingBlockDeviceResource := pe.Controller.GetExistingBlockDeviceResource(bdAPIList, deviceInfo.UUID)
					err := pe.Controller.PushBlockDeviceResource(existingBlockDeviceResource, deviceInfo)
					if err != nil {
						klog.Error(err)
						return err
					}
					return nil
				}

			}

			if bd.DeviceAttributes.DeviceType != libudevwrapper.UDEV_PARTITION &&
				len(bd.DependentDevices.Partitions) > 0 {
				klog.V(4).Infof("device: %s has partitions: %+v", bd.DevPath, bd.DependentDevices.Partitions)
				return nil
			}

			return pe.createBlockDeviceResourceIfNoHolders(bd, bdAPIList)
		}

		if err != nil {
			klog.Errorf("querying etcd failed: %+v", err)
			return err
		}

		if bdAPI.Status.ClaimState != apis.BlockDeviceUnclaimed {
			klog.V(4).Infof("device: %s is in use. update the details of the blockdevice", bd.DevPath)
			deviceInfo := pe.Controller.NewDeviceInfoFromBlockDevice(&bd)
			err = pe.Controller.PushBlockDeviceResource(bdAPI, deviceInfo)
			if err != nil {
				klog.Errorf("updating block device resource failed: %+v", err)
				return err
			}
			return nil
		}

		klog.V(4).Infof("creating resource for device: %s with uuid: %s", bd.DevPath, bd.UUID)
		deviceInfo := pe.Controller.NewDeviceInfoFromBlockDevice(&bd)
		existingBlockDeviceResource := pe.Controller.GetExistingBlockDeviceResource(bdAPIList, deviceInfo.UUID)
		err = pe.Controller.PushBlockDeviceResource(existingBlockDeviceResource, deviceInfo)
		if err != nil {
			klog.Errorf("creation of resource failed: %+v", err)
			return err
		}
		return nil
	}
	return nil
}

// createBlockDeviceResourceIfNoHolders creates/updates a blockdevice resource if it does not have any
// holder devices
func (pe *ProbeEvent) createBlockDeviceResourceIfNoHolders(bd blockdevice.BlockDevice, bdAPIList *apis.BlockDeviceList) error {
	if len(bd.DependentDevices.Holders) > 0 {
		klog.V(4).Infof("device: %s has holder devices: %+v", bd.DevPath, bd.DependentDevices.Holders)
		klog.V(4).Infof("skip creating BlockDevice resource")
		return nil
	}

	klog.V(4).Infof("creating block device resource for device: %s with uuid: %s", bd.DevPath, bd.UUID)
	deviceInfo := pe.Controller.NewDeviceInfoFromBlockDevice(&bd)
	existingBlockDeviceResource := pe.Controller.GetExistingBlockDeviceResource(bdAPIList, deviceInfo.UUID)
	err := pe.Controller.PushBlockDeviceResource(existingBlockDeviceResource, deviceInfo)
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}
