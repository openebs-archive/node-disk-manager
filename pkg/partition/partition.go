package partition

import (
	"fmt"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/partition/gpt"
	"k8s.io/klog"
)

const (
	// BytesRequiredForGPTHeader is the total bytes required to store the GPT header
	// 128 bytes are required per partition, total no of partition supported by GPT = 128
	// Therefore, total bytes = 128*128
	BytesRequiredForGPTHeader = 16384

	// GPTPartitionStartByte is the byte on the disk at which the first partition starts.
	// Normally partition starts at 1MiB, (as done by fdisk utility). This is done to
	// align the partition start to physical block sizes on the disk.
	GPTPartitionStartByte = 1048576
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
}

// ApplyPartitionTable applies the partition table to the given disk.
func (d *Disk) ApplyPartitionTable() error {
	if len(d.table.Partitions) == 0 {
		return fmt.Errorf("no partitions specified in partition table")
	}
	diskFD, err := diskfs.Open(d.DevPath)
	if err != nil {
		return fmt.Errorf("error opening diskFD for write: %v", err)
	}
	err = diskFD.Partition(d.table)
	if err != nil {
		return fmt.Errorf("unable to create partition table. %v", err)
	}
	return nil
}

// CreatePartitionTable creates gpt partition table structure with protective MBR turned on
func (d *Disk) CreatePartitionTable() error {
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

// AddPartition is used to add a partition to the partition table.
// Currently only a single partition can be created. i.e The method can be called only once for a disk
// TODO: @akhilerm, make use of partitionSize to create partition of any size
func (d *Disk) AddPartition(partitionSize uint64) error {
	var startSector, endSector uint64
	if len(d.table.Partitions) == 0 {
		// First sector of partition is aligned at 1MiB
		startSector = (GPTPartitionStartByte) / d.LogicalBlockSize
	}
	// 2 blocks for LBA 0 and LBA 1
	gptHeaderInSectors := BytesRequiredForGPTHeader/d.LogicalBlockSize + 2

	// last sector for the partition. Since GPT scheme contains a backup GPT header at
	// the last blocks of the disk. Size required for the GPT header
	endSector = (d.DiskSize / d.LogicalBlockSize) - gptHeaderInSectors

	partition := &gpt.Partition{
		Start: startSector,
		End:   endSector,
		Type:  gpt.LinuxFilesystem,
	}
	d.table.Partitions = append(d.table.Partitions, partition)
	return nil
}
