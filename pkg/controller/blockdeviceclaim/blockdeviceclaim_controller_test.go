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

package blockdeviceclaim

import (
	"context"
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	openebsv1alpha1 "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
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
	fakeHostName                   = "fake-hostname"
	diskName                       = "disk-example"
	deviceName                     = "blockdevice-example"
	blockDeviceClaimName           = "blockdeviceclaim-example"
	blockDeviceClaimUID  types.UID = "blockDeviceClaim-example-UID"
	namespace                      = ""
	capacity             uint64    = 1024000
	claimCapacity                  = resource.MustParse("1024000")
)

// TestBlockDeviceClaimController runs ReconcileBlockDeviceClaim.Reconcile() against a
// fake client that tracks a BlockDeviceClaim object.
func TestBlockDeviceClaimController(t *testing.T) {

	// Set the logger to development mode for verbose logs.
	logf.SetLogger(logf.ZapLogger(true))

	// Create a fake client to mock API calls.
	cl, s := CreateFakeClient(t)

	// Create a ReconcileDevice object with the scheme and fake client.
	r := &ReconcileBlockDeviceClaim{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      blockDeviceClaimName,
			Namespace: namespace,
		},
	}

	// Check status of deviceClaim it should be empty(Not bound)
	r.CheckBlockDeviceClaimStatus(t, req, openebsv1alpha1.BlockDeviceClaimStatusEmpty)

	// Fetch the BlockDeviceClaim CR and change capacity to invalid
	// Since Capacity is invalid, it delete device claim CR
	r.InvalidCapacityTest(t, req)

	// Create new BlockDeviceClaim CR with right capacity,
	// trigger reconcilation event. This time, it should
	// bound.
	deviceClaim := &openebsv1alpha1.BlockDeviceClaim{}
	err := r.client.Get(context.TODO(), req.NamespacedName, deviceClaim)
	if err != nil {
		t.Errorf("Get deviceClaim: (%v)", err)
	}

	deviceClaim.Spec.Resources.Requests[openebsv1alpha1.ResourceStorage] = claimCapacity
	// resetting status to empty
	deviceClaim.Status.Phase = openebsv1alpha1.BlockDeviceClaimStatusEmpty
	err = r.client.Update(context.TODO(), deviceClaim)
	if err != nil {
		t.Errorf("Update deviceClaim: (%v)", err)
	}

	res, err := r.Reconcile(req)
	if err != nil {
		t.Logf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Log("reconcile did not requeue request as expected")
	}
	r.CheckBlockDeviceClaimStatus(t, req, openebsv1alpha1.BlockDeviceClaimStatusDone)

	r.DeviceRequestedHappyPathTest(t, req)
	//TODO: Need to find a way to update deletion timestamp
	//r.DeleteBlockDeviceClaimedTest(t, req)
}

func (r *ReconcileBlockDeviceClaim) DeleteBlockDeviceClaimedTest(t *testing.T,
	req reconcile.Request) {

	devRequestInst := &openebsv1alpha1.BlockDeviceClaim{}

	// Fetch the BlockDeviceClaim CR
	err := r.client.Get(context.TODO(), req.NamespacedName, devRequestInst)
	if err != nil {
		t.Errorf("Get devClaimInst: (%v)", err)
	}

	err = r.client.Delete(context.TODO(), devRequestInst)
	if err != nil {
		t.Errorf("Delete devClaimInst: (%v)", err)
	}

	res, err := r.Reconcile(req)
	if err != nil {
		t.Logf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Log("reconcile did not requeue request as expected")
	}

	dvRequestInst := &openebsv1alpha1.BlockDeviceClaim{}
	err = r.client.Get(context.TODO(), req.NamespacedName, dvRequestInst)
	if errors.IsNotFound(err) {
		t.Logf("BlockDeviceClaim is deleted, expected")
		err = nil
	} else if err != nil {
		t.Errorf("Get dvClaimInst: (%v)", err)
	}

	time.Sleep(10 * time.Second)
	// Fetch the BlockDevice CR
	devInst := &openebsv1alpha1.BlockDevice{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: deviceName, Namespace: namespace}, devInst)
	if err != nil {
		t.Errorf("get devInst: (%v)", err)
	}

	if devInst.Spec.ClaimRef.UID == dvRequestInst.ObjectMeta.UID {
		t.Logf("BlockDevice ObjRef UID:%v match expected deviceRequest UID:%v",
			devInst.Spec.ClaimRef.UID, dvRequestInst.ObjectMeta.UID)
	} else {
		t.Fatalf("BlockDevice ObjRef UID:%v did not match expected deviceRequest UID:%v",
			devInst.Spec.ClaimRef.UID, dvRequestInst.ObjectMeta.UID)
	}

	if devInst.Status.ClaimState == openebsv1alpha1.BlockDeviceClaimed {
		t.Logf("BlockDevice Obj state:%v match expected state:%v",
			devInst.Status.ClaimState, openebsv1alpha1.BlockDeviceClaimed)
	} else {
		t.Fatalf("BlockDevice Obj state:%v did not match expected state:%v",
			devInst.Status.ClaimState, openebsv1alpha1.BlockDeviceClaimed)
	}
}

