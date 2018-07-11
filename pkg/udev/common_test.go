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

import (
	"testing"

	"github.com/openebs/node-disk-manager/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestGetSyspath(t *testing.T) {
	diskDetails, err := MockDiskDetails()
	if err != nil {
		t.Fatal(err)
	}
	newUdev, err := NewUdev()
	if err != nil {
		t.Fatal(err)
	}
	defer newUdev.UnrefUdev()
	device, err := newUdev.NewDeviceFromSysPath(diskDetails.SysPath)
	if err != nil {
		t.Fatal(err)
	}
	defer device.UdevDeviceUnref()
	assert.Equal(t, diskDetails.SysPath, device.GetSyspath())
}

func TestDiskInfoFromLibudev(t *testing.T) {
	diskDetails, err := MockDiskDetails()
	if err != nil {
		t.Fatal(err)
	}
	newUdev, err := NewUdev()
	if err != nil {
		t.Fatal(err)
	}
	defer newUdev.UnrefUdev()
	device, err := newUdev.NewDeviceFromSysPath(diskDetails.SysPath)
	if err != nil {
		t.Fatal(err)
	}
	defer device.UdevDeviceUnref()
	expectedDiskDetails := UdevDiskDetails{
		Model:  diskDetails.Model,
		Serial: diskDetails.Serial,
		Vendor: diskDetails.Vendor,
		Path:   diskDetails.DevNode,
		Size:   diskDetails.Capacity,
	}
	assert.Equal(t, expectedDiskDetails, device.DiskInfoFromLibudev())
}

func TestGetSize(t *testing.T) {
	diskDetails, err := MockDiskDetails()
	if err != nil {
		t.Fatal(err)
	}
	newUdev, err := NewUdev()
	if err != nil {
		t.Fatal(err)
	}
	defer newUdev.UnrefUdev()
	device, err := newUdev.NewDeviceFromSysPath(diskDetails.SysPath)
	if err != nil {
		t.Fatal(err)
	}
	defer device.UdevDeviceUnref()
	actualCapacity, err := device.getSize()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, diskDetails.Capacity, actualCapacity)
}

func TestIsDisk(t *testing.T) {
	udev, err := NewUdev()
	if err != nil {
		t.Fatal(err)
	}
	defer udev.UnrefUdev()
	udevEnumerate, err := udev.NewUdevEnumerate()
	if err != nil {
		t.Fatal(err)
	}
	defer udevEnumerate.UnrefUdevEnumerate()

	err = udevEnumerate.AddSubsystemFilter(UDEV_SUBSYSTEM)
	if err != nil {
		t.Fatal(err)
	}
	err = udevEnumerate.ScanDevices()
	if err != nil {
		t.Fatal(err)
	}
	for l := udevEnumerate.ListEntry(); l != nil; l = l.GetNextEntry() {
		s := l.GetName()
		newUdevice, err := udev.NewDeviceFromSysPath(s)
		if err != nil {
			t.Fatal(err)
		}
		if newUdevice.GetDevtype() == UDEV_SYSTEM && newUdevice.GetPropertyValue(UDEV_TYPE) == UDEV_SYSTEM {
			assert.Equal(t, true, newUdevice.IsDisk())
		} else {
			assert.Equal(t, false, newUdevice.IsDisk())
		}
		newUdevice.UdevDeviceUnref()
	}
}

func TestGetUid(t *testing.T) {
	diskDetails, err := MockDiskDetails()
	if err != nil {
		t.Fatal(err)
	}
	newUdev, err := NewUdev()
	if err != nil {
		t.Fatal(err)
	}
	defer newUdev.UnrefUdev()
	device, err := newUdev.NewDeviceFromSysPath(diskDetails.SysPath)
	if err != nil {
		t.Fatal(err)
	}
	defer device.UdevDeviceUnref()
	expectedUid := NDMPrefix + util.Hash(diskDetails.Wwn+diskDetails.Model+diskDetails.Serial+diskDetails.Vendor)
	assert.Equal(t, expectedUid, device.GetUid())
}
