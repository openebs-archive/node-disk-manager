package deviceclaim

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
	fakeHostName              = "fake-hostname"
	diskName                  = "disk-example"
	deviceName                = "device-example"
	deviceclaimName           = "deviceclaim-example"
	deviceClaimUID  types.UID = "deviceclaim-example-UID"
	namespace                 = ""
	capacity        uint64    = 1024000
)

// TestDeviceClaimController runs ReconcileDisk.Reconcile() against a
// fake client that tracks a DeviceClaim object.
// Test description:
func TestDeviceClaimController(t *testing.T) {

	// Set the logger to development mode for verbose logs.
	logf.SetLogger(logf.ZapLogger(true))

	// Create a fake client to mock API calls.
	cl, s := CreateFakeClient(t)

	// Create a ReconcileDevice object with the scheme and fake client.
	r := &ReconcileDeviceClaim{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      deviceclaimName,
			Namespace: namespace,
		},
	}

	// Check status of deviceClaim it should not empty(Not bound)
	r.CheckDeviceClaimStatus(t, req, openebsv1alpha1.DeviceClaimStatusEmpty)

	// Fetch the DeviceClaim CR and change capacity to invalid
	// Since Capacity is invalid, it delete deviceClaim CR
	r.InvalidCapacityTest(t, req)

	// Create new DeviceClaim CR with right capacity,
	// trigger reconilation event. This time, it should
	// bound.
	deviceClaimR := GetFakeDeviceClaimObject()
	err := r.client.Create(context.TODO(), deviceClaimR)
	if err != nil {
		t.Errorf("DeviceClaim object is not created")
	}

	res, err := r.Reconcile(req)
	if err != nil {
		t.Logf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Log("reconcile did not requeue request as expected")
	}
	r.CheckDeviceClaimStatus(t, req, openebsv1alpha1.DeviceClaimStatusDone)

	r.DeviceClaimedHappyPathTest(t, req)
	//r.DeleteDeviceClaimedTest(t, req)
}

func (r *ReconcileDeviceClaim) DeleteDeviceClaimedTest(t *testing.T,
	req reconcile.Request) {

	devClaimInst := &openebsv1alpha1.DeviceClaim{}
	// Fetch the DeviceClaim CR
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

	dvClaimInst := &openebsv1alpha1.DeviceClaim{}
	err = r.client.Get(context.TODO(), req.NamespacedName, dvClaimInst)
	if errors.IsNotFound(err) {
		t.Logf("DeviceClaim is deleted, expected")
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

func (r *ReconcileDeviceClaim) DeviceClaimedHappyPathTest(t *testing.T,
	req reconcile.Request) {

	devClaimInst := &openebsv1alpha1.DeviceClaim{}
	// Fetch the DeviceClaim CR
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

func (r *ReconcileDeviceClaim) InvalidCapacityTest(t *testing.T,
	req reconcile.Request) {

	devClaimInst := &openebsv1alpha1.DeviceClaim{}
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

	dvC := &openebsv1alpha1.DeviceClaim{}
	err = r.client.Get(context.TODO(), req.NamespacedName, dvC)
	if errors.IsNotFound(err) {
		t.Logf("DeviceClaim is deleted, expected")
		err = nil
	} else if err != nil {
		t.Errorf("Get devClaimInst: (%v)", err)
	}
}

func (r *ReconcileDeviceClaim) CheckDeviceClaimStatus(t *testing.T,
	req reconcile.Request, phase openebsv1alpha1.DeviceClaimPhase) {

	devClaimR := &openebsv1alpha1.DeviceClaim{}
	err := r.client.Get(context.TODO(), req.NamespacedName, devClaimR)
	if err != nil {
		t.Errorf("get devClaimR : (%v)", err)
	}

	// DeviceClaim should yet to bound.
	if devClaimR.Status.Phase == phase {
		t.Logf("DeviceClaim Object status:%v match expected status:%v",
			devClaimR.Status.Phase, phase)
	} else {
		t.Fatalf("DeviceClaim Object status:%v did not match expected status:%v",
			devClaimR.Status.Phase, phase)
	}
}

func GetFakeDeviceClaimObject() *openebsv1alpha1.DeviceClaim {
	deviceClaim := &openebsv1alpha1.DeviceClaim{}

	TypeMeta := metav1.TypeMeta{
		Kind:       "DeviceClaim",
		APIVersion: ndm.NDMVersion,
	}

	ObjectMeta := metav1.ObjectMeta{
		Labels:    make(map[string]string),
		Name:      deviceclaimName,
		Namespace: namespace,
		UID:       deviceClaimUID,
	}

	Spec := openebsv1alpha1.DeviceClaimSpec{
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

	deviceClaimR := GetFakeDeviceClaimObject()
	deviceclaimList := &openebsv1alpha1.DeviceClaimList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeviceClaim",
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
		fmt.Println("DeviceClaim object is not created")
	}
	return fakeNdmClient, s
}
