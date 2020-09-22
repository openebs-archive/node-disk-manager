/*
Copyright 2019 The OpenEBS Authors

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

package blockdevice

// BlockDevice is an internal representation of any block device present on the system.
// All data related to that device will be held by this struct
//
// 1. Example blockdevice struct for a partition /dev/sda1
// 		{
// 				Identifier:{
//					UUID:blockdevice-4c25d69f9adc868f61e3d891cf3a5613
//					SysPath:/sys/dev/block/8:1
//					DevPath:/dev/sda1
//				}
// 				NodeAttributes:map[hostname:my-machine]
// 				FSInfo:{
// 					FileSystemUUID:7e7f160b-0e79-478b-b006-1ebc6d0050dd
// 					FileSystem:ext4
// 					MountPoint:[/home]
// 				}
// 				Parent:/dev/sda
// 				Partitions:[]
// 				Holders:[]
// 				Slaves:[]
// 				Status:{
// 					State:Active
// 					ClaimPhase:Unclaimed
// 				}
// 			}
//
// 2. Example blockdevice struct for a partition that is part of an LVM
// 		{
// 				Identifier:{
//					UUID:blockdevice-4c25d69f9adc868f61e3d891cf3a5613
//					SysPath:/sys/dev/block/8:1
//					DevPath:/dev/sda1
//				}
// 				NodeAttributes:map[hostname:my-machine]
// 				FSInfo:{
// 					FileSystemUUID:AQkPql-2MBI-O5cY-gn3O-EFvZ-66Oe-d4mnjD
// 					FileSystem:LVM2_member
// 					MountPoint:[]
// 				}
// 				Parent:/dev/sda
// 				Partitions:[]
// 				Holders:[/dev/dm-0]
// 				Slaves:[]
// 				Status:{
// 					State:Active
// 					ClaimPhase:Unclaimed
// 				}
// 			}
// 3. Example blockdevice struct for an LVM carved from nvme partition
// 		{
// 				Identifier:{
//					UUID:blockdevice-4c25d69f9adc868f61e3d891cf3a5613
//					SysPath:/sys/dev/block/253:0
//					DevPath:/dev/dm-0
//				}
// 				NodeAttributes:map[hostname:my-machine]
// 				FSInfo:{
// 					FileSystemUUID:7e7f160b-0e79-478b-b006-1ebc6d0050dd
//					FileSystem:ext4
// 					MountPoint:[]
// 				}
// 				Parent:
// 				Partitions:[]
// 				Holders:[]
// 				Slaves:[/dev/nvme0n1p1 /dev/nvme0n1p2]
// 				Status:{
// 					State:Active
// 					ClaimPhase:Unclaimed
// 				}
// 			}
//
type BlockDevice struct {

	// Identifier is the unique identifiers that can be used to identify this
	// blockdevice
	Identifier

	// NodeAttributes contains the details of the node on which
	// the BlockDevice is attached
	NodeAttributes NodeAttribute

	// Labels for this blockdevice. These labels will be used on the k8s resource that is created
	// optional
	Labels map[string]string

	// FSInfo contains the file system related information of this
	// BlockDevice if it exists
	FSInfo FileSystemInformation

	// Capacity contains the capacity related information for this blockdevice
	Capacity CapacityInformation

	// DevLinks contain the devlinks of this blockdevice
	DevLinks []DevLink

	// DeviceAttributes contains the attributes of this device, like hardcoded
	// information on the disk
	DeviceAttributes DeviceAttribute

	DevUse DeviceUsage

	// PartitionInfo contains details if this blockdevice is a partition
	PartitionInfo PartitionInformation

	// DependentDevices stores the dependent devices
	DependentDevices DependentBlockDevices

	// TemperatureInfo stores the temperature information of the drive
	TemperatureInfo TemperatureInformation

	// Status contains the state of the blockdevice
	Status Status
}

// Identifier represents the various identifiers that can be used to
// identify this blockdevice uniquely on the host
type Identifier struct {
	// UUID is a system generated unique ID for this blockdevice
	UUID string

	// SysPath is the syspath of this device.
	// Eg : /sys/dev/block/8:1
	SysPath string

	// DevPath is the name of the device in /dev
	DevPath string
}

// NodeAttribute is the representing the various attributes of the machine
// on which this block device is present
type NodeAttribute map[string]string

const (
	// HostName is the hostname of the system on which this BD is present
	HostName string = "hostname"

	// NodeName is the nodename (may be FQDN) on which this BD is present
	NodeName string = "nodename"

	// ZoneName is the zone in which the system is present.
	//
	// NOTE: Valid only for cloud providers
	ZoneName string = "zone"

	// RegionName is the region in which the system is present.
	//
	// NOTE: Valid only for cloud providers
	RegionName string = "region"
)

const (
	// SparseBlockDeviceType is the sparse blockdevice type
	SparseBlockDeviceType = "sparse"
	// BlockDeviceType is the type for blockdevice.
	BlockDeviceType = "blockdevice"

	// BlockDevicePrefix is the prefix used in UUIDs
	BlockDevicePrefix = BlockDeviceType + "-"

	// The following blockdevice types correspond to the types as seen by the host system
	// BlockDeviceTypeDisk represents a disk type
	BlockDeviceTypeDisk = "disk"

	// BlockDeviceTypePartition represents a partition
	BlockDeviceTypePartition = "partition"

	// BlockDeviceTypeLoop represents a loop device
	BlockDeviceTypeLoop = "loop"

	BlockDeviceTypeDMDevice = "dm"

	BlockDeviceTypeLVM = "lvm"

	BlockDeviceTypeCrypt = "crypt"
)

const (
	// DriveTypeHDD represents a rotating hard disk drive
	DriveTypeHDD = "HDD"

	// DriveTypeSSD represents a solid state drive
	DriveTypeSSD = "SSD"
)

// FileSystemInformation contains the filesystem and mount information of blockdevice, if present
type FileSystemInformation struct {
	// FileSystemUUID is the UUID of the filesystem on the blockdevice
	FileSystemUUID string

	// FileSystem is the filesystem present on the blockdevice
	FileSystem string

	// MountPoint is the list of mountpoints at which this blockdevice is mounted
	MountPoint []string
}

// CapacityInformation holds the capacity related information for the device
type CapacityInformation struct {
	// Storage is the storage capacity of this blockdevice
	// in bytes
	Storage uint64
}

// DeviceAttribute represents the hardcoded information on the device.
// It is not gauranteed that all these fields should be present for a given
// blockdevice
type DeviceAttribute struct {
	// DeviceType is the type of the blockdevice.
	// Values can be sparse/disk/partition/loop/lvm/raid etc
	DeviceType string

	// DriveType is the type of backing drive for this blockdevice. HDD/SSD
	DriveType string

	// IDType is the udev ID_TYPE field. This is solely used for purpose of UUID generation
	// using the old algorithm
	IDType string

	// PhysicalBlockSize is the physical block size in bytes
	// reported by /sys/class/block/sda/queue/physical_block_size
	//
	// This is the unit in which the disk actually reads and writes
	// the data atomically. Certain 4K sector devices may use a
	// 4K 'physical_block_size' internally but expose a finer-grained
	// 512 byte 'logical_block_size' to Linux. Also the current HDDs
	// are advanced format (https://en.wikipedia.org/wiki/Advanced_Format)
	// so they support 4k physical sector size.
	PhysicalBlockSize uint32

	// LogicalBlockSize is the logical block size in bytes
	// reported by /sys/class/block/sda/queue/logical_block_size
	//
	// The smallest size the drive is able to write. This can be
	// different than hw_sector_size as disks may lie about this
	// to support lower/higher sector size. This is important for
	// direct IOs, as IOs should be aligned on a 'logical_block_size'
	// boundaries.
	LogicalBlockSize uint32

	// HardwareSectorSize is the hardware sector size in bytes
	// reported by /sys/class/block/sda/queue/hw_sector_size
	//
	// This is the actual sector size of the disk
	HardwareSectorSize uint32

	// WWN
	WWN string

	// Vendor
	Vendor string

	// Model of the device
	// Eg : PersistentDisk, Virtaul_disk, QEMU_HARDDISK, EphemeralDisk
	Model string

	// Serial number of the device
	Serial string

	// FirmwareRevision
	FirmwareRevision string

	// Compliance is implemented specifications version i.e. SPC-1, SPC-2, etc
	Compliance string
}

// DevLink represents a type of dev link for a device. A device can have multiple
// kinds of devlink and each kind can have more than one link
type DevLink struct {
	// Kind is the type of devlink.
	// Eg: by-id, by-path, by-uuid, by-partuuid
	Kind string

	// Links is the list of actual links that point to the device
	Links []string
}

// TemperatureInformation stores the temperature information of the blockdevice
type TemperatureInformation struct {
	// TemperatureDataValid specifies whether the current temperature
	// data reported is valid or not
	TemperatureDataValid bool

	// CurrentTemperature is the temperature of the drive in celsius
	CurrentTemperature int16
}

// PartitionInformation contains information related to the partition, if this
// blockdevice is a partition
type PartitionInformation struct {
	// PartitionNumber is the partition number
	PartitionNumber uint8

	// PartitionEntryUUID is the UUID of the partition
	PartitionEntryUUID string

	// PartitionTableUUID is the UUID of the partition table
	PartitionTableUUID string

	// PartitionTableType is the type of the partition (dos/gpt)
	PartitionTableType string
}

// DependentBlockDevices contains path of all devices that are
// related to this BlockDevice
type DependentBlockDevices struct {
	// Parent is the parent device of this blockdevice, if it exists.
	// It will always be a single device.
	Parent string

	// Partitions is the list of partitions(again blockdevices) for
	// this blockdevice, if it exists.
	Partitions []string

	// Holders is the list of blockdevices that are held by this blockdevice.
	// eg: sda1 can hold dm-0. Then the list of sda1 will contain dm-0.
	Holders []string

	// Slaves is the list of blockdevices to which this blockdevice is a slave.
	// Slaves are represented in SysFS under the block device layer,
	// so they are dependent on the device
	// eg: dm-0 is a slave to sda1. Then the list of dm-0 will contain sda1
	Slaves []string
}

// DeviceUsage defines if the block device is used by any known storage engines
type DeviceUsage struct {
	InUse  bool
	UsedBy StorageEngine
}

// StorageEngine is a typed string for the storage engine
type StorageEngine string

const (
	// CStor
	CStor StorageEngine = "cstor"

	// ZFSLocalPV
	ZFSLocalPV StorageEngine = "zfs-localpv"

	// Mayastor
	Mayastor StorageEngine = "mayastor"

	// LocalPV
	LocalPV StorageEngine = "localpv"

	// Jiva
	Jiva StorageEngine = "jiva"
)

// Status is used to represent the status of the blockdevice
type Status struct {
	// State is the state of this BD like Active(Online), Inactive(Offline) or
	// Unknown
	State string

	// ClaimPhase is the phase of this BD when is it is being used by NDM consumers
	ClaimPhase string
}

const (
	// Active means blockdevice is available on the host machine
	Active string = "Active"
	// Inactive means blockdevice is currently not available on the host machine
	Inactive string = "Inactive"
	// Unknown means the state cannot be determined at this point of time
	Unknown string = "Unknown"

	// Claimed means the blockdevice is in use
	Claimed string = "Claimed"
	// Released means the blockdevice is not in use, but cannot be claimed,
	// because of some pending cleanup tasks
	Released string = "Released"
	// Unclaimed means the blockdevice is free and is available for
	// claiming
	Unclaimed string = "Unclaimed"
)

// Hierarchy is the block device hierarchy on the system.
// This map will be used as in memory cache for storing the current state of
// all block devices on the system. This cache helps to get the details of any
// dependent blockdevices from a device.
// eg: If at a certain point, we are processing sda1 and we need to get the UUID
// of its parent device. We will get /dev/sda is the parent of /dev/sda1.
// This will be used to query the cache and get/generate the UUID of /dev/sda.
type Hierarchy map[string]BlockDevice
