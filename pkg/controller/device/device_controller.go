package device

import (
	"context"
	"strings"

	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	openebsv1alpha1 "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/udev"
	//corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"github.com/go-logr/logr"
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

// Add creates a new Device Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDevice{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("device-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Device
	err = c.Watch(&source.Kind{Type: &openebsv1alpha1.Device{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDevice{}

// ReconcileDevice reconciles a Device object
type ReconcileDevice struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Device object and makes changes based on the state read
// and what is in the Device.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileDevice) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Device")

	// Fetch the Device instance
	instance := &openebsv1alpha1.Device{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	err = r.CheckBackingDiskStatusAndUpdateDeviceCR(instance,
		request.NamespacedName.Namespace, reqLogger)

	if err != nil {
		// Error while reading, updating object - requeue the request.
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileDevice) CheckBackingDiskStatusAndUpdateDeviceCR(
	instance *openebsv1alpha1.Device, nameSpace string, reqLogger logr.Logger) error {

	//Find the name of diskCR that need to be read from etcd
	Uuid := strings.TrimPrefix(instance.ObjectMeta.Name, udev.NDMDevicePrefix)
	Name := udev.NDMDiskPrefix + Uuid

	// Fetch the Disk CR
	diskInstance := &openebsv1alpha1.Disk{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: Name, Namespace: nameSpace}, diskInstance)
	if err != nil {
		reqLogger.Info("No disk-CR found", "Disk Name:",
			Name, "Namespace:", nameSpace)
		if errors.IsNotFound(err) {
			reqLogger.Info("Disk CR not found")
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			dcpyInstance := instance.DeepCopy()
			dcpyInstance.Status.State = ndm.NDMInactive
			err := r.client.Update(context.TODO(), dcpyInstance)
			if err != nil {
				reqLogger.Info("Error while updating device CR")
			}
			return err
		}
		reqLogger.Info("Device CR marked Inactive", "Device:", instance.ObjectMeta.Name)
		return err
	}

	reqLogger.Info("disk-CR found", "Disk Name:",
		diskInstance.ObjectMeta.Name, "State:", diskInstance.Status.State)
	if strings.Compare(diskInstance.Status.State, ndm.NDMInactive) == 0 {
		dcpyInstance := instance.DeepCopy()
		dcpyInstance.Status.State = ndm.NDMInactive
		err := r.client.Update(context.TODO(), dcpyInstance)
		if err != nil {
			reqLogger.Info("Error while updating device CR")
			return err
		}
		reqLogger.Info("Device CR marked Inactive", "Device:", instance.ObjectMeta.Name)
	}
	return nil
}