func (r *ReconcileBlockDeviceClaim) DeviceRequestedHappyPathTest(t *testing.T,
	req reconcile.Request) {

	devRequestInst := &openebsv1alpha1.BlockDeviceClaim{}
	// Fetch the BlockDeviceClaim CR
	err := r.client.Get(context.TODO(), req.NamespacedName, devRequestInst)
	if err != nil {
		t.Errorf("Get devRequestInst: (%v)", err)
	}

	// Fetch the BlockDevice CR
	devInst := &openebsv1alpha1.BlockDevice{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: deviceName, Namespace: namespace}, devInst)
	if err != nil {
		t.Errorf("get devInst: (%v)", err)
	}

	if devInst.Spec.ClaimRef.UID == devRequestInst.ObjectMeta.UID {
		t.Logf("BlockDevice ObjRef UID:%v match expected deviceRequest UID:%v",
			devInst.Spec.ClaimRef.UID, devRequestInst.ObjectMeta.UID)
	} else {
		t.Fatalf("BlockDevice ObjRef UID:%v did not match expected deviceRequest UID:%v",
			devInst.Spec.ClaimRef.UID, devRequestInst.ObjectMeta.UID)
	}

	if devInst.Status.ClaimState == openebsv1alpha1.BlockDeviceClaimed {
		t.Logf("BlockDevice Obj state:%v match expected state:%v",
			devInst.Status.ClaimState, openebsv1alpha1.BlockDeviceClaimed)
	} else {
		t.Fatalf("BlockDevice Obj state:%v did not match expected state:%v",
			devInst.Status.ClaimState, openebsv1alpha1.BlockDeviceClaimed)
	}
}

func (r *ReconcileBlockDeviceClaim) InvalidCapacityTest(t *testing.T,
	req reconcile.Request) {

	devRequestInst := &openebsv1alpha1.BlockDeviceClaim{}
	err := r.client.Get(context.TODO(), req.NamespacedName, devRequestInst)
	if err != nil {
		t.Errorf("Get devRequestInst: (%v)", err)
	}

	devRequestInst.Spec.Resources.Requests[openebsv1alpha1.ResourceStorage] = resource.MustParse("0")
	err = r.client.Update(context.TODO(), devRequestInst)
	if err != nil {
		t.Errorf("Update devRequestInst: (%v)", err)
	}

	res, err := r.Reconcile(req)
	if err != nil {
		t.Logf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Log("reconcile did not requeue request as expected")
	}

	dvC := &openebsv1alpha1.BlockDeviceClaim{}
	err = r.client.Get(context.TODO(), req.NamespacedName, dvC)
	if err != nil {
		t.Errorf("Get devRequestInst: (%v)", err)
	}
	r.CheckBlockDeviceClaimStatus(t, req, openebsv1alpha1.BlockDeviceClaimStatusInvalidCapacity)
}

