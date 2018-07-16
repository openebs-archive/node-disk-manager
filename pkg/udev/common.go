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
	"strings"
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
	UDEV_DEVLINKS       = "DEVLINKS"        //udev attrinute contain devlinks of a disk
	BY_ID_LINK          = "by-id"           // by-path devlink contains this string
	BY_PATH_LINK        = "by-path"         // by-path devlink contains this string
)

// UdevDiskDetails struct contain different attribute of disk.
type UdevDiskDetails struct {
	Model          string   // Model is Model of disk.
	Serial         string   // Serial is Serial of a disk.
	Vendor         string   // Vendor is Vendor of a disk.
	Path           string   // Path is Path of a disk.
	Size           uint64   // Size is capacity of disk
	ByIdDevLinks   []string // ByIdDevLinks contains by-id devlinks
	ByPathDevLinks []string // ByPathDevLinks contains by-path devlinks
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
	devLinks := device.GetDevLinks()
	diskDetails := UdevDiskDetails{
		Model:          device.GetPropertyValue(UDEV_MODEL),
		Serial:         device.GetPropertyValue(UDEV_SERIAL),
		Vendor:         device.GetPropertyValue(UDEV_VENDOR),
		Path:           device.GetPropertyValue(UDEV_DEVNAME),
		Size:           size,
		ByIdDevLinks:   devLinks[BY_ID_LINK],
		ByPathDevLinks: devLinks[BY_PATH_LINK],
	}
	return diskDetails
}

// GetUid returns unique id for the disk block device
func (device *UdevDevice) GetUid() string {
	return NDMPrefix +
		util.Hash(device.GetPropertyValue(UDEV_WWN)+
			device.GetPropertyValue(UDEV_MODEL)+
			device.GetPropertyValue(UDEV_SERIAL)+
			device.GetPropertyValue(UDEV_VENDOR))
}

// IsDisk returns true if device is a disk
func (device *UdevDevice) IsDisk() bool {
	return device.GetDevtype() == UDEV_SYSTEM && device.GetPropertyValue(UDEV_TYPE) == UDEV_SYSTEM
}

// GetSyspath returns syspath of a disk using syspath we can fell details
// in diskInfo struct using udev probe
func (device *UdevDevice) GetSyspath() string {
	major := device.GetPropertyValue(UDEV_MAJOR)
	minor := device.GetPropertyValue(UDEV_MINOR)
	syspath := UDEV_SYSPATH_PREFIX + major + ":" + minor
	return syspath
}

// getSize returns size of a disk.
func (device *UdevDevice) getSize() (uint64, error) {
	var sector []byte
	var sec int64
	n, err := strconv.ParseInt(device.GetSysattrValue("size"), 10, 64)
	if err != nil {
		return 0, err
	}
	// should we use disk smart queries to get the sector size?
	fname := "/sys" + device.GetPropertyValue(UDEV_PATH) + "/queue/hw_sector_size"
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

// GetDevLinks returns syspath of a disk using syspath we can fell details
// in diskInfo struct using udev probe
func (device *UdevDevice) GetDevLinks() map[string][]string {
	devLinkMap := make(map[string][]string)
	byIdLink := make([]string, 0)
	byPathLink := make([]string, 0)
	for _, link := range strings.Split(device.GetPropertyValue(UDEV_DEVLINKS), " ") {
		parts := strings.Split(link, "/")
		if util.Contains(parts, BY_ID_LINK) {
			byIdLink = append(byIdLink, link)
		}
		if util.Contains(parts, BY_PATH_LINK) {
			byPathLink = append(byPathLink, link)
		}
	}
	devLinkMap[BY_ID_LINK] = byIdLink
	devLinkMap[BY_PATH_LINK] = byPathLink
	return devLinkMap
}
