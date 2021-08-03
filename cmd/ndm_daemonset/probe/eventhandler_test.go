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
	"sync"
	"testing"

	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ndmFakeClientset "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	ignoreDiskDevPath = "/dev/sdZ"
	fakeHostName      = "node-name"
	fakeModel         = "fake-disk-model"
	fakeSerial        = "fake-disk-serial"
	fakeVendor        = "fake-disk-vendor"
	fakeWWN           = "fake-WWN"
	fakeBDType        = "blockdevice"
)

var (
	fakeBD1 = blockdevice.BlockDevice{
		Identifier: blockdevice.Identifier{
			DevPath: "/dev/sdX",
		},
		DeviceAttributes: blockdevice.DeviceAttribute{
			WWN:    fakeWWN,
			Serial: fakeSerial,
		},
	}
	fakeBD2 = blockdevice.BlockDevice{
		Identifier: blockdevice.Identifier{
			DevPath: ignoreDiskDevPath,
		},
		DeviceAttributes: blockdevice.DeviceAttribute{
			WWN: fakeWWN,
		},
	}
)

var (
	fakeBD1Uuid, _ = generateUUID(fakeBD1)
	fakeBD2Uuid, _ = generateUUID(fakeBD2)
)

func mockEmptyBlockDeviceCr() apis.BlockDevice {
	fakeBDr := apis.BlockDevice{}
	fakeObjectMeta := metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   fakeBD1Uuid,
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
	s.AddKnownTypes(apis.GroupVersion, deviceR)
	s.AddKnownTypes(apis.GroupVersion, deviceList)

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
	return fakeDiskInfo.DevPath != ignoreDiskDevPath
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
		BDHierarchy:    make(blockdevice.Hierarchy),
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
	eventmsg = append(eventmsg, &fakeBD1)
	// blockdevice-2 details
	eventmsg = append(eventmsg, &fakeBD2)
	// Creating one event message
	eventDetails := controller.EventMessage{
		Action:  libudevwrapper.UDEV_ACTION_ADD,
		Devices: eventmsg,
	}
	probeEvent.addBlockDeviceEvent(eventDetails)
	// Retrieve disk resource
	cdr1, err1 := fakeController.GetBlockDevice(fakeBD1Uuid)

	// Retrieve disk resource
	cdr2, _ := fakeController.GetBlockDevice(fakeBD2Uuid)
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
	fakeDr.Spec.Path = "/dev/sdX"

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
			compareBlockDevice(t, test.expectedDisk, test.actualDisk)
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
		BDHierarchy: blockdevice.Hierarchy{
			"/dev/sdX": fakeBD1,
		},
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
	eventmsg = append(eventmsg, &fakeBD1)
	eventDetails := controller.EventMessage{
		Action:  libudevwrapper.UDEV_ACTION_REMOVE,
		Devices: eventmsg,
	}
	probeEvent.deleteBlockDeviceEvent(eventDetails)

	// Retrieve resources
	bdR1, err1 := fakeController.GetBlockDevice(fakeBD1Uuid)

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
			compareBlockDevice(t, test.expectedBD, test.actualBD)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

// compareBlockDevice is the custom blockdevice comparison function. Only those values that need to be checked
// for equality will be checked here. Resource version field will not be checked as it
// will be updated on every write. Refer https://github.com/kubernetes-sigs/controller-runtime/pull/620
func compareBlockDevice(t *testing.T, bd1, bd2 apis.BlockDevice) {
	assert.Equal(t, bd1.Name, bd2.Name)
	assert.Equal(t, bd1.Labels, bd2.Labels)
	// devlinks will be compared separately
	assert.Equal(t, len(bd1.Spec.DevLinks), len(bd2.Spec.DevLinks))
	if len(bd1.Spec.DevLinks) != len(bd2.Spec.DevLinks) {
		assert.Fail(t, "Devlinks, expected: %+v \n actual: %+v", bd1.Spec.DevLinks, bd2.Spec.DevLinks)
		return
	}
	// compare each set of devlinks
	for i := 0; i < len(bd1.Spec.DevLinks); i++ {
		assert.True(t, unorderedEqual(bd1.Spec.DevLinks[i].Links, bd2.Spec.DevLinks[i].Links))
	}
	// links will be made nil since they are already compared
	bd1.Spec.DevLinks = nil
	bd2.Spec.DevLinks = nil

	assert.Equal(t, bd1.Spec, bd2.Spec)
	assert.Equal(t, bd1.Status, bd2.Status)

}

// compareBlockDeviceList is the custom comparison function for blockdevice list
func compareBlockDeviceList(t *testing.T, bdList1, bdList2 apis.BlockDeviceList) {
	assert.Equal(t, len(bdList1.Items), len(bdList2.Items))
	for i := 0; i < len(bdList2.Items); i++ {
		compareBlockDevice(t, bdList1.Items[i], bdList2.Items[i])
	}
}

func TestIsParentOrSlaveDevice(t *testing.T) {
	tests := map[string]struct {
		bd             blockdevice.BlockDevice
		erroredDevices []string
		want           bool
	}{
		"no devices in errored state": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent:     "/dev/sda",
					Partitions: nil,
					Holders:    nil,
					Slaves:     nil,
				},
			},
			erroredDevices: nil,
			want:           false,
		},
		"multiple devices in errored state with no matching BD": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent:     "/dev/sda",
					Partitions: nil,
					Holders:    nil,
					Slaves:     nil,
				},
			},
			erroredDevices: []string{"/dev/sdb", "/dev/sdc"},
			want:           false,
		},
		"one device in errored state that is the parent of the given BD": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent:     "/dev/sda",
					Partitions: nil,
					Holders:    nil,
					Slaves:     nil,
				},
			},
			erroredDevices: []string{"/dev/sda"},
			want:           true,
		},
		"one device in errored state that that the BD is a slave to": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/dm-0",
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent:     "",
					Partitions: nil,
					Holders:    nil,
					Slaves:     []string{"/dev/sda", "/dev/sdb"},
				},
			},
			erroredDevices: []string{"/dev/sda", "/dev/sdc"},
			want:           true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := isParentOrSlaveDevice(tt.bd, tt.erroredDevices)
			assert.Equal(t, tt.want, got)
		})
	}
}
