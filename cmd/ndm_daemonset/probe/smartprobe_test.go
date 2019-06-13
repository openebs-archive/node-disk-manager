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

	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	smart "github.com/openebs/node-disk-manager/pkg/smart"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func mockOsDiskToAPIBySmart() (apis.Disk, error) {

	mockOsDiskDetails, err := smart.MockScsiBasicDiskInfo()
	if err != nil {
		return apis.Disk{}, err
	}

	mockOsDiskDetailsFromUdev, err := libudevwrapper.MockDiskDetails()
	if err != nil {
		return apis.Disk{}, err
	}

	fakeDetails := apis.DiskDetails{
		Compliance:       mockOsDiskDetails.Compliance,
		FirmwareRevision: mockOsDiskDetails.FirmwareRevision,
	}

	fakeCapacity := apis.DiskCapacity{
		Storage:           mockOsDiskDetails.Capacity,
		LogicalSectorSize: mockOsDiskDetails.LBSize,
	}

	fakeObj := apis.DiskSpec{
		Capacity: fakeCapacity,
		Details:  fakeDetails,
		FileSystem: apis.FileSystemInfo{
			IsFormated: true,
		},
	}

	devLinks := make([]apis.DiskDevLink, 0)
	fakeObj.DevLinks = devLinks

	fakeTypeMeta := metav1.TypeMeta{
		Kind:       controller.NDMDiskKind,
		APIVersion: controller.NDMVersion,
	}

	fakeObjectMeta := metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   mockOsDiskDetailsFromUdev.Uid,
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
}

func TestFillDiskDetailsBySmart(t *testing.T) {
	mockOsDiskDetails, err := smart.MockScsiBasicDiskInfo()
	if err != nil {
		t.Fatal(err)
	}
	sProbe := smartProbe{}
	actualDiskInfo := controller.NewDiskInfo()
	actualDiskInfo.ProbeIdentifiers.SmartIdentifier = mockOsDiskDetails.DevPath
	sProbe.FillDiskDetails(actualDiskInfo)
	expectedDiskInfo := &controller.DiskInfo{}
	expectedDiskInfo.ProbeIdentifiers.SmartIdentifier = mockOsDiskDetails.DevPath
	expectedDiskInfo.Capacity = mockOsDiskDetails.Capacity
	expectedDiskInfo.FirmwareRevision = mockOsDiskDetails.FirmwareRevision
	expectedDiskInfo.Compliance = mockOsDiskDetails.Compliance
	expectedDiskInfo.LogicalSectorSize = mockOsDiskDetails.LBSize
	expectedDiskInfo.DiskType = "disk"
	assert.Equal(t, expectedDiskInfo, actualDiskInfo)
}

func TestSmartProbe(t *testing.T) {
	mockOsDiskDetails, err := smart.MockScsiBasicDiskInfo()
	if err != nil {
		t.Fatal(err)
	}
	mockOsDiskDetailsUsingUdev, err := libudevwrapper.MockDiskDetails()
	if err != nil {
		t.Fatal(err)
	}

	fakeHostName := "node-name"
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	probes := make([]*controller.Probe, 0)
	filters := make([]*controller.Filter, 0)
	mutex := &sync.Mutex{}
	fakeController := &controller.Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
		Mutex:         mutex,
		Probes:        probes,
		Filters:       filters,
	}

	smartProbe := newSmartProbe("fakeController")
	var pi controller.ProbeInterface = smartProbe

	newRegisterProbe := &registerProbe{
		priority:   1,
		name:       "smart probe",
		state:      true,
		pi:         pi,
		controller: fakeController,
	}

	newRegisterProbe.register()

	// Add one filter
	filter := &alwaysTrueFilter{}
	filter1 := &controller.Filter{
		Name:      "filter1",
		State:     true,
		Interface: filter,
	}

	fakeController.AddNewFilter(filter1)
	probeEvent := &ProbeEvent{
		Controller: fakeController,
	}

	eventmsg := make([]*controller.DiskInfo, 0)
	deviceDetails := &controller.DiskInfo{}
	deviceDetails.ProbeIdentifiers.Uuid = mockOsDiskDetailsUsingUdev.Uid
	deviceDetails.ProbeIdentifiers.SmartIdentifier = mockOsDiskDetails.DevPath
	eventmsg = append(eventmsg, deviceDetails)

	eventDetails := controller.EventMessage{
		Action:  libudevwrapper.UDEV_ACTION_ADD,
		Devices: eventmsg,
	}
	probeEvent.addDiskEvent(eventDetails)

	// Retrieve disk resource
	cdr1, err1 := fakeController.GetDisk(mockOsDiskDetailsUsingUdev.Uid)
	if err1 != nil {
		t.Fatal(err1)
	}

	fakeDr, err := mockOsDiskToAPIBySmart()
	if err != nil {
		t.Fatal(err)
	}
	fakeDr.ObjectMeta.Labels[controller.NDMHostKey] = fakeController.HostName
	fakeDr.ObjectMeta.Labels[controller.NDMDiskTypeKey] = "disk"
	fakeDr.ObjectMeta.Labels[controller.NDMManagedKey] = controller.TrueString

	tests := map[string]struct {
		actualDisk    apis.Disk
		expectedDisk  apis.Disk
		actualError   error
		expectedError error
	}{
		"add event for resource with 'fake-disk-uid' uuid": {actualDisk: *cdr1, expectedDisk: fakeDr, actualError: err1, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}
