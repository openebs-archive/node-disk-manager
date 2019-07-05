package v1alpha1

const (
	// DiskResourceKind is the kind of Disk CRD
	DiskResourceKind = "Disk"
	// DiskResourceListKind is the list kind for Disk
	DiskResourceListKind = "DiskList"
	// DiskResourcePlural is the plural form used for disk
	DiskResourcePlural = "disks"
	// DiskResourceShort is the short name used for disk CRD
	DiskResourceShort = "disk"
	// DiskResourceName is the name of the disk resource
	DiskResourceName = DiskResourcePlural + "." + GroupName

	// BlockDeviceResourceKind is the kind of block device CRD
	BlockDeviceResourceKind = "BlockDevice"
	// BlockDeviceResourceListKind is the list kind for block device
	BlockDeviceResourceListKind = "BlockDeviceList"
	// BlockDeviceResourcePlural is the plural form used for block device
	BlockDeviceResourcePlural = "blockdevices"
	// BlockDeviceResourceShort is the short name used for block device CRD
	BlockDeviceResourceShort = "bd"
	// BlockDeviceResourceName is the name of the block device resource
	BlockDeviceResourceName = BlockDeviceResourcePlural + "." + GroupName

	// BlockDeviceClaimResourceKind is the kind of block device claim CRD
	BlockDeviceClaimResourceKind = "BlockDeviceClaim"
	// BlockDeviceClaimResourceListKind is the list kind for block device claim
	BlockDeviceClaimResourceListKind = "BlockDeviceClaimList"
	// BlockDeviceClaimResourcePlural is the plural form used for block device claim
	BlockDeviceClaimResourcePlural = "blockdeviceclaims"
	// BlockDeviceClaimResourceShort is the short name used for block device claim CRD
	BlockDeviceClaimResourceShort = "bdc"
	// BlockDeviceClaimResourceName is the name of the block device claim resource
	BlockDeviceClaimResourceName = BlockDeviceClaimResourcePlural + "." + GroupName
)
