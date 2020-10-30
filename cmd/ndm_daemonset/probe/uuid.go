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
	"os"
	"strings"

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/pkg/util"

	"k8s.io/klog"
)

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
	case bd.DeviceAttributes.DeviceType == blockdevice.BlockDeviceTypeLoop:
		// hostname and device name, i.e /dev/loopX will be used for generating uuid
		hostName, _ := os.Hostname()
		klog.Infof("device(%s) is a loop device, using node name: %s and path: %s", bd.DevPath, hostName, bd.DevPath)
		uuidField = hostName + bd.DevPath
		ok = true
	case util.Contains(blockdevice.DeviceMapperDeviceTypes, bd.DeviceAttributes.DeviceType):
		// if a DM device, use the DM uuid
		klog.Infof("device(%s) is a dm device, using DM UUID: %s", bd.DevPath, bd.DMInfo.DMUUID)
		// TODO add a check if DM uuid is present, else may need to add mitigation steps
		uuidField = bd.DMInfo.DMUUID
		ok = true
	case bd.DeviceAttributes.DeviceType == blockdevice.BlockDeviceTypePartition:
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
		uuid = blockdevice.BlockDevicePrefix + util.Hash(uuidField)
		klog.Infof("generated uuid: %s for device: %s", uuid, bd.DevPath)
	}

	return uuid, ok
}

// generate old UUID, returns true if the UUID has used path or hostname for generation.
func generateLegacyUUID(bd blockdevice.BlockDevice) (string, bool) {
	localDiskModels := []string{
		"EphemeralDisk",
		"Virtual_disk",
		"QEMU_HARDDISK",
	}
	uid := bd.DeviceAttributes.WWN +
		bd.DeviceAttributes.Model +
		bd.DeviceAttributes.Serial +
		bd.DeviceAttributes.Vendor
	uuidUsesPath := false
	if len(bd.DeviceAttributes.IDType) == 0 || util.Contains(localDiskModels, bd.DeviceAttributes.Model) {
		host, _ := os.Hostname()
		uid += host + bd.DevPath
		uuidUsesPath = true
	}
	uuid := blockdevice.BlockDevicePrefix + util.Hash(uid)

	return uuid, uuidUsesPath
}

// generateUUIDFromPartitionTable generates a blockdevice uuid from the partition table uuid.
// currently this is only used by zfs localPV
//
//TODO, this currently supports cases where a complete disk is used for ZFS localPV. If multiple
// partitions on the same disk are used for pools, each one should be shown as a separate BD.
// For achieving that partition uuid can be used, same as used in the generic UUID generation algorithm
func generateUUIDFromPartitionTable(bd blockdevice.BlockDevice) (string, bool) {
	uuidField := bd.PartitionInfo.PartitionTableUUID
	if len(uuidField) > 0 {
		return blockdevice.BlockDevicePrefix + util.Hash(uuidField), true
	}
	return "", false
}

// TODO move it to some other pkg
func isDM(devPath string) bool {
	devName := strings.Replace(devPath, "/dev/", "", 1)
	if devName[0:3] == "dm-" {
		return true
	}
	return false
}
