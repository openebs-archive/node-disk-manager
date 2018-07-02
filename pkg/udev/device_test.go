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

	"github.com/stretchr/testify/assert"
)

func TestUdevDeviceGetPropertyValue(t *testing.T) {
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
	deviceType := device.UdevDeviceGetPropertyValue(UDEV_TYPE)
	assert.Equal(t, diskDetails.DevType, deviceType)
}

func TestUdevDeviceGetSysattrValue(t *testing.T) {
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
	expectedSize := device.UdevDeviceGetSysattrValue("size")
	assert.Equal(t, expectedSize, diskDetails.Size)
}

func TestUdevDeviceGetDevtype(t *testing.T) {
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
	assert.Equal(t, diskDetails.DevType, device.UdevDeviceGetDevtype())
}

func TestUdevDeviceGetDevnode(t *testing.T) {
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
	assert.Equal(t, diskDetails.DevNode, device.UdevDeviceGetDevnode())
}