func (r *ReconcileBlockDeviceClaim) CheckBlockDeviceClaimStatus(t *testing.T,
	req reconcile.Request, phase openebsv1alpha1.DeviceClaimPhase) {

	devRequestCR := &openebsv1alpha1.BlockDeviceClaim{}
	err := r.client.Get(context.TODO(), req.NamespacedName, devRequestCR)
	if err != nil {
		t.Errorf("get devRequestCR : (%v)", err)
	}

	// BlockDeviceClaim should yet to bound.
	if devRequestCR.Status.Phase == phase {
		t.Logf("BlockDeviceClaim Object status:%v match expected status:%v",
			devRequestCR.Status.Phase, phase)
	} else {
		t.Fatalf("BlockDeviceClaim Object status:%v did not match expected status:%v",
			devRequestCR.Status.Phase, phase)
	}
}

func GetFakeBlockDeviceClaimObject() *openebsv1alpha1.BlockDeviceClaim {
	deviceRequestCR := &openebsv1alpha1.BlockDeviceClaim{}

	TypeMeta := metav1.TypeMeta{
		Kind:       "BlockDeviceClaim",
		APIVersion: ndm.NDMVersion,
	}

	ObjectMeta := metav1.ObjectMeta{
		Labels:    make(map[string]string),
		Name:      blockDeviceClaimName,
		Namespace: namespace,
		UID:       blockDeviceClaimUID,
	}

	Requests := corev1.ResourceList{openebsv1alpha1.ResourceStorage: claimCapacity}

	Requirements := openebsv1alpha1.DeviceClaimResources{
		Requests: Requests,
	}

	Spec := openebsv1alpha1.DeviceClaimSpec{
		Resources:  Requirements,
		DeviceType: "",
		HostName:   fakeHostName,
	}

	deviceRequestCR.ObjectMeta = ObjectMeta
	deviceRequestCR.TypeMeta = TypeMeta
	deviceRequestCR.Spec = Spec
	deviceRequestCR.Status.Phase = openebsv1alpha1.BlockDeviceClaimStatusEmpty
	return deviceRequestCR
}

func GetFakeDeviceObject(bdName string, bdCapacity uint64) *openebsv1alpha1.BlockDevice {
	device := &openebsv1alpha1.BlockDevice{}

	TypeMeta := metav1.TypeMeta{
		Kind:       ndm.NDMBlockDeviceKind,
		APIVersion: ndm.NDMVersion,
	}

	ObjectMeta := metav1.ObjectMeta{
		Labels:    make(map[string]string),
		Name:      bdName,
		Namespace: namespace,
	}

	Spec := openebsv1alpha1.DeviceSpec{
		Path: "dev/disk-fake-path",
		Capacity: openebsv1alpha1.DeviceCapacity{
			Storage: bdCapacity, // Set blockdevice size.
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
	disk.ObjectMeta.Labels[ndm.NDMHostKey] = fakeHostName
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

	deviceR := GetFakeDeviceObject(deviceName, capacity)

	deviceList := &openebsv1alpha1.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "",
		},
	}

	deviceClaimR := GetFakeBlockDeviceClaimObject()
	deviceclaimList := &openebsv1alpha1.BlockDeviceClaimList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDeviceClaim",
			APIVersion: "",
		},
	}

	objects := []runtime.Object{diskR, deviceR, deviceClaimR}
	s := scheme.Scheme

	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, diskR)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, diskList)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, deviceR)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, deviceList)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, deviceClaimR)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, deviceclaimList)

	fakeNdmClient := fake.NewFakeClient(objects...)
	if fakeNdmClient == nil {
		fmt.Println("NDMClient is not created")
	}

	// Create a new blockdevice obj
	err := fakeNdmClient.Create(context.TODO(), deviceR)
	if err != nil {
		fmt.Println("BlockDevice object is not created", err)
	}

	// Create a new deviceclaim obj
	err = fakeNdmClient.Create(context.TODO(), deviceClaimR)
	if err != nil {
		fmt.Println("BlockDeviceClaim object is not created", err)
	}
	return fakeNdmClient, s
}
