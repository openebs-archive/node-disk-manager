package disk

import (
	"context"
	"testing"

	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	openebsv1alpha1 "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	//appsv1 "k8s.io/api/apps/v1"
	//corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	fakeHostName        = "fake-hostname"
	name                = "disk-example"
	namespace           = ""
	capacity     uint64 = 1024000
)

// TestDiskController runs ReconcileDisk.Reconcile() against a
// fake client that tracks a Disk object.
func TestDiskController(t *testing.T) {

	// Set the logger to development mode for verbose logs.
	logf.SetLogger(logf.ZapLogger(true))

	// Get Disk and diskList resource
	disk := GetFakeDiskObject()
	diskList := GetFakeDiskList()

	// Objects to track in the fake client.
	objs := []runtime.Object{
		disk,
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, disk)
	s.AddKnownTypes(openebsv1alpha1.SchemeGroupVersion, diskList)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	// Create a ReconcileDisk object with the scheme and fake client.
	r := &ReconcileDisk{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
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

	diskInstance := &openebsv1alpha1.Disk{}
	err = r.client.Get(context.TODO(), req.NamespacedName, diskInstance)
	if err != nil {
		t.Errorf("get diskInstance : (%v)", err)
	}

	// Disk Status state should be Active as expected.
	if diskInstance.Status.State == ndm.NDMActive {
		t.Logf("Disk Object state:%v match expected state:%v", diskInstance.Status.State, ndm.NDMActive)
	} else {
		t.Fatalf("Disk Object state:%v did not match expected state:%v", diskInstance.Status.State, ndm.NDMActive)
	}
}

func GetFakeDiskList() *openebsv1alpha1.DiskList {

	diskList := &openebsv1alpha1.DiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Disk",
			APIVersion: "",
		},
	}
	return diskList
}

func GetFakeDiskObject() *openebsv1alpha1.Disk {

	disk := &openebsv1alpha1.Disk{
		TypeMeta: metav1.TypeMeta{
			Kind:       ndm.NDMDiskKind,
			APIVersion: ndm.NDMVersion,
		},

		ObjectMeta: metav1.ObjectMeta{
			Labels:    make(map[string]string),
			Name:      name,
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
