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

package probe

import (
	"fmt"
	"github.com/openebs/node-disk-manager/blockdevice"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/partition"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
)

const (
	internalUUIDAnnotation = "internal.openebs.io/uuid-scheme"
	legacyUUIDScheme       = "legacy"
)

func (pe *ProbeEvent) addBlockDevice(bd blockdevice.BlockDevice) error {

	pe.addOrUpdateDeviceToCache(bd)

	bdAPIList, err := pe.Controller.ListBlockDeviceResource(true)
	if err != nil {
		klog.Error(err)
		return err
	}

	if ok, err := pe.deviceInUse(bd, bdAPIList); err != nil {
		klog.Errorf("error checking device: %s in use by zfs-localPV or mayastor. Error: %v", bd.DevPath, err)
		return err
	} else if !ok {
		// device in use by zfs local pv or mayastor
		klog.V(4).Infof("processed device: %s being used by zfs-localPV / mayastor", bd.DevPath)
		return nil
	}

	if ok, err := pe.upgradeBD(bd, bdAPIList); err != nil {
		klog.Errorf("upgrade of device: %s failed. Error: %v", bd.DevPath, err)
		return err
	} else if !ok {
		klog.V(4).Infof("device: %s upgraded", bd.DevPath)
		return nil
	}

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

func (pe *ProbeEvent) upgradeBD(bd blockdevice.BlockDevice, bdAPIList *apis.BlockDeviceList) (bool, error) {
	// if the device is a partition and parent is in use, then the event
	// is skipped and no further processing is required.
	if bd.DeviceAttributes.DeviceType == libudevwrapper.UDEV_PARTITION {
		parentBD := pe.Controller.BDHierarchy[bd.DevPath]
		if parentBD.DevUse.InUse {
			klog.V(4).Infof("parent device: %s of %s is in use, hence ignoring event", parentBD.DevPath, bd.DevPath)
			return false, nil
		}
	}

	// if not in use, there is no need of upgrade.
	if !bd.DevUse.InUse {
		return true, nil
	}

	// device is in use by cstor or localpv
	klog.V(4).Infof("device: %s in use by cstor / localPV. hence ignoring event", bd.DevPath)

	// try to generate old UUID for the device
	legacyUUID, isVirt := generateLegacyUUID(bd)
	klog.V(4).Infof("tried generating legacy UUID: %s for device: %s", legacyUUID, bd.DevPath)

	legacyBDResource := pe.Controller.GetExistingBlockDeviceResource(bdAPIList, legacyUUID)

	// blockdevice with old uuid exists and is in use. The internal annotation will be added.
	if legacyBDResource != nil {
		bd.UUID = legacyUUID
		klog.V(4).Infof("blockdevice resource with legacy UUID exists, adding internal annotaion")
		// we used old algorithm for this BD, update the details, with old annotation
		deviceInfo := pe.Controller.NewDeviceInfoFromBlockDevice(&bd)
		if legacyBDResource.Annotations == nil {
			legacyBDResource.Annotations = make(map[string]string)
		}
		legacyBDResource.Annotations[internalUUIDAnnotation] = legacyUUIDScheme
		err := pe.Controller.PushBlockDeviceResource(legacyBDResource, deviceInfo)
		if err != nil {
			klog.Errorf("adding %s:%s annotation on %s failed. Error: %v", internalUUIDAnnotation, legacyUUIDScheme, bd.UUID)
			return false, err
		}

	}
	// If the resource does not exist, it can happen in 2 cases:
	// 1. This device do not use the old legacyUUID scheme,
	// 2. The device uses old scheme, but is a virtual drive and hostname or path has changed
	//
	// There can also be cases of manually created blockdevices, but those should be excluded by the filter.
	klog.V(4).Infof("device: %s may be virtual / device uses new uuid scheme", bd.DevPath)

	// try generating new UUID and if it exists, uses new scheme
	// return and further processing required
	uuid, ok := generateUUID(bd)
	klog.V(4).Infof("tried generating new UUID: %s, for device: %s", uuid, bd.DevPath)
	if ok {
		// check if bd with new uuid exists
		// uses the new algorithm,
		if pe.Controller.GetExistingBlockDeviceResource(bdAPIList, uuid) != nil {
			klog.V(4).Infof("device: %s uses GPT based uuid algorithm", bd.DevPath)
			return true, nil
		} else {
			// should never reach this case.
			// only reaches if manually created blockdevices are present.
			// this should have been filtered out in filter configs, no firther processing required.
			return false, nil
		}
	}

	klog.V(4).Infof("device: %s uses legacy uuid scheme with path/hostname", bd.DevPath)
	// which means path is used and path or hostname has changed, or is not an NDM managed device that has path/ it should be filtered
	// here we have to always create a new resource
	if isVirt {
		// creating new bd resource with legacy uuid
		klog.V(4).Infof("creating BD resource: %s for virtual device: %s", legacyUUID, bd.DevPath)
		bd.UUID = legacyUUID
		// we used old algorithm for this BD, update the details, with old annotation
		deviceInfo := pe.Controller.NewDeviceInfoFromBlockDevice(&bd)

		legacyVirtBDResource := deviceInfo.ToDevice()
		if legacyVirtBDResource.Annotations == nil {
			legacyVirtBDResource.Annotations = make(map[string]string)
		}
		legacyVirtBDResource.Annotations[internalUUIDAnnotation] = legacyUUIDScheme
		err := pe.Controller.CreateBlockDevice(legacyVirtBDResource)
		if err != nil {
			klog.Errorf("adding %s:%s annotation on %s failed. Error: %v", internalUUIDAnnotation, legacyUUIDScheme, bd.UUID)
			return false, err
		}
	}
	// log what is the case
	return false, nil
}
