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

package partition

import (
	"fmt"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/partition/gpt"
	"github.com/openebs/node-disk-manager/pkg/blkid"

	"k8s.io/klog/v2"
)

// TODO this package needs to be upstreamed to diskfs/go-diskfs

const (
	// BytesRequiredForGPTPartitionEntries is the total bytes required to store the GPT partition
	// entries. 128 bytes are required per partition, total no of partition supported by GPT = 128
	// Therefore, total bytes = 128*128
	BytesRequiredForGPTPartitionEntries = 16384

	// GPTPartitionStartByte is the byte on the disk at which the first partition starts.
	// Normally partition starts at 1MiB, (as done by fdisk utility). This is done to
	// align the partition start to physical block sizes on the disk.
	GPTPartitionStartByte = 1048576

	// NoOfLogicalBlocksForGPTHeader is the no. of logical blocks for the GPT header.
	NoOfLogicalBlocksForGPTHeader = 1

	// OpenEBSNDMPartitionName is the name meta info for openEBS created partitions.
	OpenEBSNDMPartitionName = "OpenEBS_NDM"
)

// Disk struct represents a disk which needs to be partitioned
type Disk struct {
	// DevPath is the /dev/sdX entry of the disk
	DevPath string
	// DiskSize is size of disk in bytes
	DiskSize uint64
	// LogicalBlockSize is the block size of the disk normally 512 or 4k
	LogicalBlockSize uint64

	table *gpt.Table

	disk *disk.Disk
}

// applyPartitionTable applies the partition table to the given disk.
func (d *Disk) applyPartitionTable() error {
	if len(d.table.Partitions) == 0 {
		return fmt.Errorf("no partitions specified in partition table")
	}

	err := d.disk.Partition(d.table)
	if err != nil {
		return fmt.Errorf("unable to create/write partition table. %v", err)
	}
	return nil
}

// createPartitionTable creates gpt partition table structure with protective MBR turned on
func (d *Disk) createPartitionTable() error {
	if d.DiskSize == 0 {
		klog.Errorf("disk %s has size zero", d.DevPath)
		return fmt.Errorf("disk size is zero, unable to initialize partition table")
	}
	if d.LogicalBlockSize == 0 {
		klog.Warningf("logical block size of %s not set, falling back to 512 bytes", d.DevPath)
		klog.Warning("partitioning may fail.")
		d.LogicalBlockSize = 512
	}
	// set protective MBR to true.
	// https://en.wikipedia.org/wiki/GUID_Partition_Table#Protective_MBR_(LBA_0)
	d.table = &gpt.Table{
		LogicalSectorSize: int(d.LogicalBlockSize),
		ProtectiveMBR:     true,
	}
	return nil
}

// addPartition is used to add a partition to the partition table.
// Currently only a single partition can be created i.e, The method can be called only once for a disk.
// TODO: @akhilerm, add method to create partition with given size
func (d *Disk) addPartition() error {
	var startSector, endSector uint64
	if len(d.table.Partitions) == 0 {
		// First sector of partition is aligned at 1MiB
		startSector = (GPTPartitionStartByte) / d.LogicalBlockSize
	}

	PrimaryPartitionTableSize := BytesRequiredForGPTPartitionEntries/d.LogicalBlockSize + NoOfLogicalBlocksForGPTHeader

	// last sector for the partition. Since GPT scheme contains a backup partition table at
	// the last blocks of the disk.
	endSector = (d.DiskSize / d.LogicalBlockSize) - PrimaryPartitionTableSize - 1

	partition := &gpt.Partition{
		Start: startSector,
		End:   endSector,
		Type:  gpt.LinuxFilesystem,
		Name:  OpenEBSNDMPartitionName,
	}
	d.table.Partitions = append(d.table.Partitions, partition)
	return nil
}

// CreateSinglePartition creates a single GPT partition on the disk
// that spans the entire disk
func (d *Disk) CreateSinglePartition() error {
	fd, err := diskfs.Open(d.DevPath)
	if err != nil {
		return fmt.Errorf("error opening disk fd for disk %s: %v", d.DevPath, err)
	}
	d.disk = fd

	// check for any existing partition table on the disk
	if _, err := d.disk.GetPartitionTable(); err == nil {
		klog.Errorf("aborting partition creation, disk %s already contains a known partition table", d.DevPath)
		return fmt.Errorf("disk %s contains a partition table, cannot create a single partition", d.DevPath)
	}

	// check for any existing filesystem on the disk
	deviceIdentifier := blkid.DeviceIdentifier{
		DevPath: d.DevPath,
	}
	if fs := deviceIdentifier.GetOnDiskFileSystem(); len(fs) != 0 {
		klog.Errorf("aborting partition creation, disk %s contains a known filesystem: %s", d.DevPath, fs)
		return fmt.Errorf("disk %s contains a known filesyste: %s, cannot create a single partition", d.DevPath, fs)
	}

	err = d.createPartitionTable()
	if err != nil {
		klog.Error("partition table initialization failed")
		return err
	}

	err = d.addPartition()
	if err != nil {
		klog.Error("could not add a partition to partition table")
		return err
	}

	err = d.applyPartitionTable()
	if err != nil {
		klog.Error("writing partition table to disk failed")
		return err
	}
	klog.Infof("created a single partition on disk %s", d.DevPath)
	return nil
}

// CreatePartitionTable create a GPT header on the disk
func (d *Disk) CreatePartitionTable() error {
	fd, err := diskfs.Open(d.DevPath)
	if err != nil {
		return fmt.Errorf("error opening disk fd for disk %s: %v", d.DevPath, err)
	}
	d.disk = fd

	// check for any existing partition table on the disk
	if _, err := d.disk.GetPartitionTable(); err == nil {
		klog.Errorf("aborting partition creation, disk %s already contains a known partition table", d.DevPath)
		return fmt.Errorf("disk %s contains a partition table, cannot create a new partition table", d.DevPath)
	}

	// check for any existing filesystem on the disk
	deviceIdentifier := blkid.DeviceIdentifier{
		DevPath: d.DevPath,
	}
	if fs := deviceIdentifier.GetOnDiskFileSystem(); len(fs) != 0 {
		klog.Errorf("aborting partition creation, disk %s contains a known filesystem: %s", d.DevPath, fs)
		return fmt.Errorf("disk %s contains a known filesyste: %s, cannot create a partition table", d.DevPath, fs)
	}

	err = d.createPartitionTable()
	if err != nil {
		klog.Error("partition table initialization failed")
		return err
	}

	err = d.disk.Partition(d.table)
	if err != nil {
		return fmt.Errorf("unable to create/write partition table. %v", err)
	}
	klog.Infof("created partition table on disk %s", d.DevPath)
	return nil
}
