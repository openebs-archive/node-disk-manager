package devicerequest

import (
	"context"
	"fmt"
	"testing"
	"time"

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
	fakeHostName                = "fake-hostname"
	diskName                    = "disk-example"
	deviceName                  = "device-example"
	devicerequestName           = "devicerequest-example"
	deviceClaimUID    types.UID = "deviceRequest-example-UID"
	namespace                   = ""
	capacity          uint64    = 1024000
)

// TestDeviceRequestController runs ReconcileDeviceRequest.Reconcile() against a
// fake client that tracks a DeviceRequest object.
// Test description:
func TestDeviceRequestController(t *testing.T) {

	// Set the logger to development mode for verbose logs.
	logf.SetLogger(logf.ZapLogger(true))

	// Create a fake client to mock API calls.
	cl, s := CreateFakeClient(t)

	// Create a ReconcileDevice object with the scheme and fake client.
	r := &ReconcileDeviceRequest{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      devicerequestName,
			Namespace: namespace,
		},
	}

	// Check status of deviceRequest it should not be empty(Not bound)
	r.CheckDeviceRequestStatus(t, req, openebsv1alpha1.DeviceClaimStatusEmpty)

	// Fetch the DeviceRequest CR and change capacity to invalid
	// Since Capacity is invalid, it delete deviceClaim CR
	r.InvalidCapacityTest(t, req)

	// Create new DeviceRequest CR with right capacity,
	// trigger reconilation event. This time, it should
	// bound.
	deviceClaimR := GetFakeDeviceRequestObject()
	err := r.client.Create(context.TODO(), deviceClaimR)
	if err != nil {
		t.Errorf("DeviceRequest object is not created")
	}

	res, err := r.Reconcile(req)
	if err != nil {
		t.Logf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Log("reconcile did not requeue request as expected")
	}
	r.CheckDeviceRequestStatus(t, req, openebsv1alpha1.DeviceClaimStatusDone)

	r.DeviceRequestedHappyPathTest(t, req)
	//TODO: Need to find a way to update deletion timestamp
	//r.DeleteDeviceRequestedTest(t, req)
}

func (r *ReconcileDeviceRequest) DeleteDeviceRequestedTest(t *testing.T,
	req reconcile.Request) {

	devClaimInst := &openebsv1alpha1.DeviceRequest{}

	// Fetch the DeviceRequest CR
	err := r.client.Get(context.TODO(), req.NamespacedName, devClaimInst)
	if err != nil {
		t.Errorf("Get devClaimInst: (%v)", err)
	}

	err = r.client.Delete(context.TODO(), devClaimInst)
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

	dvClaimInst := &openebsv1alpha1.DeviceRequest{}
	err = r.client.Get(context.TODO(), req.NamespacedName, dvClaimInst)
	if errors.IsNotFound(err) {
		t.Logf("DeviceRequest is deleted, expected")
		err = nil
	} else if err != nil {
		t.Errorf("Get dvClaimInst: (%v)", err)
	}

	time.Sleep(10 * time.Second)
	// Fetch the Device CR
	devInst := &openebsv1alpha1.Device{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: deviceName, Namespace: namespace}, devInst)
	if err != nil {
		t.Errorf("get devInst: (%v)", err)
	}

	if devInst.Claim.DeviceClaimUID == dvClaimInst.ObjectMeta.UID {
		t.Logf("Device Obj ClaimUID:%v match expected ClaimUID:%v",
			devInst.Claim.DeviceClaimUID, dvClaimInst.ObjectMeta.UID)
	} else {
		t.Fatalf("Device Obj ClaimUID:%v did not match expected ClaimUID:%v",
			devInst.Claim.DeviceClaimUID, dvClaimInst.ObjectMeta.UID)
	}

	if devInst.ClaimState.State == ndm.NDMClaimed {
		t.Logf("Device Obj state:%v match expected state:%v",
			devInst.ClaimState.State, ndm.NDMClaimed)
	} else {
		t.Fatalf("Device Obj state:%v did not match expected state:%v",
			devInst.ClaimState.State, ndm.NDMClaimed)
	}
}

