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

package mount

const (
	hostMountFilePath = "/host/proc/1/mounts" // hostMountFilePath is the file path mounted inside container
)

// Identifier is an identifer for the mount probe. It will be a devpath like
// /dev/sda, /dev/sda1 etc
type Identifier struct {
	DevPath string
}

// DeviceMountAttr is the struct used for returning all available mount related information like
// mount point, filesystem type etc.
// It helps to find mountpoint of a partition/block
type DeviceMountAttr struct {
	DevPath    string // DevPath of the device/block
	MountPoint string // MountPoint of the the device/block
	FileSystem string // FileSystem in the device that is mounted
}

// DeviceBasicMountInfo gives the mount attributes of a device that is attached. The mount
// attributes include the filesystem type, mountpoint, device path etc. These mount attributes
// are fetched by parsing a mounts file (/proc/1/mounts) and getting the relevant data. If the
// device is not mounted, then the function will return an error.
func (I *Identifier) DeviceBasicMountInfo() (DeviceMountAttr, error) {
	mountUtil := NewMountUtil(hostMountFilePath, I.DevPath, "")
	mountAttr, err := mountUtil.getDeviceMountAttr(mountUtil.getMountName)
	return mountAttr, err
}
