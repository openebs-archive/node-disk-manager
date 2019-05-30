package blockdevice

import (
	"context"
	"github.com/go-logr/logr"
	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	openebsv1alpha1 "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/udev"
	"k8s.io/apimachinery/pkg/types"
	"strings"

	//corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_device")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new BlockDevice Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBlockDevice{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("blockdevice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource BlockDevice
	err = c.Watch(&source.Kind{Type: &openebsv1alpha1.BlockDevice{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileBlockDevice{}

// ReconcileBlockDevice reconciles a BlockDevice object
type ReconcileBlockDevice struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a BlockDevice object and makes changes based on the state read
// and what is in the BlockDevice.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBlockDevice) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling BlockDevice")

	// Fetch the BlockDevice instance
	instance := &openebsv1alpha1.BlockDevice{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Requested object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	err = r.CheckBackingDiskStatusAndUpdateDeviceCR(instance, reqLogger)

	if err != nil {
		// Error while reading, updating object - requeue the request.
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

// CheckBackingDiskStatusAndUpdateDeviceCR checks the status of the backing DiskCR for a block device CR, and updates
// the status of Block Device. If the backing Disk of a block-device is marked inactive, then all block devices
// which use that disk should be marked inactive.
func (r *ReconcileBlockDevice) CheckBackingDiskStatusAndUpdateDeviceCR(
	instance *openebsv1alpha1.BlockDevice, reqLogger logr.Logger) error {

	// If the BlockDevice is of type sparse, then we need not check the backing disk
	// status, As sparse block devices does not have a backing disk.
	if instance.Spec.Details.DeviceType == ndm.SparseBlockDeviceType {
		return nil
	}

	// Find the name of diskCR that need to be read from etcd
	// TODO: This need to be changed, Currently name of disk and blockdevice
	// are using same string except prefix which would be "blockdevice"/"disk"
	UUID := strings.TrimPrefix(instance.ObjectMeta.Name, udev.NDMBlockDevicePrefix)
	Name := udev.NDMDiskPrefix + UUID

	// Fetch the Disk CR
	diskInstance := &openebsv1alpha1.Disk{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: Name}, diskInstance)
	if err != nil {
		reqLogger.Error(err, "Error while getting Disk-CR", "Disk-CR:", Name)
		if errors.IsNotFound(err) {
			reqLogger.Error(err, "Disk CR not found", Name)
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			dcpyInstance := instance.DeepCopy()
			dcpyInstance.Status.State = ndm.NDMInactive
			err := r.client.Update(context.TODO(), dcpyInstance)
			if err != nil {
				reqLogger.Error(err, "Error while updating BlockDevice-CR", "BlockDevice-CR", instance.ObjectMeta.Name)
			}
			reqLogger.Info("BlockDevice-CR marked Inactive", "BlockDevice:", instance.ObjectMeta.Name)
			return err
		}
		return err
	}

	reqLogger.Info("Disk-CR found", "Disk Name:",
		diskInstance.ObjectMeta.Name, "State:", diskInstance.Status.State)
	if strings.Compare(diskInstance.Status.State, ndm.NDMInactive) == 0 {
		dcpyInstance := instance.DeepCopy()
		dcpyInstance.Status.State = ndm.NDMInactive
		err := r.client.Update(context.TODO(), dcpyInstance)
		if err != nil {
			reqLogger.Error(err, "Error while updating BlockDevice-CR", "BlockDevice-CR", instance.ObjectMeta.Name)
			return err
		}
		reqLogger.Info("BlockDevice-CR marked Inactive", "BlockDevice:", instance.ObjectMeta.Name)
	}
	return nil
}
