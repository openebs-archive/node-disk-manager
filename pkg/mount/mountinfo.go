package mount

const (
	hostMountFilePath = "/host/proc/1/mounts" // hostMountFilePath is the file path mounted inside container
)

type Identifier struct {
	DevPath string
}

// MountAttr contains mount point and device mount attributes
// It helps to find mountpoint of a partition/block
type MountAttr struct {
	MountPoint string // MountPoint of the the device/block
	FileSystem string // FileSystem in the device that is mounted
}

func (I *Identifier) DeviceBasicMountInfo() (MountAttr, error) {
	mountUtil := newMountUtil(hostMountFilePath, I.DevPath)
	mountAttr, err := mountUtil.getMountAttr()
	return mountAttr, err
}
