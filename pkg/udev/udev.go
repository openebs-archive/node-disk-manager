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
import (
	"errors"
)

// Udev wraps a libudev udev struct
type Udev struct {
	uptr *C.struct_udev
}

// newUdev is a helper function and returns a pointer to a new udev
func newUdev(ptr *C.struct_udev) (*Udev, error) {
	if ptr == nil {
		return nil, errors.New("unable to create Udevice object for null struct struct_udev")
	}
	u := &Udev{
		uptr: ptr,
	}
	return u, nil
}

// NewUdev is a function and which returns a pointer to a new Udev
func NewUdev() (*Udev, error) {
	udev := C.udev_new()
	return newUdev(udev)
}

// NewUdevEnumerate returns a pointer to a Udevenumerate
func (u *Udev) NewUdevEnumerate() (*UdevEnumerate, error) {
	if u == nil {
		return nil, errors.New("unable to create Udevenumerate object for for null struct struct_udev_enumerate")
	}
	enumerateptr := C.udev_enumerate_new(u.uptr)
	if enumerateptr == nil {
		return nil, errors.New("unable to create Udevenumerate object for for null struct struct_udev_enumerate")
	}
	ue := &UdevEnumerate{
		ueptr: enumerateptr,
	}
	return ue, nil
}

// UnrefUdev frees udev structure.
func (u *Udev) UnrefUdev() {
	C.udev_unref(u.uptr)
}

// NewDeviceFromSysPath identify the block device currently attached to the system
// at given sysPath and returns that pointer of C udev_device structure. The
// caller can query all(available) the disk properties using returned C udev_device structure.
func (u *Udev) NewDeviceFromSysPath(sysPath string) (*UdevDevice, error) {
	syspath := C.CString(sysPath)
	defer freeCharPtr(syspath)
	return newUdevDevice(C.udev_device_new_from_syspath(u.uptr, syspath))
}

// NewDeviceFromNetlink use newUdeviceMonitor() and use returns UdevMonitor pointer in success
// The function returns nil on failure it can monitor udev or kernel as source.
func (u *Udev) NewDeviceFromNetlink(source string) (*UdevMonitor, error) {
	monitorSources := C.CString(source)
	defer freeCharPtr(monitorSources)
	return newUdeviceMonitor(C.udev_monitor_new_from_netlink(u.uptr, monitorSources))
}
