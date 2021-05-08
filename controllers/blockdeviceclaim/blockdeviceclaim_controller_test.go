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

	"github.com/openebs/node-disk-manager/db/kubernetes"

	openebsv1alpha1 "github.com/openebs/node-disk-manager/api/v1alpha1"
	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
	fakeRecorder                   = record.NewFakeRecorder(50)
)

// TestBlockDeviceClaimController runs ReconcileBlockDeviceClaim.Reconcile() against a
// fake client that tracks a BlockDeviceClaim object.
func TestBlockDeviceClaimController(t *testing.T) {

	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	// Create a fake client to mock API calls.
	cl, s := CreateFakeClient()
	deviceR := GetFakeDeviceObject(deviceName, capacity)
	deviceR.Labels[kubernetes.KubernetesHostNameLabel] = fakeHostName

	deviceClaimR := GetFakeBlockDeviceClaimObject()
	// Create a new blockdevice obj
	err := cl.Create(context.TODO(), deviceR)
	if err != nil {
		fmt.Println("BlockDevice object is not created", err)
	}

	// Create a new deviceclaim obj
	err = cl.Create(context.TODO(), deviceClaimR)
	if err != nil {
		fmt.Println("BlockDeviceClaim object is not created", err)
	}

	// Create a ReconcileDevice object with the scheme and fake client.
	r := &BlockDeviceClaimReconciler{Client: cl, Scheme: s, recorder: fakeRecorder}

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
	// trigger reconciliation event. This time, it should
	// bound.
	deviceClaim := &openebsv1alpha1.BlockDeviceClaim{}
	err = r.Client.Get(context.TODO(), req.NamespacedName, deviceClaim)
	if err != nil {
		t.Errorf("Get deviceClaim: (%v)", err)
	}
	orignalDeviceClaim := deviceClaim.DeepCopy()
	deviceClaim.Spec.Resources.Requests[openebsv1alpha1.ResourceStorage] = claimCapacity
	// resetting status to empty
	deviceClaim.Status.Phase = openebsv1alpha1.BlockDeviceClaimStatusEmpty
	err = r.Client.Patch(context.TODO(), deviceClaim, client.MergeFrom(orignalDeviceClaim))
	if err != nil {
		t.Errorf("Update deviceClaim: (%v)", err)
	}

	res, err := r.Reconcile(context.TODO(), req)
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

func (r *BlockDeviceClaimReconciler) DeleteBlockDeviceClaimedTest(t *testing.T,
	req reconcile.Request) {

	devRequestInst := &openebsv1alpha1.BlockDeviceClaim{}

	// Fetch the BlockDeviceClaim CR
	err := r.Client.Get(context.TODO(), req.NamespacedName, devRequestInst)
	if err != nil {
		t.Errorf("Get devClaimInst: (%v)", err)
	}

	err = r.Client.Delete(context.TODO(), devRequestInst)
	if err != nil {
		t.Errorf("Delete devClaimInst: (%v)", err)
	}

	res, err := r.Reconcile(context.TODO(), req)
	if err != nil {
		t.Logf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Log("reconcile did not requeue request as expected")
	}

	dvRequestInst := &openebsv1alpha1.BlockDeviceClaim{}
	err = r.Client.Get(context.TODO(), req.NamespacedName, dvRequestInst)
	if errors.IsNotFound(err) {
		t.Logf("BlockDeviceClaim is deleted, expected")
		err = nil
	} else if err != nil {
		t.Errorf("Get dvClaimInst: (%v)", err)
	}

	time.Sleep(10 * time.Second)
	// Fetch the BlockDevice CR
	devInst := &openebsv1alpha1.BlockDevice{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: deviceName, Namespace: namespace}, devInst)
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

func (r *BlockDeviceClaimReconciler) DeviceRequestedHappyPathTest(t *testing.T,
	req reconcile.Request) {

	devRequestInst := &openebsv1alpha1.BlockDeviceClaim{}
	// Fetch the BlockDeviceClaim CR
	err := r.Client.Get(context.TODO(), req.NamespacedName, devRequestInst)
	if err != nil {
		t.Errorf("Get devRequestInst: (%v)", err)
	}

	// Fetch the BlockDevice CR
	devInst := &openebsv1alpha1.BlockDevice{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: deviceName, Namespace: namespace}, devInst)
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

func (r *BlockDeviceClaimReconciler) InvalidCapacityTest(t *testing.T,
	req reconcile.Request) {

	devRequestInst := &openebsv1alpha1.BlockDeviceClaim{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, devRequestInst)
	if err != nil {
		t.Errorf("Get devRequestInst: (%v)", err)
	}

	orignalDevRequestInst := devRequestInst.DeepCopy()
	devRequestInst.Spec.Resources.Requests[openebsv1alpha1.ResourceStorage] = resource.MustParse("0")
	err = r.Client.Patch(context.TODO(), devRequestInst, client.MergeFrom(orignalDevRequestInst))
	if err != nil {
		t.Errorf("Update devRequestInst: (%v)", err)
	}

	res, err := r.Reconcile(context.TODO(), req)
	if err != nil {
		t.Logf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Log("reconcile did not requeue request as expected")
	}

	dvC := &openebsv1alpha1.BlockDeviceClaim{}
	err = r.Client.Get(context.TODO(), req.NamespacedName, dvC)
	if err != nil {
		t.Errorf("Get devRequestInst: (%v)", err)
	}
	r.CheckBlockDeviceClaimStatus(t, req, openebsv1alpha1.BlockDeviceClaimStatusPending)
}

func TestBlockDeviceClaimsLabelSelector(t *testing.T) {
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	tests := map[string]struct {
		bdLabels           map[string]string
		selector           *metav1.LabelSelector
		expectedClaimPhase openebsv1alpha1.DeviceClaimPhase
	}{
		"only hostname label is present and no selector": {
			bdLabels: map[string]string{
				kubernetes.KubernetesHostNameLabel: fakeHostName,
			},
			selector:           nil,
			expectedClaimPhase: openebsv1alpha1.BlockDeviceClaimStatusDone,
		},
		"custom label and hostname present on bd and selector": {
			bdLabels: map[string]string{
				ndm.KubernetesHostNameLabel: fakeHostName,
				"ndm.io/test":               "1234",
			},
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					ndm.KubernetesHostNameLabel: fakeHostName,
					"ndm.io/test":               "1234",
				},
			},
			expectedClaimPhase: openebsv1alpha1.BlockDeviceClaimStatusDone,
		},
		"custom labels and hostname": {
			bdLabels: map[string]string{
				ndm.KubernetesHostNameLabel: fakeHostName,
				"ndm.io/test":               "1234",
			},
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"ndm.io/test": "1234",
				},
			},
			expectedClaimPhase: openebsv1alpha1.BlockDeviceClaimStatusDone,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// pinning the variables
			bdLabels := test.bdLabels
			selector := test.selector
			expectedClaimPhase := test.expectedClaimPhase

			// Create a fake client to mock API calls.
			cl, s := CreateFakeClient()

			// Create a ReconcileDevice object with the scheme and fake client.
			r := &BlockDeviceClaimReconciler{Client: cl, Scheme: s, recorder: fakeRecorder}

			// Mock request to simulate Reconcile() being called on an event for a
			// watched resource .
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      blockDeviceClaimName,
					Namespace: namespace,
				},
			}

			bdcR := GetFakeBlockDeviceClaimObject()
			if err := cl.Create(context.TODO(), bdcR); err != nil {
				t.Fatal(err)
			}

			bd := GetFakeDeviceObject("bd-1", capacity*10)

			bd.Labels = bdLabels
			// mark the device as unclaimed
			bd.Spec.ClaimRef = nil
			bd.Status.ClaimState = openebsv1alpha1.BlockDeviceUnclaimed

			err := cl.Create(context.TODO(), bd)
			if err != nil {
				t.Fatalf("error updating BD. %v", err)
			}
			bdc := GetFakeBlockDeviceClaimObject()
			orignalBDC := bdc.DeepCopy()
			bdc.Spec.BlockDeviceName = ""
			bdc.Spec.Selector = selector
			err = cl.Patch(context.TODO(), bdc, client.MergeFrom(orignalBDC))
			if err != nil {
				t.Fatalf("error updating BDC. %v", err)
			}
			_, _ = r.Reconcile(context.TODO(), req)

			err = cl.Get(context.TODO(), req.NamespacedName, bdc)
			if err != nil {
				t.Fatalf("error getting BDC. %v", err)
			}

			err = cl.Delete(context.TODO(), bd)
			if err != nil {
				t.Fatalf("error deleting BDC. %v", err)
			}
			assert.Equal(t, expectedClaimPhase, bdc.Status.Phase)
		})
	}
}