func (r *ReconcileDeviceRequest) DeviceRequestedHappyPathTest(t *testing.T,
	req reconcile.Request) {

	devClaimInst := &openebsv1alpha1.DeviceRequest{}
	// Fetch the DeviceRequest CR
	err := r.client.Get(context.TODO(), req.NamespacedName, devClaimInst)
	if err != nil {
		t.Errorf("Get devClaimInst: (%v)", err)
	}

	// Fetch the Device CR
	devInst := &openebsv1alpha1.Device{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: deviceName, Namespace: namespace}, devInst)
	if err != nil {
		t.Errorf("get devInst: (%v)", err)
	}

	if devInst.Claim.DeviceClaimUID == devClaimInst.ObjectMeta.UID {
		t.Logf("Device Obj ClaimUID:%v match expected ClaimUID:%v",
			devInst.Claim.DeviceClaimUID, devClaimInst.ObjectMeta.UID)
	} else {
		t.Fatalf("Device Obj ClaimUID:%v did not match expected ClaimUID:%v",
			devInst.Claim.DeviceClaimUID, devClaimInst.ObjectMeta.UID)
	}

	if devInst.ClaimState.State == ndm.NDMClaimed {
		t.Logf("Device Obj state:%v match expected state:%v",
			devInst.ClaimState.State, ndm.NDMClaimed)
	} else {
		t.Fatalf("Device Obj state:%v did not match expected state:%v",
			devInst.ClaimState.State, ndm.NDMClaimed)
	}
}

func (r *ReconcileDeviceRequest) InvalidCapacityTest(t *testing.T,
	req reconcile.Request) {

	devClaimInst := &openebsv1alpha1.DeviceRequest{}
	err := r.client.Get(context.TODO(), req.NamespacedName, devClaimInst)
	if err != nil {
		t.Errorf("Get devClaimInst: (%v)", err)
	}

	devClaimInst.Spec.Capacity = 0
	err = r.client.Update(context.TODO(), devClaimInst)
	if err != nil {
		t.Errorf("Update devClaimInst: (%v)", err)
	}

	res, err := r.Reconcile(req)
	if err != nil {
		t.Logf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Log("reconcile did not requeue request as expected")
	}

	dvC := &openebsv1alpha1.DeviceRequest{}
	err = r.client.Get(context.TODO(), req.NamespacedName, dvC)
	if errors.IsNotFound(err) {
		t.Logf("DeviceRequest is deleted, expected")
		err = nil
	} else if err != nil {
		t.Errorf("Get devClaimInst: (%v)", err)
	}
}

func (r *ReconcileDeviceRequest) CheckDeviceRequestStatus(t *testing.T,
	req reconcile.Request, phase openebsv1alpha1.DeviceClaimPhase) {

	devClaimR := &openebsv1alpha1.DeviceRequest{}
	err := r.client.Get(context.TODO(), req.NamespacedName, devClaimR)
	if err != nil {
		t.Errorf("get devClaimR : (%v)", err)
	}

	// DeviceRequest should yet to bound.
	if devClaimR.Status.Phase == phase {
		t.Logf("DeviceRequest Object status:%v match expected status:%v",
			devClaimR.Status.Phase, phase)
	} else {
		t.Fatalf("DeviceRequest Object status:%v did not match expected status:%v",
			devClaimR.Status.Phase, phase)
	}
}

func GetFakeDeviceRequestObject() *openebsv1alpha1.DeviceRequest {
	deviceClaim := &openebsv1alpha1.DeviceRequest{}

	TypeMeta := metav1.TypeMeta{
		Kind:       "DeviceRequest",
		APIVersion: ndm.NDMVersion,
	}

	ObjectMeta := metav1.ObjectMeta{
		Labels:    make(map[string]string),
		Name:      devicerequestName,
		Namespace: namespace,
		UID:       deviceClaimUID,
	}

	Spec := openebsv1alpha1.DeviceRequestSpec{
		Capacity:  capacity, // Set deviceclaim size
		DriveType: "",
		HostName:  fakeHostName,
	}

	deviceClaim.ObjectMeta = ObjectMeta
	deviceClaim.TypeMeta = TypeMeta
	deviceClaim.Spec = Spec
	deviceClaim.Status.Phase = openebsv1alpha1.DeviceClaimStatusEmpty
	return deviceClaim
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

	deviceR := GetFakeDeviceObject()

	deviceList := &openebsv1alpha1.DeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Device",
			APIVersion: "",
		},
	}

	deviceClaimR := GetFakeDeviceRequestObject()
	deviceclaimList := &openebsv1alpha1.DeviceRequestList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeviceRequest",
			APIVersion: "",
		},
	}

	diskObjs := []runtime.Object{diskR}
	s := scheme.Scheme

	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, diskR)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, diskList)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, deviceR)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, deviceList)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, deviceClaimR)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, deviceclaimList)

	fakeNdmClient := fake.NewFakeClient(diskObjs...)
	if fakeNdmClient == nil {
		fmt.Println("NDMClient is not created")
	}

	// Create a new device obj
	err := fakeNdmClient.Create(context.TODO(), deviceR)
	if err != nil {
		fmt.Println("Device object is not created")
	}

	// Create a new deviceclaim obj
	err = fakeNdmClient.Create(context.TODO(), deviceClaimR)
	if err != nil {
		fmt.Println("DeviceRequest object is not created")
	}
	return fakeNdmClient, s
}
