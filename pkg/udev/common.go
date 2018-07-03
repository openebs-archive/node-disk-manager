// +build linux,cgo

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

package udev

/*
  #cgo LDFLAGS: -ludev
  #include <stdlib.h>
*/
import "C"
import (
	"io/ioutil"
	"strconv"
	"unsafe"

	"github.com/openebs/node-disk-manager/pkg/util"
)

const (
	NDMPrefix           = "disk-"
	UDEV_SUBSYSTEM      = "block"           // udev to filter this device type
	UDEV_SYSTEM         = "disk"            // used to filter devices other than disk which udev tracks (eg. CD ROM)
	UDEV_PATH           = "DEVPATH"         // udev attribute to get device path
	UDEV_WWN            = "ID_WWN"          // udev attribute to get device WWN number
	UDEV_SERIAL         = "ID_SERIAL_SHORT" // udev attribute to get device serial number
	UDEV_MODEL          = "ID_MODEL"        // udev attribute to get device model number
	UDEV_VENDOR         = "ID_VENDOR"       // udev attribute to get device vendor details
	UDEV_TYPE           = "ID_TYPE"         // udev attribute to get device type
	UDEV_MAJOR          = "MAJOR"           // udev attribute to get device major no
	UDEV_MINOR          = "MINOR"           // udev attribute to get device minor no
	UDEV_UUID           = "UDEV_UUID"       // ndm attribute to get device uuid
	UDEV_SYSPATH        = "UDEV_SYSPATH"    // udev attribute to get device syspath
	UDEV_ACTION         = "UDEV_ACTION"     // udev attribute to get monitor device action
	UDEV_ACTION_ADD     = "add"             // udev attribute constant for add action
	UDEV_ACTION_REMOVE  = "remove"          // udev attribute constant for remove action
	UDEV_DEVTYPE        = "DEVTYPE"         // udev attribute to get device device type ie - disk or part
	UDEV_SOURCE         = "udev"            // udev source constant
	UDEV_SYSPATH_PREFIX = "/sys/dev/block/" // udev syspath prefix
	UDEV_DEVNAME        = "DEVNAME"         // udev attrinute contain disk name given by kernel
)

// UdevDiskDetails struct contain different attribute of disk.
type UdevDiskDetails struct {
	Model  string // Model is Model of disk.
	Serial string // Serial is Serial of a disk.
	Vendor string // Vendor is Vendor of a disk.
	Path   string // Path is Path of a disk.
	Size   uint64 // Size is capacity of disk
}

// freeCharPtr frees c pointer
func freeCharPtr(s *C.char) {
	C.free(unsafe.Pointer(s))
}

//DiskInfoFromLibudev returns disk attribute extracted using libudev apicalls.
func (device *UdevDevice) DiskInfoFromLibudev() UdevDiskDetails {
	size, err := device.getSize()
	if err != nil {
	}
	diskDetails := UdevDiskDetails{
		Model:  device.UdevDeviceGetPropertyValue(UDEV_MODEL),
		Serial: device.UdevDeviceGetPropertyValue(UDEV_SERIAL),
		Vendor: device.UdevDeviceGetPropertyValue(UDEV_VENDOR),
		Path:   device.UdevDeviceGetPropertyValue(UDEV_DEVNAME),
		Size:   size,
	}
	return diskDetails
}

// GetUid returns unique id for the disk block device
func (device *UdevDevice) GetUid() string {
	return NDMPrefix +
		util.Hash(device.UdevDeviceGetPropertyValue(UDEV_WWN)+
			device.UdevDeviceGetPropertyValue(UDEV_MODEL)+
			device.UdevDeviceGetPropertyValue(UDEV_SERIAL)+
			device.UdevDeviceGetPropertyValue(UDEV_VENDOR))
}

// IsDisk returns true if device is a disk
func (device *UdevDevice) IsDisk() bool {
	return device.UdevDeviceGetDevtype() == UDEV_SYSTEM && device.UdevDeviceGetPropertyValue(UDEV_TYPE) == UDEV_SYSTEM
}

// GetSyspath returns syspath of a disk using syspath we can fell details
// in diskInfo struct using udev probe
func (device *UdevDevice) GetSyspath() string {
	major := device.UdevDeviceGetPropertyValue(UDEV_MAJOR)
	minor := device.UdevDeviceGetPropertyValue(UDEV_MINOR)
	syspath := UDEV_SYSPATH_PREFIX + major + ":" + minor
	return syspath
}

// getSize returns size of a disk.
func (device *UdevDevice) getSize() (uint64, error) {
	var sector []byte
	var sec int64
	n, err := strconv.ParseInt(device.UdevDeviceGetSysattrValue("size"), 10, 64)
	if err != nil {
		return 0, err
	}
	// should we use disk smart queries to get the sector size?
	fname := "/sys" + device.UdevDeviceGetPropertyValue(UDEV_PATH) + "/queue/hw_sector_size"
	sector, err = ioutil.ReadFile(fname)
	if err != nil {
		return 0, err
	}
	sec, err = strconv.ParseInt(string(sector[:len(sector)-1]), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint64(n * sec), nil
}
