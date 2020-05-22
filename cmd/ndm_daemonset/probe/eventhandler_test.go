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
	"fmt"
	"github.com/openebs/node-disk-manager/blockdevice"
	"sync"
	"testing"

	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ndmFakeClientset "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	//mockDiskuid    = "disk-fake-uid"
	mockBDuid      = "blockdevice-fake-uid"
	ignoreDiskUuid = "ignore-disk-uuid"
	fakeHostName   = "node-name"
	fakeModel      = "fake-disk-model"
	fakeSerial     = "fake-disk-serial"
	fakeVendor     = "fake-disk-vendor"
	fakeDiskType   = "disk"
	fakeBDType     = "blockdevice"
)

func mockEmptyBlockDeviceCr() apis.BlockDevice {
	fakeBDr := apis.BlockDevice{}
	fakeObjectMeta := metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   mockBDuid,
	}
	fakeTypeMeta := metav1.TypeMeta{
		Kind:       controller.NDMBlockDeviceKind,
		APIVersion: controller.NDMVersion,
	}
	fakeBDr.ObjectMeta = fakeObjectMeta
	fakeBDr.TypeMeta = fakeTypeMeta
	fakeBDr.Status.State = controller.NDMActive
	fakeBDr.Status.ClaimState = apis.BlockDeviceUnclaimed
	fakeBDr.Spec.DevLinks = make([]apis.DeviceDevLink, 0)
	return fakeBDr
}

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
	s.AddKnownTypes(apis.SchemeGroupVersion, deviceR)
	s.AddKnownTypes(apis.SchemeGroupVersion, deviceList)

	fakeNdmClient := ndmFakeClientset.NewFakeClient()
	if fakeNdmClient == nil {
		fmt.Println("NDMClient is not created")
	}
	return fakeNdmClient
}

type fakeFilter struct{}

func (nf *fakeFilter) Start() {}

func (nf *fakeFilter) Include(fakeDiskInfo *blockdevice.BlockDevice) bool {
	return true
}

func (nf *fakeFilter) Exclude(fakeDiskInfo *blockdevice.BlockDevice) bool {
	return fakeDiskInfo.UUID != ignoreDiskUuid
}

func TestAddBlockDeviceEvent(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	nodeAttributes := make(map[string]string)
	nodeAttributes[controller.HostNameKey] = fakeHostName
	fakeController := &controller.Controller{
		Clientset:      fakeNdmClient,
		Mutex:          &sync.Mutex{},
		Filters:        make([]*controller.Filter, 0),
		Probes:         make([]*controller.Probe, 0),
		NodeAttributes: nodeAttributes,
	}
	//add one filter
	filter := &fakeFilter{}
	filter1 := &controller.Filter{
		Name:      "filter1",
		State:     true,
		Interface: filter,
	}
	fakeController.AddNewFilter(filter1)
	// add one probe
	testProbe := &fakeProbe{}
	probe1 := &controller.Probe{
		Name:      "probe1",
		State:     true,
		Interface: testProbe,
	}
	fakeController.AddNewProbe(probe1)

	probeEvent := &ProbeEvent{
		Controller: fakeController,
	}
	// blockdevice-1 details
	eventmsg := make([]*blockdevice.BlockDevice, 0)
	device1Details := &blockdevice.BlockDevice{}
	device1Details.UUID = mockBDuid
	eventmsg = append(eventmsg, device1Details)
	// blockdevice-2 details
	device2Details := &blockdevice.BlockDevice{}
	device2Details.UUID = ignoreDiskUuid
	eventmsg = append(eventmsg, device2Details)
	// Creating one event message
	eventDetails := controller.EventMessage{
		Action:  libudevwrapper.UDEV_ACTION_ADD,
		Devices: eventmsg,
	}
	probeEvent.addBlockDeviceEvent(eventDetails)
	// Retrieve disk resource
	cdr1, err1 := fakeController.GetBlockDevice(mockBDuid)

	// Retrieve disk resource
	cdr2, _ := fakeController.GetBlockDevice(ignoreDiskUuid)
	if cdr2 != nil {
		t.Error("resource with ignoreDiskUuid should not be present in etcd")
	}
	// Create one fake disk resource
	fakeDr := mockEmptyBlockDeviceCr()
	fakeDr.ObjectMeta.Labels[controller.KubernetesHostNameLabel] = fakeController.NodeAttributes[controller.HostNameKey]
	fakeDr.ObjectMeta.Labels[controller.NDMDeviceTypeKey] = fakeBDType
	fakeDr.ObjectMeta.Labels[controller.NDMManagedKey] = controller.TrueString
	fakeDr.Spec.Details.Model = fakeModel
	fakeDr.Spec.Details.Serial = fakeSerial
	fakeDr.Spec.Details.Vendor = fakeVendor
	fakeDr.Spec.Partitioned = controller.NDMNotPartitioned

	tests := map[string]struct {
		actualDisk    apis.BlockDevice
		expectedDisk  apis.BlockDevice
		actualError   error
		expectedError error
	}{
		"resource with 'fake-disk-uid' uuid for create resource": {actualDisk: *cdr1, expectedDisk: fakeDr, actualError: err1, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestDeleteDiskEvent(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	probes := make([]*controller.Probe, 0)
	nodeAttributes := make(map[string]string)
	nodeAttributes[controller.HostNameKey] = fakeHostName
	mutex := &sync.Mutex{}
	fakeController := &controller.Controller{
		Clientset:      fakeNdmClient,
		Probes:         probes,
		Mutex:          mutex,
		NodeAttributes: nodeAttributes,
	}

	// Create one fake block device resource
	fakeBDr := mockEmptyBlockDeviceCr()
	fakeBDr.ObjectMeta.Labels[controller.KubernetesHostNameLabel] = fakeController.NodeAttributes[controller.HostNameKey]
	fakeBDr.ObjectMeta.Labels[controller.NDMDeviceTypeKey] = fakeBDType
	fakeBDr.ObjectMeta.Labels[controller.NDMManagedKey] = controller.TrueString
	fakeController.CreateBlockDevice(fakeBDr)

	probeEvent := &ProbeEvent{
		Controller: fakeController,
	}
	eventmsg := make([]*blockdevice.BlockDevice, 0)
	deviceDetails := &blockdevice.BlockDevice{}
	deviceDetails.UUID = mockBDuid
	eventmsg = append(eventmsg, deviceDetails)
	eventDetails := controller.EventMessage{
		Action:  libudevwrapper.UDEV_ACTION_REMOVE,
		Devices: eventmsg,
	}
	probeEvent.deleteBlockDeviceEvent(eventDetails)

	// Retrieve resources
	bdR1, err1 := fakeController.GetBlockDevice(mockBDuid)

	fakeBDr.Status.State = controller.NDMInactive
	tests := map[string]struct {
		actualBD      apis.BlockDevice
		expectedBD    apis.BlockDevice
		actualError   error
		expectedError error
	}{
		"remove resource with 'fake-disk-uid' uuid": {actualBD: *bdR1, expectedBD: fakeBDr, actualError: err1, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedBD, test.actualBD)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}
