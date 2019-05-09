package mount

const (
	hostMountFilePath = "/host/proc/1/mounts" // hostMountFilePath is the file path mounted inside container
)

// Identifier is an identifer for the mount probe. It will be a devpath like
// /dev/sda, /dev/sda1 etc
type Identifier struct {
	DevPath string
}

// DeviceAttr is the struct used for returning all available mount related information like
// mount point, filesystem type etc.
// It helps to find mountpoint of a partition/block
type DeviceAttr struct {
	DevPath    string // DevPath of the device/block
	MountPoint string // MountPoint of the the device/block
	FileSystem string // FileSystem in the device that is mounted
}

// DeviceBasicMountInfo gives the mount attributes of a device that is attached. The mount
// attributes include the filesystem type, mountpoint, device path etc. These mount attributes
// are fetched by parsing a mounts file (/proc/1/mounts) and getting the relevant data. If the
// device is not mounted, then the function will return an error.
func (I *Identifier) DeviceBasicMountInfo() (DeviceAttr, error) {
	mountUtil := NewMountUtil(hostMountFilePath, I.DevPath, "")
	mountAttr, err := mountUtil.getDeviceMountAttr(mountUtil.getMountName)
	return mountAttr, err
}
