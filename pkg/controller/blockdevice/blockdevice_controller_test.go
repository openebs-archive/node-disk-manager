/*
Copyright 2019 The OpenEBS Authors

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

package blockdevice

import (
	"context"
	//"math/rand"
	//"reflect"
	"fmt"
	"testing"

	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	openebsv1alpha1 "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	fakeHostName        = "fake-hostname"
	diskName            = "disk-example"
	deviceName          = "blockdevice-example"
	namespace           = ""
	capacity     uint64 = 1024000
)

// TestDeviceController runs ReconcileDisk.Reconcile() against a
// fake client that tracks a BlockDevice object.
// Test description:
// Create a disk obj and associated blockdevice obj, check status of blockdevice obj,
// it should be Active, now mark disk Inactive and trigger reconcile logic
// on blockdevice, it would mark blockdevice Inactive as well. Check status of blockdevice,
// this time it should be Inactive.
func TestDeviceController(t *testing.T) {

	// Set the logger to development mode for verbose logs.
	logf.SetLogger(logf.ZapLogger(true))

	// Create a fake client to mock API calls.
	cl, s := CreateFakeClient(t)

	// Create a ReconcileBlockDevice object with the scheme and fake client.
	r := &ReconcileBlockDevice{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      deviceName,
			Namespace: namespace,
		},
	}

	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Log("reconcile did not requeue request as expected")
	}

	deviceInstance := &openebsv1alpha1.BlockDevice{}
	err = r.client.Get(context.TODO(), req.NamespacedName, deviceInstance)
	if err != nil {
		t.Errorf("get deviceInstance : (%v)", err)
	}

	// Disk Status state should be Active as expected.
	if deviceInstance.Status.State == ndm.NDMActive {
		t.Logf("BlockDevice Object state:%v match expected state:%v", deviceInstance.Status.State, ndm.NDMActive)
	} else {
		t.Fatalf("BlockDevice Object state:%v did not match expected state:%v", deviceInstance.Status.State, ndm.NDMActive)
	}

	// Fetch the Disk CR
	diskInstance := &openebsv1alpha1.Disk{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: diskName, Namespace: namespace}, diskInstance)
	if err != nil {
		t.Errorf("get diskInstance : (%v)", err)
	}

	diskInstance.Status.State = ndm.NDMInactive
	err = r.client.Update(context.TODO(), diskInstance)
	if err != nil {
		t.Errorf("Error while updating disk obj")
	}

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Log("reconcile did not requeue request as expected")
	}

	// Check Status of Disk object.
	err = r.client.Get(context.TODO(), req.NamespacedName, deviceInstance)
	if err != nil {
		t.Errorf("get deviceInstance : (%v)", err)
	}

	// Disk Status state should be InActive as expected.
	if deviceInstance.Status.State == ndm.NDMInactive {
		t.Logf("BlockDevice Object state:%v match expected state:%v", deviceInstance.Status.State, ndm.NDMInactive)
	} else {
		t.Errorf("BlockDevice Object state:%v did not match expected state:%v", deviceInstance.Status.State, ndm.NDMInactive)
	}
}

func GetFakeDeviceObject() *openebsv1alpha1.BlockDevice {
	device := &openebsv1alpha1.BlockDevice{}
	labels := map[string]string{ndm.NDMManagedKey: ndm.TrueString}

	TypeMeta := metav1.TypeMeta{
		Kind:       ndm.NDMBlockDeviceKind,
		APIVersion: ndm.NDMVersion,
	}
	ObjectMeta := metav1.ObjectMeta{
		Labels:    labels,
		Name:      deviceName,
		Namespace: namespace,
	}

	Spec := openebsv1alpha1.DeviceSpec{
		Path: "dev/disk-fake-path",
		Capacity: openebsv1alpha1.DeviceCapacity{
			Storage: capacity, // Set blockdevice size.
		},
		DevLinks:    make([]openebsv1alpha1.DeviceDevLink, 0),
		Partitioned: ndm.NDMNotPartitioned,
	}

	device.ObjectMeta = ObjectMeta
	device.TypeMeta = TypeMeta
	device.Status.ClaimState = openebsv1alpha1.BlockDeviceUnclaimed
	device.Status.State = ndm.NDMActive
	device.Spec = Spec
	return device
}

func GetFakeDiskObject() *openebsv1alpha1.Disk {

	disk := &openebsv1alpha1.Disk{
		TypeMeta: metav1.TypeMeta{
			Kind:       ndm.NDMDiskKind,
			APIVersion: ndm.NDMVersion,
		},

		ObjectMeta: metav1.ObjectMeta{
			Labels:    make(map[string]string),
			Name:      diskName,
			Namespace: namespace,
		},

		Spec: openebsv1alpha1.DiskSpec{
			Path: "dev/disk-fake-path",
			Capacity: openebsv1alpha1.DiskCapacity{
				Storage: capacity, // Set disk size.
			},
			Details: openebsv1alpha1.DiskDetails{
				Model:  "disk-fake-model",
				Serial: "disk-fake-serial",
				Vendor: "disk-fake-vendor",
			},
			DevLinks: make([]openebsv1alpha1.DiskDevLink, 0),
		},
		Status: openebsv1alpha1.DiskStatus{
			State: ndm.NDMActive,
		},
	}
	disk.ObjectMeta.Labels[ndm.KubernetesHostNameLabel] = fakeHostName
	return disk
}

func CreateFakeClient(t *testing.T) (client.Client, *runtime.Scheme) {

	diskR := GetFakeDiskObject()

	diskList := &openebsv1alpha1.DiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Disk",
			APIVersion: "",
		},
	}

	deviceR := GetFakeDeviceObject()

	deviceList := &openebsv1alpha1.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "",
		},
	}

	diskObjs := []runtime.Object{diskR}
	s := scheme.Scheme

	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, diskR)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, diskList)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, deviceR)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, deviceList)

	fakeNdmClient := fake.NewFakeClient(diskObjs...)
	if fakeNdmClient == nil {
		fmt.Println("NDMClient is not created")
	}
	err := fakeNdmClient.Create(context.TODO(), deviceR)
	if err != nil {
		fmt.Println("BlockDevice object is not created")
	}
	return fakeNdmClient, s
}
