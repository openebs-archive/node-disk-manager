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
  #include <stdlib.h>
*/
import "C"

import "unsafe"

const (
	UDEV_SUBSYSTEM = "block"           // udev to filter this device type
	UDEV_SYSTEM    = "disk"            // used to filter devices other than disk which udev tracks (eg. CD ROM)
	UDEV_PATH      = "DEVPATH"         // udev attribute to get device path
	UDEV_WWN       = "ID_WWN"          // udev attribute to get device WWN number
	UDEV_SERIAL    = "ID_SERIAL_SHORT" // udev attribute to get device serial number
	UDEV_MODEL     = "ID_MODEL"        // udev attribute to get device model number
	UDEV_VENDOR    = "ID_VENDOR"       // udev attribute to get device vendor details
	UDEV_TYPE      = "ID_TYPE"         // udev attribute to get device type
)

func freeCharPtr(s *C.char) {
	C.free(unsafe.Pointer(s))
}

// Device wraps a libudev device object
type Udevice struct {
	ptr *C.struct_udev_device
}

// newDevice is a private helper function and returns a pointer to a new device.
// The device is also added t the devices map in the udev context.
// The agrument ptr is a pointer to the underlying C udev_device structure.
// The function returns nil if the pointer passed is NULL.
func newUdevice(ptr *C.struct_udev_device) (d *Udevice) {
	// If passed a NULL pointer, return nil
	if ptr == nil {
		return nil
	}
	// Create a new device object
	d = &Udevice{
		ptr: ptr,
	}
	// Return the device object
	return
}

// PropertyValue retrieves the value of a device property
func (d *Udevice) PropertyValue(key string) string {
	k := C.CString(key)
	defer freeCharPtr(k)
	return C.GoString(C.udev_device_get_property_value(d.ptr, k))
}

// SysattrValue retrieves the content of a sys attribute file, and returns an empty string if there is no sys attribute value.
// The retrieved value is cached in the device.
// Repeated calls will return the same value and not open the attribute again.
func (d *Udevice) SysattrValue(sysattr string) string {
	s := C.CString(sysattr)
	defer freeCharPtr(s)
	return C.GoString(C.udev_device_get_sysattr_value(d.ptr, s))
}

// Devtype returns the devtype string of the udev device.
func (d *Udevice) Devtype() string {
	return C.GoString(C.udev_device_get_devtype(d.ptr))
}

// Devnode returns the device node file name belonging to the udev device.
// The path is an absolute path, and starts with the device directory.
func (d *Udevice) Devnode() string {
	return C.GoString(C.udev_device_get_devnode(d.ptr))
}

// ListDevices lists all the block devices currently attached to the system
// and returns them into array of pointers of C udev_device structure. The
// caller can traverse it and query all(available) the disk properties using
// this C udev_device structure.
func ListDevices() (m []*Udevice) {
	udev := C.udev_new()
	if udev == nil {
		return nil
	}

	defer C.udev_unref(udev)

	enumerate := C.udev_enumerate_new(udev)
	if enumerate == nil {
		return nil
	}

	defer C.udev_enumerate_unref(enumerate)

	subsystem := C.CString(UDEV_SUBSYSTEM)
	if subsystem == nil {
		return nil
	}

	defer freeCharPtr(subsystem)

	err := C.udev_enumerate_add_match_subsystem(enumerate, subsystem)
	if err < 0 {
		return nil
	}

	C.udev_enumerate_scan_devices(enumerate)
	if err < 0 {
		return nil
	}

	m = make([]*Udevice, 0)

	for l := C.udev_enumerate_get_list_entry(enumerate); l != nil; l = C.udev_list_entry_get_next(l) {
		s := C.udev_list_entry_get_name(l)
		m = append(m, newUdevice(C.udev_device_new_from_syspath(udev, s)))
	}

	return m
}
