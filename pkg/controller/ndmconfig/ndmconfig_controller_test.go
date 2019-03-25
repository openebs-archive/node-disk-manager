/*
Copyright 2019 The OpenEBS Author

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
package ndmconfig

import (
	"context"
	"fmt"
	"testing"

	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	openebsv1alpha1 "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	corev1 "k8s.io/api/core/v1"
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
	fakeHostName         = "fake-hostname"
	diskName             = "disk-example"
	deviceName           = "device-example"
	ndmconfigName        = "ndmconfig-example"
	podName              = "example"
	namespace            = ""
	capacity      uint64 = 1024000
)

// TestNdmConfigController runs ReconcileNdmConfig.Reconcile() against a
// fake client that tracks a ndmconfig object.
func TestNdmConfigController(t *testing.T) {

	// Set the logger to development mode for verbose logs.
	logf.SetLogger(logf.ZapLogger(true))

	// Create a fake client to mock API calls.
	cl, s := CreateFakeClient(t)

	// Create a ReconcileNdmConfig object with the scheme and fake client.
	r := &ReconcileNdmConfig{client: cl, scheme: s}

	r.CheckDiskDeviceStatus(t, ndm.NDMInactive)
	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource.
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      ndmconfigName,
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

	r.CheckDiskDeviceStatus(t, ndm.NDMActive)

	ndmconfigR := &openebsv1alpha1.NdmConfig{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: ndmconfigName, Namespace: namespace}, ndmconfigR)
	if err != nil {
		fmt.Println("Unable to fetch ndmconfigR object")
	}

	var sec, nsec int64
	deleteTime := metav1.Unix(sec, nsec).Rfc3339Copy()
	ndmconfigR.ObjectMeta.DeletionTimestamp = &deleteTime
	err = r.client.Update(context.TODO(), ndmconfigR)
	if err != nil {
		fmt.Println("ndmconfigR object is not deleted")
	}

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Log("reconcile did not requeue request as expected")
	}

	r.CheckDiskDeviceStatus(t, ndm.NDMUnknown)
}

func (r *ReconcileNdmConfig) CheckDiskDeviceStatus(t *testing.T, State string) {
	deviceInstance := &openebsv1alpha1.Device{}

	err := r.client.Get(context.TODO(), types.NamespacedName{Name: deviceName, Namespace: namespace}, deviceInstance)
	if err != nil {
		t.Errorf("get deviceInstance : (%v)", err)
	}

	// Disk Status state should be Active as expected.
	if deviceInstance.Status.State == State {
		t.Logf("Device Object state:%v match expected state:%v", deviceInstance.Status.State, State)
	} else {
		t.Fatalf("Device Object state:%v did not match expected state:%v", deviceInstance.Status.State, State)
	}

	// Fetch the Disk CR
	diskInstance := &openebsv1alpha1.Disk{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: diskName, Namespace: namespace}, diskInstance)
	if err != nil {
		t.Errorf("get diskInstance : (%v)", err)
	}

	if diskInstance.Status.State == State {
		t.Logf("Disk Object state:%v match expected state:%v", diskInstance.Status.State, State)
	} else {
		t.Fatalf("Disk Object state:%v did not match expected state:%v", diskInstance.Status.State, State)
	}
}

func GetFakeDeviceObject() *openebsv1alpha1.Device {
	device := &openebsv1alpha1.Device{}

	TypeMeta := metav1.TypeMeta{
		Kind:       ndm.NDMDeviceKind,
		APIVersion: ndm.NDMVersion,
	}

	ObjectMeta := metav1.ObjectMeta{
		Labels:    make(map[string]string),
		Name:      deviceName,
		Namespace: namespace,
	}

	Spec := openebsv1alpha1.DeviceSpec{
		Path: "dev/disk-fake-path",
		Capacity: openebsv1alpha1.DeviceCapacity{
			Storage: capacity, // Set device size.
		},
		DevLinks:    make([]openebsv1alpha1.DeviceDevLink, 0),
		Partitioned: ndm.NDMNotPartitioned,
	}

	device.ObjectMeta = ObjectMeta
	device.TypeMeta = TypeMeta
	device.ClaimState.State = ndm.NDMUnclaimed
	device.Status.State = ndm.NDMInactive
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
			State: ndm.NDMInactive,
		},
	}
	disk.ObjectMeta.Labels[ndm.NDMHostKey] = fakeHostName
	return disk
}

func GetFakeNdmConfigObject() *openebsv1alpha1.NdmConfig {

	ndmconfigCR := &openebsv1alpha1.NdmConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NdmConfig",
			APIVersion: "openebs.io/v1alpha1",
		},

		ObjectMeta: metav1.ObjectMeta{
			Labels:    make(map[string]string),
			Name:      ndmconfigName,
			Namespace: namespace,
		},

		Status: openebsv1alpha1.NdmConfigStatus{
			Phase: openebsv1alpha1.NdmConfigPhaseInit,
		},
	}
	ndmconfigCR.ObjectMeta.Labels[ndm.NDMHostKey] = fakeHostName
	return ndmconfigCR
}

func fakePodObject() *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:    make(map[string]string),
			Name:      podName,
			Namespace: "default",
		},
	}
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

	deviceList := &openebsv1alpha1.DeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Device",
			APIVersion: "",
		},
	}

	ndmconfigR := GetFakeNdmConfigObject()
	podR := fakePodObject()

	diskObjs := []runtime.Object{diskR}
	s := scheme.Scheme

	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, diskR)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, diskList)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, deviceR)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, deviceList)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, ndmconfigR)
	s.AddKnownTypes(corev1.SchemeGroupVersion, podR)

	fakeNdmClient := fake.NewFakeClient(diskObjs...)
	if fakeNdmClient == nil {
		fmt.Println("NDMClient is not created")
	}
	err := fakeNdmClient.Create(context.TODO(), deviceR)
	if err != nil {
		fmt.Println("Device object is not created")
	}

	err = fakeNdmClient.Create(context.TODO(), ndmconfigR)
	if err != nil {
		fmt.Println("ndmconfigR object is not created")
	}

	err = fakeNdmClient.Create(context.TODO(), podR)
	if err != nil {
		fmt.Println("podR object is not created")
	}
	return fakeNdmClient, s
}
