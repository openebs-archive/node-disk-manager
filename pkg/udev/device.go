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

// UdevDevice wraps a libudev device object
type UdevDevice struct {
	udptr *C.struct_udev_device
}

// newUdevDevice is a private helper function and returns a pointer to a new device
// it returns nil if the pointer passed is NULL.
func newUdevDevice(ptr *C.struct_udev_device) (*UdevDevice, error) {
	if ptr == nil {
		return nil, errors.New("unable to create Udevice object for for null struct struct_udev_device")
	}
	ud := &UdevDevice{
		udptr: ptr,
	}
	return ud, nil
}

// UdevDeviceGetPropertyValue retrieves the value of a device property
func (ud *UdevDevice) UdevDeviceGetPropertyValue(key string) string {
	k := C.CString(key)
	defer freeCharPtr(k)
	return C.GoString(C.udev_device_get_property_value(ud.udptr, k))
}

// UdevDeviceGetSysattrValue retrieves the content of a sys attribute file
// returns an empty string if there is no sys attribute value.
// The retrieved value is cached in the device.
// Repeated calls will return the same value and not open the attribute again.
func (ud *UdevDevice) UdevDeviceGetSysattrValue(sysattr string) string {
	k := C.CString(sysattr)
	defer freeCharPtr(k)
	return C.GoString(C.udev_device_get_sysattr_value(ud.udptr, k))
}

// UdevDeviceGetDevtype returns the devtype string of the udev device.
func (ud *UdevDevice) UdevDeviceGetDevtype() string {
	return C.GoString(C.udev_device_get_devtype(ud.udptr))
}

// UdevDeviceGetDevnode returns the device node file name belonging to the udev device.
// The path is an absolute path, and starts with the device directory.
func (ud *UdevDevice) UdevDeviceGetDevnode() string {
	return C.GoString(C.udev_device_get_devnode(ud.udptr))
}

// UdevDeviceUnref frees udev_device structure.
func (ud *UdevDevice) UdevDeviceUnref() {
	C.udev_device_unref(ud.udptr)
}
