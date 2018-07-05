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

package probe

import (
	"sync"
	"testing"

	"github.com/openebs/node-disk-manager/cmd/controller"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs.io/v1alpha1"
	ndmFakeClientset "github.com/openebs/node-disk-manager/pkg/client/clientset/versioned/fake"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func mockOsDiskToAPI() (apis.Disk, error) {
	mockOsdiskDeails, err := libudevwrapper.MockDiskDetails()
	if err != nil {
		return apis.Disk{}, err
	}
	fakeCapacity := apis.DiskCapacity{
		Storage: mockOsdiskDeails.Capacity,
	}
	fakeDetails := apis.DiskDetails{
		Model:  mockOsdiskDeails.Model,
		Serial: mockOsdiskDeails.Serial,
		Vendor: mockOsdiskDeails.Vendor,
	}
	fakeObj := apis.DiskSpec{
		Path:     mockOsdiskDeails.DevNode,
		Capacity: fakeCapacity,
		Details:  fakeDetails,
	}
	fakeTypeMeta := metav1.TypeMeta{
		Kind:       controller.NDMKind,
		APIVersion: controller.NDMVersion,
	}
	fakeObjectMeta := metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   mockOsdiskDeails.Uid,
	}
	fakeDiskStatus := apis.DiskStatus{
		State: controller.NDMActive,
	}
	fakeDr := apis.Disk{
		TypeMeta:   fakeTypeMeta,
		ObjectMeta: fakeObjectMeta,
		Spec:       fakeObj,
		Status:     fakeDiskStatus,
	}
	return fakeDr, nil
	//fakeDr.ObjectMeta.Labels[controller.NDMHostKey] = fakeController.HostName
}

func TestFillDiskDetails(t *testing.T) {
	mockOsdiskDeails, err := libudevwrapper.MockDiskDetails()
	if err != nil {
		t.Fatal(err)
	}
	uProbe := udevProbe{}
	actualDiskInfo := controller.NewDiskInfo()
	actualDiskInfo.ProbeIdentifiers.UdevIdentifier = mockOsdiskDeails.SysPath
	uProbe.FillDiskDetails(actualDiskInfo)
	expectedDiskInfo := &controller.DiskInfo{}
	expectedDiskInfo.ProbeIdentifiers.UdevIdentifier = mockOsdiskDeails.SysPath
	expectedDiskInfo.Model = mockOsdiskDeails.Model
	expectedDiskInfo.Path = mockOsdiskDeails.DevNode
	expectedDiskInfo.Serial = mockOsdiskDeails.Serial
	expectedDiskInfo.Vendor = mockOsdiskDeails.Vendor
	expectedDiskInfo.Capacity = mockOsdiskDeails.Capacity
	assert.Equal(t, expectedDiskInfo, actualDiskInfo)
}

func TestUdevProbe(t *testing.T) {
	mockOsdiskDeails, err := libudevwrapper.MockDiskDetails()
	if err != nil {
		t.Fatal(err)
	}
	fakeHostName := "node-name"
	fakeNdmClient := ndmFakeClientset.NewSimpleClientset()
	fakeKubeClient := fake.NewSimpleClientset()
	probes := make([]*controller.Probe, 0)
	mutex := &sync.Mutex{}
	fakeController := &controller.Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
		Probes:        probes,
		Mutex:         mutex,
	}
	udevprobe := newUdevProbe(fakeController)
	var pi controller.ProbeInterface = udevprobe
	newPrgisterProbe := &registerProbe{
		priority:       1,
		probeName:      "udev probe",
		probeState:     true,
		probeInterface: pi,
		controller:     fakeController,
	}
	newPrgisterProbe.register()
	probeEvent := &ProbeEvent{
		Controller: fakeController,
	}
	eventmsg := make([]*controller.DiskInfo, 0)
	deviceDetails := &controller.DiskInfo{}
	deviceDetails.ProbeIdentifiers.Uuid = mockOsdiskDeails.Uid
	deviceDetails.ProbeIdentifiers.UdevIdentifier = mockOsdiskDeails.SysPath
	eventmsg = append(eventmsg, deviceDetails)
	eventDetails := controller.EventMessage{
		Action:  libudevwrapper.UDEV_ACTION_ADD,
		Devices: eventmsg,
	}
	probeEvent.addDiskEvent(eventDetails)
	cdr1, err1 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(mockOsdiskDeails.Uid, metav1.GetOptions{})

	fakeDr, err := mockOsDiskToAPI()
	if err != nil {
		t.Error(err)
	}
	fakeDr.ObjectMeta.Labels[controller.NDMHostKey] = fakeController.HostName
	tests := map[string]struct {
		actualDisk    apis.Disk
		expectedDisk  apis.Disk
		actualError   error
		expectedError error
	}{
		"add event for resouce with 'fake-disk-uid' uuid": {actualDisk: *cdr1, expectedDisk: fakeDr, actualError: err1, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}
