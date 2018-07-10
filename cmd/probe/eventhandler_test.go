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

var (
	mockuid      = "fake-disk-uid"
	fakeHostName = "node-name"
)

func mockEmptyDiskCr() apis.Disk {
	fakeDr := apis.Disk{}
	fakeObjectMeta := metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   mockuid,
	}
	fakeTypeMeta := metav1.TypeMeta{
		Kind:       controller.NDMKind,
		APIVersion: controller.NDMVersion,
	}
	fakeDr.ObjectMeta = fakeObjectMeta
	fakeDr.TypeMeta = fakeTypeMeta
	fakeDr.Status.State = controller.NDMActive
	return fakeDr
}
func TestAddDiskEvent(t *testing.T) {
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
	probeEvent := &ProbeEvent{
		Controller: fakeController,
	}
	// Creating one event message
	eventmsg := make([]*controller.DiskInfo, 0)
	deviceDetails := &controller.DiskInfo{}
	deviceDetails.ProbeIdentifiers.Uuid = mockuid
	eventmsg = append(eventmsg, deviceDetails)
	eventDetails := controller.EventMessage{
		Action:  libudevwrapper.UDEV_ACTION_ADD,
		Devices: eventmsg,
	}
	probeEvent.addDiskEvent(eventDetails)
	cdr1, err1 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(mockuid, metav1.GetOptions{})
	probeEvent.addDiskEvent(eventDetails)
	cdr2, err2 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(mockuid, metav1.GetOptions{})
	// Create one fake disk resource
	fakeDr := mockEmptyDiskCr()
	fakeDr.ObjectMeta.Labels[controller.NDMHostKey] = fakeController.HostName

	tests := map[string]struct {
		actualDisk    apis.Disk
		expectedDisk  apis.Disk
		actualError   error
		expectedError error
	}{
		"add event for resouce with 'fake-disk-uid' uuid for create resource": {actualDisk: *cdr1, expectedDisk: fakeDr, actualError: err1, expectedError: nil},
		"add event for resouce with 'fake-disk-uid' uuid for update resource": {actualDisk: *cdr2, expectedDisk: fakeDr, actualError: err2, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestDeleteDiskEvent(t *testing.T) {
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
	// Create one fake disk resource
	fakeDr := mockEmptyDiskCr()
	fakeDr.ObjectMeta.Labels[controller.NDMHostKey] = fakeController.HostName
	fakeController.CreateDisk(fakeDr)

	probeEvent := &ProbeEvent{
		Controller: fakeController,
	}
	eventmsg := make([]*controller.DiskInfo, 0)
	deviceDetails := &controller.DiskInfo{}
	deviceDetails.ProbeIdentifiers.Uuid = mockuid
	eventmsg = append(eventmsg, deviceDetails)
	eventDetails := controller.EventMessage{
		Action:  libudevwrapper.UDEV_ACTION_REMOVE,
		Devices: eventmsg,
	}
	probeEvent.deleteDiskEvent(eventDetails)
	cdr1, err1 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(mockuid, metav1.GetOptions{})
	fakeDr.Status.State = controller.NDMInactive
	tests := map[string]struct {
		actualDisk    apis.Disk
		expectedDisk  apis.Disk
		actualError   error
		expectedError error
	}{
		"remove resouce with 'fake-disk-uid' uuid": {actualDisk: *cdr1, expectedDisk: fakeDr, actualError: err1, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}
