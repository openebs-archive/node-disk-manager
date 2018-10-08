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
	"os"
	"strings"
	"unsafe"

	"github.com/openebs/node-disk-manager/pkg/util"
)

const (
	NDMPrefix           = "disk-"           // NDMPrefix used as disk's uuid prefix
	UDEV_SUBSYSTEM      = "block"           // udev to filter this device type
	UDEV_SYSTEM         = "disk"            // used to filter devices other than disk which udev tracks (eg. CD ROM)
	UDEV_PATH           = "DEVPATH"         // udev attribute to get device path
	UDEV_WWN            = "ID_WWN"          // udev attribute to get device WWN number
	UDEV_SERIAL         = "ID_SERIAL_SHORT" // udev attribute to get device serial number
	UDEV_SERIAL_FULL    = "ID_SERIAL"       // udev attribute to get - separated vendor, model, serial
	UDEV_BUS            = "ID_BUS"          // udev attribute to get bus name
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
	UDEV_DEVNAME        = "DEVNAME"         // udev attribute contain disk name given by kernel
	UDEV_DEVLINKS       = "DEVLINKS"        // udev attribute contain devlinks of a disk
	BY_ID_LINK          = "by-id"           // by-path devlink contains this string
	BY_PATH_LINK        = "by-path"         // by-path devlink contains this string
	LINK_ID_INDEX       = 4                 // this is used to get link index from dev link
	UDEV_SIZE           = "SIZE"            // udev attribute to get number of block/sector in disk
)

// UdevDiskDetails struct contain different attribute of disk.
type UdevDiskDetails struct {
	Model          string   // Model is Model of disk.
	Serial         string   // Serial is Serial of a disk.
	Vendor         string   // Vendor is Vendor of a disk.
	Path           string   // Path is Path of a disk.
	ByIdDevLinks   []string // ByIdDevLinks contains by-id devlinks
	ByPathDevLinks []string // ByPathDevLinks contains by-path devlinks
	NoOfBlocks     string   // NoOfBlocks contains number of block/sector in a disk.
}

// freeCharPtr frees c pointer
func freeCharPtr(s *C.char) {
	C.free(unsafe.Pointer(s))
}

//DiskInfoFromLibudev returns disk attribute extracted using libudev apicalls.
func (device *UdevDevice) DiskInfoFromLibudev() UdevDiskDetails {
	devLinks := device.GetDevLinks()
	diskDetails := UdevDiskDetails{
		Model:          device.GetPropertyValue(UDEV_MODEL),
		Serial:         device.GetPropertyValue(UDEV_SERIAL),
		Vendor:         device.GetPropertyValue(UDEV_VENDOR),
		Path:           device.GetPropertyValue(UDEV_DEVNAME),
		ByIdDevLinks:   devLinks[BY_ID_LINK],
		ByPathDevLinks: devLinks[BY_PATH_LINK],
		NoOfBlocks:     device.GetSysattrValue(UDEV_SIZE),
	}
	return diskDetails
}

// GetUid returns unique id for the disk block device
func (device *UdevDevice) GetUid() string {
	uid := device.GetPropertyValue(UDEV_WWN) +
		device.GetPropertyValue(UDEV_MODEL) +
		device.GetPropertyValue(UDEV_SERIAL) +
		device.GetPropertyValue(UDEV_VENDOR)

	idtype := device.GetPropertyValue(UDEV_TYPE)

	model := device.GetPropertyValue(UDEV_MODEL)

	// Virtual disks either have no attributes or they all have
	// the same attributes. Adding hostname in uid so that disks from different
	// nodes can be differentiated. Also, putting devpath in uid so that disks
	// from the same node also can be differentiated.
	// 	On Gke, we have the ID_TYPE property, but still disks will have
	// same attributes. We have to put a special check to handle it and process
	// it like a Virtual disk.
	localDiskModels := make([]string, 0)
	localDiskModels = append(localDiskModels, "EphemeralDisk")
	localDiskModels = append(localDiskModels, "Virtual_disk")
	if len(idtype) == 0 || util.Contains(localDiskModels, model) {
		// as hostNetwork is true, os.Hostname will give you the node's Hostname
		host, _ := os.Hostname()
		uid += host + device.GetPropertyValue(UDEV_DEVNAME)
	}

	return NDMPrefix + util.Hash(uid)
}

// IsDisk returns true if device is a disk
func (device *UdevDevice) IsDisk() bool {
	return device.GetDevtype() == UDEV_SYSTEM
}

// GetSyspath returns syspath of a disk using syspath we can fell details
// in diskInfo struct using udev probe
func (device *UdevDevice) GetSyspath() string {
	major := device.GetPropertyValue(UDEV_MAJOR)
	minor := device.GetPropertyValue(UDEV_MINOR)
	syspath := UDEV_SYSPATH_PREFIX + major + ":" + minor
	return syspath
}

// GetDevLinks returns syspath of a disk using syspath we can fell details
// in diskInfo struct using udev probe
func (device *UdevDevice) GetDevLinks() map[string][]string {
	devLinkMap := make(map[string][]string)
	byIdLink := make([]string, 0)
	byPathLink := make([]string, 0)
	for _, link := range strings.Split(device.GetPropertyValue(UDEV_DEVLINKS), " ") {
		/*
			devlink is like - /dev/disk/by-id/scsi-0Google_PersistentDisk_demo-disk
			parts = ["", "dev", "disk", "by-id", "scsi-0Google_PersistentDisk_demo-disk"]
			parts[4] contains link index like model or wwn or sysPath (wwn-0x5000c5009e3a8d2b) (ata-ST500LM021-1KJ152_W6HFGR)
		*/
		parts := strings.Split(link, "/")
		if util.Contains(parts, BY_ID_LINK) {
			/*
				A default by-id link is observed to be created for all types of disks (physical, virtual and cloud).
				This link has the format - bus, vendor, model, serial - all appended in the same order. Keeping this
				link as the first element of array for consistency purposes.
			*/
			if strings.HasPrefix(parts[LINK_ID_INDEX], device.GetPropertyValue(UDEV_BUS)) && strings.HasSuffix(parts[LINK_ID_INDEX], device.GetPropertyValue(UDEV_SERIAL_FULL)) {
				byIdLink = append([]string{link}, byIdLink...)
			} else {
				byIdLink = append(byIdLink, link)
			}
		}
		if util.Contains(parts, BY_PATH_LINK) {
			byPathLink = append(byPathLink, link)
		}
	}
	devLinkMap[BY_ID_LINK] = byIdLink
	devLinkMap[BY_PATH_LINK] = byPathLink
	return devLinkMap
}
