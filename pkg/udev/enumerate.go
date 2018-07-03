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
  #include <libudev.h>
*/
import "C"
import "errors"

// UdevEnumerate wraps a libudev udev_enumerate object
type UdevEnumerate struct {
	ueptr *C.struct_udev_enumerate
}

// UdevEnumerateAddMatchSubsystem adds filter in UdeviceMon struct.
// This filter is efficiently executed inside kernel, and libudev
// subscribers will usually not be woken up for devices which do not match.
func (ue *UdevEnumerate) UdevEnumerateAddMatchSubsystem(subSystem string) error {
	subsystem := C.CString(subSystem)
	defer freeCharPtr(subsystem)
	ret := C.udev_enumerate_add_match_subsystem(ue.ueptr, subsystem)
	if ret < 0 {
		return errors.New("unable to apply sybsystem filter")
	}
	return nil
}

// UdevEnumerateScanDevices scan devices in system
func (ue *UdevEnumerate) UdevEnumerateScanDevices() error {
	ret := C.udev_enumerate_scan_devices(ue.ueptr)
	if ret < 0 {
		return errors.New("unable to scan device list")
	}
	return nil
}

// UdevEnumerateGetListEntry returns UdevListEntry struct from which wecan get device.
func (ue *UdevEnumerate) UdevEnumerateGetListEntry() *UdevListEntry {
	return newUdevListEntry(C.udev_enumerate_get_list_entry(ue.ueptr))
}

// UnrefUdevEnumerate frees udev_enumerate structure.
func (ue *UdevEnumerate) UnrefUdevEnumerate() {
	C.udev_enumerate_unref(ue.ueptr)
}
