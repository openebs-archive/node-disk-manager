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
	deviceRequestUID  types.UID = "deviceRequest-example-UID"
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
	r.CheckDeviceRequestStatus(t, req, openebsv1alpha1.DeviceRequestStatusEmpty)

	// Fetch the DeviceRequest CR and change capacity to invalid
	// Since Capacity is invalid, it delete deviceRequest CR
	r.InvalidCapacityTest(t, req)

	// Create new DeviceRequest CR with right capacity,
	// trigger reconilation event. This time, it should
	// bound.
	deviceRequestR := GetFakeDeviceRequestObject()
	err := r.client.Create(context.TODO(), deviceRequestR)
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
	r.CheckDeviceRequestStatus(t, req, openebsv1alpha1.DeviceRequestStatusDone)

	r.DeviceRequestedHappyPathTest(t, req)
	//TODO: Need to find a way to update deletion timestamp
	//r.DeleteDeviceRequestedTest(t, req)
}

func (r *ReconcileDeviceRequest) DeleteDeviceRequestedTest(t *testing.T,
	req reconcile.Request) {

	devRequestInst := &openebsv1alpha1.DeviceRequest{}

	// Fetch the DeviceRequest CR
	err := r.client.Get(context.TODO(), req.NamespacedName, devRequestInst)
	if err != nil {
		t.Errorf("Get devRequestInst: (%v)", err)
	}

	err = r.client.Delete(context.TODO(), devRequestInst)
	if err != nil {
		t.Errorf("Delete devRequestInst: (%v)", err)
	}

	res, err := r.Reconcile(req)
	if err != nil {
		t.Logf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Log("reconcile did not requeue request as expected")
	}

	dvRequestInst := &openebsv1alpha1.DeviceRequest{}
	err = r.client.Get(context.TODO(), req.NamespacedName, dvRequestInst)
	if errors.IsNotFound(err) {
		t.Logf("DeviceRequest is deleted, expected")
		err = nil
	} else if err != nil {
		t.Errorf("Get dvRequestInst: (%v)", err)
	}

	time.Sleep(10 * time.Second)
	// Fetch the Device CR
	devInst := &openebsv1alpha1.Device{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: deviceName, Namespace: namespace}, devInst)
	if err != nil {
		t.Errorf("get devInst: (%v)", err)
	}

	if devInst.ClaimRef.UID == dvRequestInst.ObjectMeta.UID {
		t.Logf("Device ObjRef UID:%v match expected deviceRequest UID:%v",
			devInst.ClaimRef.UID, dvRequestInst.ObjectMeta.UID)
	} else {
		t.Fatalf("Device ObjRef UID:%v did not match expected deviceRequest UID:%v",
			devInst.ClaimRef.UID, dvRequestInst.ObjectMeta.UID)
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

	devRequestInst := &openebsv1alpha1.DeviceRequest{}
	// Fetch the DeviceRequest CR
	err := r.client.Get(context.TODO(), req.NamespacedName, devRequestInst)
	if err != nil {
		t.Errorf("Get devRequestInst: (%v)", err)
	}

	// Fetch the Device CR
	devInst := &openebsv1alpha1.Device{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: deviceName, Namespace: namespace}, devInst)
	if err != nil {
		t.Errorf("get devInst: (%v)", err)
	}

	if devInst.ClaimRef.UID == devRequestInst.ObjectMeta.UID {
		t.Logf("Device ObjRef UID:%v match expected deviceRequest UID:%v",
			devInst.ClaimRef.UID, devRequestInst.ObjectMeta.UID)
	} else {
		t.Fatalf("Device ObjRef UID:%v did not match expected deviceRequest UID:%v",
			devInst.ClaimRef.UID, devRequestInst.ObjectMeta.UID)
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

	devRequestInst := &openebsv1alpha1.DeviceRequest{}
	err := r.client.Get(context.TODO(), req.NamespacedName, devRequestInst)
	if err != nil {
		t.Errorf("Get devRequestInst: (%v)", err)
	}

	devRequestInst.Spec.Capacity = 0
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

	dvC := &openebsv1alpha1.DeviceRequest{}
	err = r.client.Get(context.TODO(), req.NamespacedName, dvC)
	if errors.IsNotFound(err) {
		t.Logf("DeviceRequest is deleted, expected")
		err = nil
	} else if err != nil {
		t.Errorf("Get devRequestInst: (%v)", err)
	}
}

func (r *ReconcileDeviceRequest) CheckDeviceRequestStatus(t *testing.T,
	req reconcile.Request, phase openebsv1alpha1.DeviceRequestPhase) {

	devRequestCR := &openebsv1alpha1.DeviceRequest{}
	err := r.client.Get(context.TODO(), req.NamespacedName, devRequestCR)
	if err != nil {
		t.Errorf("get devRequestCR : (%v)", err)
	}

	// DeviceRequest should yet to bound.
	if devRequestCR.Status.Phase == phase {
		t.Logf("DeviceRequest Object status:%v match expected status:%v",
			devRequestCR.Status.Phase, phase)
	} else {
		t.Fatalf("DeviceRequest Object status:%v did not match expected status:%v",
			devRequestCR.Status.Phase, phase)
	}
}

func GetFakeDeviceRequestObject() *openebsv1alpha1.DeviceRequest {
	deviceRequestCR := &openebsv1alpha1.DeviceRequest{}

	TypeMeta := metav1.TypeMeta{
		Kind:       "DeviceRequest",
		APIVersion: ndm.NDMVersion,
	}

	ObjectMeta := metav1.ObjectMeta{
		Labels:    make(map[string]string),
		Name:      devicerequestName,
		Namespace: namespace,
		UID:       deviceRequestUID,
	}

	Spec := openebsv1alpha1.DeviceRequestSpec{
		Capacity:   capacity, // Set deviceclaim size
		DeviceType: "",
		HostName:   fakeHostName,
	}

	deviceRequestCR.ObjectMeta = ObjectMeta
	deviceRequestCR.TypeMeta = TypeMeta
	deviceRequestCR.Spec = Spec
	deviceRequestCR.Status.Phase = openebsv1alpha1.DeviceRequestStatusEmpty
	return deviceRequestCR
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

	deviceRequestCR := GetFakeDeviceRequestObject()
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
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, deviceRequestCR)
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
	err = fakeNdmClient.Create(context.TODO(), deviceRequestCR)
	if err != nil {
		fmt.Println("DeviceRequest object is not created")
	}
	return fakeNdmClient, s
}