func (r *BlockDeviceClaimReconciler) CheckBlockDeviceClaimStatus(t *testing.T,
	req reconcile.Request, phase openebsv1alpha1.DeviceClaimPhase) {

	devRequestCR := &openebsv1alpha1.BlockDeviceClaim{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, devRequestCR)
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

func CreateFakeClient() (client.Client, *runtime.Scheme) {

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

	s := scheme.Scheme

	s.AddKnownTypes(openebsv1alpha1.GroupVersion, deviceR)
	s.AddKnownTypes(openebsv1alpha1.GroupVersion, deviceList)
	s.AddKnownTypes(openebsv1alpha1.GroupVersion, deviceClaimR)
	s.AddKnownTypes(openebsv1alpha1.GroupVersion, deviceclaimList)

	fakeNdmClient := fake.NewFakeClientWithScheme(s)
	if fakeNdmClient == nil {
		fmt.Println("NDMClient is not created")
	}

	return fakeNdmClient, s
}

func TestGenerateSelector(t *testing.T) {
	tests := map[string]struct {
		bdc  openebsv1alpha1.BlockDeviceClaim
		want *metav1.LabelSelector
	}{
		"hostname/node attributes not given and no selector": {
			bdc: openebsv1alpha1.BlockDeviceClaim{
				Spec: openebsv1alpha1.DeviceClaimSpec{},
			},
			want: &metav1.LabelSelector{
				MatchLabels: make(map[string]string),
			},
		},
		"hostname is given, node attributes not given and no selector": {
			bdc: openebsv1alpha1.BlockDeviceClaim{
				Spec: openebsv1alpha1.DeviceClaimSpec{
					HostName: "hostname",
				},
			},
			want: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					ndm.KubernetesHostNameLabel: "hostname",
				},
			},
		},
		"hostname is not given, node attribute is given and no selector": {
			bdc: openebsv1alpha1.BlockDeviceClaim{
				Spec: openebsv1alpha1.DeviceClaimSpec{
					BlockDeviceNodeAttributes: openebsv1alpha1.BlockDeviceNodeAttributes{
						HostName: "hostname",
					},
				},
			},
			want: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					ndm.KubernetesHostNameLabel: "hostname",
				},
			},
		},
		"same hostname, node attribute is given and no selector": {
			bdc: openebsv1alpha1.BlockDeviceClaim{
				Spec: openebsv1alpha1.DeviceClaimSpec{
					HostName: "hostname",
					BlockDeviceNodeAttributes: openebsv1alpha1.BlockDeviceNodeAttributes{
						HostName: "hostname",
					},
				},
			},
			want: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					ndm.KubernetesHostNameLabel: "hostname",
				},
			},
		},
		"different hostname and node attributes is given and no selector": {
			bdc: openebsv1alpha1.BlockDeviceClaim{
				Spec: openebsv1alpha1.DeviceClaimSpec{
					HostName: "hostname1",
					BlockDeviceNodeAttributes: openebsv1alpha1.BlockDeviceNodeAttributes{
						HostName: "hostname2",
					},
				},
			},
			want: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					ndm.KubernetesHostNameLabel: "hostname2",
				},
			},
		},
		"no hostname and custom selector is given": {
			bdc: openebsv1alpha1.BlockDeviceClaim{
				Spec: openebsv1alpha1.DeviceClaimSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"ndm.io/test": "test",
						},
					},
				},
			},
			want: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"ndm.io/test": "test",
				},
			},
		},
		"hostname given and selector also contains custom label name": {
			bdc: openebsv1alpha1.BlockDeviceClaim{
				Spec: openebsv1alpha1.DeviceClaimSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							ndm.KubernetesHostNameLabel: "hostname1",
							"ndm.io/test":               "test",
						},
					},
					HostName: "hostname2",
				},
			},
			want: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					ndm.KubernetesHostNameLabel: "hostname2",
					"ndm.io/test":               "test",
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := generateSelector(test.bdc)
			assert.Equal(t, test.want, got)
		})
	}
}
