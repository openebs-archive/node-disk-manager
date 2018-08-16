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

package controller

import (
	"testing"

	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs.io/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
In this unit test case we will create one diskInfo struct with path attribute
and get path value from getPath() and compare them
*/
func TestGetPath(t *testing.T) {
	diskPath := "sample-disk-path"
	fakeDiskInfo := NewDiskInfo()
	fakeDiskInfo.Path = diskPath
	tests := map[string]struct {
		actualPath   string
		expectedPath string
	}{
		"create one diskinfo object and get disk path from it": {actualPath: fakeDiskInfo.getPath(), expectedPath: diskPath},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.actualPath, test.expectedPath)
		})
	}
}

/*
In this unit test case we will create one diskInfo struct with mock Capacity details
and get Capacity value from getDiskCapacity() and compare them
*/
func TestGetDiskCapacity(t *testing.T) {
	diskCapacity := uint64(123)
	// add mock fields in diskInfo struct
	fakeCapacity := apis.DiskCapacity{}
	fakeCapacity.Storage = diskCapacity

	// Creating one diskDetails struct using mock value
	fakeDiskInfo := NewDiskInfo()
	fakeDiskInfo.Capacity = diskCapacity

	tests := map[string]struct {
		actualCapacity   apis.DiskCapacity
		expectedCapacity apis.DiskCapacity
	}{
		"create one diskinfo object and get capacity from it": {actualCapacity: fakeDiskInfo.getDiskCapacity(), expectedCapacity: fakeCapacity},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.actualCapacity, test.expectedCapacity)
		})
	}
}

/*
In this unit test case we will create one diskInfo struct with mock details
and get DiskDetails value from getDiskDetails() and compare them
*/
func TestGetDiskDetails(t *testing.T) {
	fakeModel := "disk-model-no"
	fakeSerial := "disk-serial-no"
	fakeVendor := "disk-vendor"
	// add mock fields in diskInfo struct
	fakeDiskInfo := NewDiskInfo()
	fakeDiskInfo.Model = fakeModel
	fakeDiskInfo.Serial = fakeSerial
	fakeDiskInfo.Vendor = fakeVendor

	// Creating one diskDetails struct using mock value
	fakeDiskDetails := apis.DiskDetails{}
	fakeDiskDetails.Model = fakeModel
	fakeDiskDetails.Serial = fakeSerial
	fakeDiskDetails.Vendor = fakeVendor

	tests := map[string]struct {
		actualDetails   apis.DiskDetails
		expectedDetails apis.DiskDetails
	}{
		"create one fakeDiskDetails object and compare it": {actualDetails: fakeDiskInfo.getDiskDetails(), expectedDetails: fakeDiskDetails},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.actualDetails, test.expectedDetails)
		})
	}
}

/*
In this unit test case we will create one diskInfo struct with mock details
and get TypeMeta value from getTypeMeta() and compare them
*/
func TestGetTypeMeta(t *testing.T) {
	fakeDiskInfo := NewDiskInfo()
	fakeTypeMeta := metav1.TypeMeta{
		Kind:       NDMKind,
		APIVersion: NDMVersion,
	}

	tests := map[string]struct {
		actualTypeMeta   metav1.TypeMeta
		expectedTypeMeta metav1.TypeMeta
	}{
		"create one mock typeMeta object and compare it": {actualTypeMeta: fakeDiskInfo.getTypeMeta(), expectedTypeMeta: fakeTypeMeta},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.actualTypeMeta, test.expectedTypeMeta)
		})
	}
}

/*
In this unit test case we will create one diskInfo struct with mock details
and get ObjectMeta value from getObjectMeta() and compare them
*/
func TestGetObjectMeta(t *testing.T) {
	diskUid := "disk-uuid"
	hostName := "host-name"

	// add mock fields in diskInfo struct
	fakeDiskInfo := NewDiskInfo()
	fakeDiskInfo.Uuid = diskUid
	fakeDiskInfo.HostName = hostName

	//create fakeObjectMeta using mock data
	fakeObjectMeta := metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   diskUid,
	}
	fakeObjectMeta.Labels[NDMHostKey] = hostName
	fakeObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType

	tests := map[string]struct {
		actualObjectMeta   metav1.ObjectMeta
		expectedObjectMeta metav1.ObjectMeta
	}{
		"create one mock objectMeta object and compare it": {actualObjectMeta: fakeDiskInfo.getObjectMeta(), expectedObjectMeta: fakeObjectMeta},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.actualObjectMeta, test.expectedObjectMeta)
		})
	}
}

