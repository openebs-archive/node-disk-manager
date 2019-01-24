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
	"fmt"
	"testing"

	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ndmFakeClientset "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	fakeDiskUid     = "fake-disk-uid"
	newFakeDiskUid  = "new-fake-disk-uid"
	fakeHostName    = "fake-host-name"
	newFakeHostName = "new-fake-host-name"

	fakeDeviceUid    = "fake-device-uid"
	newFakeDeviceUid = "new-fake-device-uid"
)

var (
	// mock data for new disk.
	fakeCapacity = apis.DiskCapacity{
		Storage: 100000,
	}

	fakeDetails = apis.DiskDetails{
		Model:  "disk-fake-model",
		Serial: "disk-fake-serial",
		Vendor: "disk-fake-vendor",
	}

	fakeObj = apis.DiskSpec{
		Path:     "dev/disk-fake-path",
		Capacity: fakeCapacity,
		Details:  fakeDetails,
		DevLinks: make([]apis.DiskDevLink, 0),
	}

	fakeTypeMeta = metav1.TypeMeta{
		Kind:       NDMDiskKind,
		APIVersion: NDMVersion,
	}

	fakeObjectMeta = metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   fakeDiskUid,
	}

	fakeDiskStatus = apis.DiskStatus{
		State: NDMActive,
	}

	fakeDr = apis.Disk{
		TypeMeta:   fakeTypeMeta,
		ObjectMeta: fakeObjectMeta,
		Spec:       fakeObj,
		Status:     fakeDiskStatus,
	}

	// mock data for new disk.
	newFakeCapacity = apis.DiskCapacity{
		Storage: 200000,
	}

	newFakeDetails = apis.DiskDetails{
		Model:  "disk-fake-model-new",
		Serial: "disk-fake-serial-new",
		Vendor: "disk-fake-vendor-new",
	}

	newFakeObj = apis.DiskSpec{
		Path:     "dev/disk-fake-path-new",
		Capacity: newFakeCapacity,
		Details:  newFakeDetails,
		DevLinks: make([]apis.DiskDevLink, 0),
	}

	newFakeTypeMeta = metav1.TypeMeta{
		Kind:       NDMDiskKind,
		APIVersion: NDMVersion,
	}

	newFakeObjectMeta = metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   newFakeDiskUid,
	}

	newFakeDiskStatus = apis.DiskStatus{
		State: NDMActive,
	}

	newFakeDr = apis.Disk{
		TypeMeta:   newFakeTypeMeta,
		ObjectMeta: newFakeObjectMeta,
		Spec:       newFakeObj,
		Status:     newFakeDiskStatus,
	}

	// mock data for device
	fakeDeviceCapacity = apis.DeviceCapacity{
		Storage: 100000,
	}

	fakeDeviceDetails = apis.DeviceDetails{
		Model:  "disk-fake-model",
		Serial: "disk-fake-serial",
		Vendor: "disk-fake-vendor",
	}

	fakeDeviceObj = apis.DeviceSpec{
		Path:        "dev/disk-fake-path",
		Capacity:    fakeDeviceCapacity,
		Details:     fakeDeviceDetails,
		DevLinks:    make([]apis.DeviceDevLink, 0),
		Partitioned: NDMNotPartitioned,
	}

	fakeDeviceTypeMeta = metav1.TypeMeta{
		Kind:       NDMDeviceKind,
		APIVersion: NDMVersion,
	}

	fakeDeviceObjectMeta = metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   fakeDeviceUid,
	}

	fakeDeviceClaimState = apis.DeviceClaimState{
		State: NDMUnclaimed,
	}

	fakeDeviceStatus = apis.DeviceStatus{
		State: NDMActive,
	}

	fakeDevice = apis.Device{
		TypeMeta:   fakeDeviceTypeMeta,
		ObjectMeta: fakeDeviceObjectMeta,
		Spec:       fakeDeviceObj,
		ClaimState: fakeDeviceClaimState,
		Status:     fakeDeviceStatus,
	}

	// mock data for device
	newFakeDeviceCapacity = apis.DeviceCapacity{
		Storage: 200000,
	}

	newFakeDeviceDetails = apis.DeviceDetails{
		Model:  "disk-fake-model-new",
		Serial: "disk-fake-serial-new",
		Vendor: "disk-fake-vendor-new",
	}

	newFakeDeviceObj = apis.DeviceSpec{
		Path:     "dev/disk-fake-path-new",
		Capacity: newFakeDeviceCapacity,
		Details:  newFakeDeviceDetails,
		DevLinks: make([]apis.DeviceDevLink, 0),
	}

	newFakeDeviceTypeMeta = metav1.TypeMeta{
		Kind:       NDMDeviceKind,
		APIVersion: NDMVersion,
	}

	newFakeDeviceObjectMeta = metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   newFakeDeviceUid,
	}

	newFakeDeviceClaimState = apis.DeviceClaimState{
		State: NDMUnclaimed,
	}

	newFakeDeviceStatus = apis.DeviceStatus{
		State: NDMActive,
	}

	newFakeDevice = apis.Device{
		TypeMeta:   newFakeDeviceTypeMeta,
		ObjectMeta: newFakeDeviceObjectMeta,
		Spec:       newFakeDeviceObj,
		ClaimState: newFakeDeviceClaimState,
		Status:     newFakeDeviceStatus,
	}
)

func CreateFakeClient(t *testing.T) client.Client {
	diskR := &apis.Disk{
		ObjectMeta: metav1.ObjectMeta{
			Labels: make(map[string]string),
			Name:   "dummy-disk",
		},
	}

	diskList := &apis.DiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Disk",
			APIVersion: "",
		},
	}

	deviceR := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Labels: make(map[string]string),
			Name:   "dummy-device",
		},
	}

	deviceList := &apis.DeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Device",
			APIVersion: "",
		},
	}

	objs := []runtime.Object{diskR}
	s := scheme.Scheme

	s.AddKnownTypes(apis.SchemeGroupVersion, diskR)
	s.AddKnownTypes(apis.SchemeGroupVersion, diskList)
	s.AddKnownTypes(apis.SchemeGroupVersion, deviceR)
	s.AddKnownTypes(apis.SchemeGroupVersion, deviceList)

	fakeNdmClient := ndmFakeClientset.NewFakeClient(objs...)
	if fakeNdmClient == nil {
		fmt.Println("NDMClient is not created")
	}
	return fakeNdmClient
}
