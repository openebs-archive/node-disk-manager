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

<<<<<<< HEAD
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
=======
	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
>>>>>>> 3bfc5e1e... Inital project structuring and adding BlockDevice type
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ndmFakeClientset "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	fakeHostName = "fake-host-name"

	fakeDeviceUID    = "fake-blockdevice-uid"
	newFakeDeviceUID = "new-fake-blockdevice-uid"
)

var (
	// mock data for blockdevice
	fakeDeviceCapacity = apis.DeviceCapacity{
		Storage: 100000,
	}

	fakeDeviceDetails = apis.DeviceDetails{
		Model:  "disk-fake-model",
		Serial: "disk-fake-serial",
		Vendor: "disk-fake-vendor",
	}

<<<<<<< HEAD
	fakeDeviceObj = apis.DeviceSpec{
=======
	fakeDeviceObj = apis.BlockDeviceSpec{
>>>>>>> 3bfc5e1e... Inital project structuring and adding BlockDevice type
		Path:        "dev/disk-fake-path",
		Capacity:    fakeDeviceCapacity,
		Details:     fakeDeviceDetails,
		DevLinks:    make([]apis.DeviceDevLink, 0),
		Partitioned: NDMNotPartitioned,
	}

	fakeDeviceTypeMeta = metav1.TypeMeta{
		Kind:       NDMBlockDeviceKind,
		APIVersion: NDMVersion,
	}

	fakeDeviceObjectMeta = metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   fakeDeviceUID,
	}

<<<<<<< HEAD
	fakeDeviceStatus = apis.DeviceStatus{
=======
	fakeDeviceStatus = apis.BlockDeviceStatus{
>>>>>>> 3bfc5e1e... Inital project structuring and adding BlockDevice type
		ClaimState: apis.BlockDeviceUnclaimed,
		State:      NDMActive,
	}

	fakeDevice = apis.BlockDevice{
		TypeMeta:   fakeDeviceTypeMeta,
		ObjectMeta: fakeDeviceObjectMeta,
		Spec:       fakeDeviceObj,
		Status:     fakeDeviceStatus,
	}

	// mock data for blockdevice
	newFakeDeviceCapacity = apis.DeviceCapacity{
		Storage: 200000,
	}

	newFakeDeviceDetails = apis.DeviceDetails{
		Model:  "disk-fake-model-new",
		Serial: "disk-fake-serial-new",
		Vendor: "disk-fake-vendor-new",
	}

<<<<<<< HEAD
	newFakeDeviceObj = apis.DeviceSpec{
=======
	newFakeDeviceObj = apis.BlockDeviceSpec{
>>>>>>> 3bfc5e1e... Inital project structuring and adding BlockDevice type
		Path:     "dev/disk-fake-path-new",
		Capacity: newFakeDeviceCapacity,
		Details:  newFakeDeviceDetails,
		DevLinks: make([]apis.DeviceDevLink, 0),
	}

	newFakeDeviceTypeMeta = metav1.TypeMeta{
		Kind:       NDMBlockDeviceKind,
		APIVersion: NDMVersion,
	}

	newFakeDeviceObjectMeta = metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   newFakeDeviceUID,
	}

	newFakeDeviceStatus = apis.BlockDeviceStatus{
		ClaimState: apis.BlockDeviceUnclaimed,
		State:      NDMActive,
	}

	newFakeDevice = apis.BlockDevice{
		TypeMeta:   newFakeDeviceTypeMeta,
		ObjectMeta: newFakeDeviceObjectMeta,
		Spec:       newFakeDeviceObj,
		Status:     newFakeDeviceStatus,
	}
)

func CreateFakeClient(t *testing.T) client.Client {

	deviceR := &apis.BlockDevice{
		ObjectMeta: metav1.ObjectMeta{
			Labels: make(map[string]string),
			Name:   "dummy-blockdevice",
		},
	}

	deviceList := &apis.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "",
		},
	}

	s := scheme.Scheme

<<<<<<< HEAD
	s.AddKnownTypes(apis.SchemeGroupVersion, deviceR)
	s.AddKnownTypes(apis.SchemeGroupVersion, deviceList)
=======
	s.AddKnownTypes(apis.GroupVersion, deviceR)
	s.AddKnownTypes(apis.GroupVersion, deviceList)
>>>>>>> 3bfc5e1e... Inital project structuring and adding BlockDevice type

	fakeNdmClient := ndmFakeClientset.NewFakeClient()
	if fakeNdmClient == nil {
		fmt.Println("NDMClient is not created")
	}
	return fakeNdmClient
}