/*
In this unit test case we will create one diskInfo struct with mock details
and get ObjectMeta value from getObjectMeta() and compare them
*/
func TestGetStatus(t *testing.T) {
	// add mock fields in diskInfo struct
	fakeDiskInfo := NewDiskInfo()

	//create fakeObjectMeta using mock data
	fakeDiskStatus := apis.DiskStatus{
		State: NDMActive,
	}

	tests := map[string]struct {
		actualDiskStatus   apis.DiskStatus
		expectedDiskStatus apis.DiskStatus
	}{
		"create one mock DiskStatus object and compare it": {actualDiskStatus: fakeDiskInfo.getStatus(), expectedDiskStatus: fakeDiskStatus},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.actualDiskStatus, test.expectedDiskStatus)
		})
	}
}

/*
In this unit test case we will create one diskInfo struct with mock details
and get DiskLinks value from getDiskLinks() and compare them
*/
func TestGetDiskLinks(t *testing.T) {
	fakeByIdLink := "fake-by-Id-link"
	fakeByPathLink := "fake-by-Path-link"
	byIdLink := make([]string, 0)
	byPathLink := make([]string, 0)
	byIdLink = append(byIdLink, fakeByIdLink)
	byPathLink = append(byPathLink, fakeByPathLink)
	// add mock fields in diskInfo struct
	fakeDiskInfo := NewDiskInfo()
	fakeDiskInfo.ByIdDevLinks = byIdLink
	fakeDiskInfo.ByPathDevLinks = byPathLink

	// Creating one DevLink struct using mock value
	fakeDevLinks := make([]apis.DiskDevLink, 0)
	fakeByIdLinks := apis.DiskDevLink{
		Kind:  "by-id",
		Links: byIdLink,
	}
	fakeByPathLinks := apis.DiskDevLink{
		Kind:  "by-path",
		Links: byPathLink,
	}
	fakeDevLinks = append(fakeDevLinks, fakeByIdLinks)
	fakeDevLinks = append(fakeDevLinks, fakeByPathLinks)

	tests := map[string]struct {
		actualDiskLink   []apis.DiskDevLink
		expectedDiskLink []apis.DiskDevLink
	}{
		"create one fakeDevLinks object and compare it": {actualDiskLink: fakeDiskInfo.getDiskLinks(), expectedDiskLink: fakeDevLinks},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.actualDiskLink, test.expectedDiskLink)
		})
	}
}

/*
In this unit test case we will create one diskInfo struct with mock details
and get Disk value from ToDisk() and compare them
*/
func TestToDisk(t *testing.T) {
	diskUid := "disk-uuid"
	hostName := "host-name"
	fakeDevLinks := make([]apis.DiskDevLink, 0)
	// add mock fields in diskInfo struct
	fakeDiskInfo := NewDiskInfo()
	fakeDiskInfo.Uuid = diskUid
	fakeDiskInfo.HostName = hostName

	// Creating one Disk struct using mock value
	expectedDisk := apis.Disk{}
	fakeTypeMeta := metav1.TypeMeta{
		Kind:       NDMKind,
		APIVersion: NDMVersion,
	}
	expectedDisk.TypeMeta = fakeTypeMeta
	fakeObjectMeta := metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   diskUid,
	}
	fakeObjectMeta.Labels[NDMHostKey] = hostName
	fakeObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	expectedDisk.ObjectMeta = fakeObjectMeta
	expectedDisk.Spec.DevLinks = fakeDevLinks
	expectedDisk.Status.State = NDMActive

	tests := map[string]struct {
		actualDisk   apis.Disk
		expectedDisk apis.Disk
	}{
		"create one diskinfo object and convert it into api.Disk": {actualDisk: fakeDiskInfo.ToDisk(), expectedDisk: expectedDisk}}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
		})
	}
}
