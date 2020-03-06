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
	"path/filepath"
	"testing"

	"github.com/openebs/node-disk-manager/pkg/util"
	"github.com/stretchr/testify/assert"
)

/*
	In this test case we create os disk's UdevDevice object.
	get the syspath of that disk and compare with mock details
*/
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
	tests := map[string]struct {
		actualSyspath   string
		expectedSysPath string
	}{
		"compare syspath of os disk": {actualSyspath: device.GetSyspath(), expectedSysPath: diskDetails.SysPath},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.actualSyspath, test.expectedSysPath)
		})
	}
}

/*
	In this test case we create os disk's UdevDevice object.
	and get the data what udevProbe can fill and compare it
	with mock UdevDiskDetails struct
*/
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
		Model:              diskDetails.Model,
		Serial:             diskDetails.Serial,
		Vendor:             diskDetails.Vendor,
		WWN:                diskDetails.Wwn,
		DiskType:           diskDetails.DevType,
		Path:               diskDetails.DevNode,
		ByIdDevLinks:       diskDetails.ByIdDevLinks,
		ByPathDevLinks:     diskDetails.ByPathDevLinks,
		PartitionTableType: diskDetails.PartTableType,
	}
	tests := map[string]struct {
		actualDetails   UdevDiskDetails
		expectedDetails UdevDiskDetails
	}{
		"check for details which udev probe can fill": {actualDetails: device.DiskInfoFromLibudev(), expectedDetails: expectedDiskDetails},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDetails, test.actualDetails)
		})
	}
}

/*
	In this test case we create os disk's UdevDevice object.
	as it is a disk it should return true
*/
func TestIsDisk(t *testing.T) {
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
	tests := map[string]struct {
		actual   bool
		expected bool
	}{
		"check if os disk is disk or not": {actual: device.IsDisk(), expected: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.actual, test.expected)
		})
	}
}

/*
	In this test case we create os disk's UdevDevice object from that we
	get uuid of os disk. Then we compare it with uuid generation logic.
*/
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
	expectedUid := NDMBlockDevicePrefix + util.Hash(diskDetails.Wwn+diskDetails.Model+diskDetails.Serial+diskDetails.Vendor)
	tests := map[string]struct {
		actualUuid   string
		expectedUuid string
	}{
		"check for os disk uuid": {actualUuid: device.GetUid(), expectedUuid: expectedUid},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.actualUuid, test.expectedUuid)
		})
	}
}

/*
	Get devlinks and devnode (/dev/sda or /dev/sdb ... ) of os
	disk, read file path of those devlinks and match with devnode.
	Each devlink should be valid and file path of those links
	equal with devnode of os disk.
*/
func TestGetDevLinks(t *testing.T) {
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
	osDiskPath := diskDetails.DevNode
	var expectedError error
	for name, links := range device.GetDevLinks() {
		for _, link := range links {
			t.Run(name, func(t *testing.T) {
				path, err := filepath.EvalSymlinks(link)
				assert.Equal(t, expectedError, err)
				assert.Equal(t, osDiskPath, path)
			})
		}
	}
}
