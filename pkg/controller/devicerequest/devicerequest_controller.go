package devicerequest

import (
	"context"
	"fmt"
	"strings"

	openebsv1alpha1 "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	//corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/go-logr/logr"
	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"k8s.io/client-go/tools/reference"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_devicerequest")

const (
	FinalizerName = "devicerequest.finalizer"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new DeviceRequest Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDeviceRequest{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("devicerequest-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource DeviceRequest
	err = c.Watch(&source.Kind{Type: &openebsv1alpha1.DeviceRequest{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDeviceRequest{}

// ReconcileDeviceRequest reconciles a DeviceRequest object
type ReconcileDeviceRequest struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a DeviceRequest object and makes changes based on the state read
// and what is in the DeviceRequest.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileDeviceRequest) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling DeviceRequest")

	// Fetch the DeviceRequest instance
	instance := &openebsv1alpha1.DeviceRequest{}
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

	// New deviceRequest if Phase is NULL
	if instance.Status.Phase == openebsv1alpha1.DeviceRequestStatusEmpty {
		err := r.claimDeviceForDeviceRequestCR(instance, reqLogger)
		if err != nil {
			//reqLogger.Error(err, "DeviceRequest failed:", instance.ObjectMeta.Name)
			return reconcile.Result{}, err
		}
	} else if instance.Status.Phase == openebsv1alpha1.DeviceRequestStatusPending {

	} else if (instance.Status.Phase == openebsv1alpha1.DeviceRequestStatusDone) ||
		(instance.Status.Phase == openebsv1alpha1.DeviceRequestStatusInvalidCapacity) {

		reqLogger.Info("In process of deleting deviceRequest")
		err := r.FinalizerHandling(instance, reqLogger)
		if err != nil {
			reqLogger.Error(err, "Finalizer handling failed", "DeviceRequest-CR:", instance.ObjectMeta.Name)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// Capacity check is to check if a valid capacity requested or not
func (r *ReconcileDeviceRequest) isValidCapacityRequested(instance *openebsv1alpha1.DeviceRequest,
	reqLogger logr.Logger) error {

	// Check if deviceCalim has valid capacity request
	if instance.Spec.Capacity <= 0 {
		err1 := fmt.Errorf("Invalid Capacity requested")

		//Update deviceRequest CR with error string
		instance_cpy := instance.DeepCopy()
		instance_cpy.Status.Phase = openebsv1alpha1.DeviceRequestStatusInvalidCapacity

		err := r.client.Delete(context.TODO(), instance_cpy)
		if err != nil {
			reqLogger.Error(err, "Invalid capacity requested, deletion failed", "DeviceRequest-CR:", instance.ObjectMeta.Name)
			return err
		}
		return err1
	}
	return nil
}

func (r *ReconcileDeviceRequest) getListofDevicesOnHost(instance *openebsv1alpha1.DeviceRequest,
	reqLogger logr.Logger) (*openebsv1alpha1.DeviceList, error) {

	//Initialize a deviceList object.
	listDVR := &openebsv1alpha1.DeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Device",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	//Set filter option, in our case we are filtering based on hostname/node
	opts := &client.ListOptions{}
	filter := ndm.NDMHostKey + "=" + instance.Spec.HostName
	//reqLogger.Info("Filter string", "filter:", filter, "instance:", instance)

	opts.SetLabelSelector(filter)

	//Fetch deviceList with matching criteria
	err := r.client.List(context.TODO(), opts, listDVR)
	if err != nil {
		reqLogger.Error(err, "Error getting DeviceList", "DeviceRequest-CR:", instance.ObjectMeta.Name)
		return nil, err
	}

	//Check if listDVR is null or not
	length := len(listDVR.Items)
	if length == 0 {
		err = fmt.Errorf("No device found with matching criteria")
		return nil, err
	}
	return listDVR, nil
}

func (r *ReconcileDeviceRequest) claimDeviceCR(instance *openebsv1alpha1.DeviceRequest,
	listDVR *openebsv1alpha1.DeviceList, reqLogger logr.Logger) error {

	var driveTypeSpecified bool = false

	// Check if request is for SSD or HDD
	length := len(instance.Spec.DeviceType)
	if length != 0 {
		reqLogger.Info("DriveType specified in DeviceRequest CR")
		driveTypeSpecified = true
	}

	//fmt.Print("Device List:", listDVR)

	//Find a device which is free and have available
	//space more than or equal to requested
	for _, item := range listDVR.Items {
		if (strings.Compare(item.ClaimState.State, ndm.NDMUnclaimed) == 0) &&
			(strings.Compare(item.Status.State, ndm.NDMActive) == 0) &&
			(item.Spec.Capacity.Storage >= instance.Spec.Capacity) {

			if driveTypeSpecified == true {
				if strings.Compare(item.Spec.Details.DeviceType, instance.Spec.DeviceType) != 0 {
					continue
				}
			}

			reqLogger.Info("Found matching device", "Device Name:",
				item.ObjectMeta.Name, "Device Capacity:",
				item.Spec.Capacity.Storage)
			claimRef, err := reference.GetReference(r.scheme, instance)
			if err != nil {
				reqLogger.Error(err, "Unexpected error getting claim reference", "Device-CR:", instance.ObjectMeta.Name)
				return err
			}

			// Found free device, mark it claimed, put ClaimRef of DeviceRequest-CR
			dvr := item.DeepCopy()
			dvr.ClaimState.State = ndm.NDMClaimed
			dvr.ClaimRef = claimRef
			err = r.client.Update(context.TODO(), dvr)
			if err != nil {
				reqLogger.Error(err, "Error while updating Device-CR", "Device-CR:", dvr.ObjectMeta.Name)
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("No device found to claim")
}

/*
 * New DeviceRequest CR is created, try to findout device which is
 * free and has size equal/greater than DeviceRequest request.
 */
func (r *ReconcileDeviceRequest) claimDeviceForDeviceRequestCR(
	instance *openebsv1alpha1.DeviceRequest, reqLogger logr.Logger) error {

	// Check if capacity requested is 0 or -
	err := r.isValidCapacityRequested(instance, reqLogger)
	if err != nil {
		return err
	}

	//Get Device list for particular host
	listDVR, err := r.getListofDevicesOnHost(instance, reqLogger)
	if err != nil {
		return err
	}

	err = r.claimDeviceCR(instance, listDVR, reqLogger)
	if err != nil {
		reqLogger.Error(err, "Error claiming Device", "DeviceRequest-CR:", instance.ObjectMeta.Name)
		return err
	}

	// Update DeviceRequest CR to show that "claim is done" also set
	// finalizer string on instance
	instance_cpy := instance.DeepCopy()
	instance_cpy.Status.Phase = openebsv1alpha1.DeviceRequestStatusDone
	instance_cpy.ObjectMeta.Finalizers = append(instance_cpy.ObjectMeta.Finalizers, FinalizerName)
	err = r.client.Update(context.TODO(), instance_cpy)
	if err != nil {
		reqLogger.Error(err, "Error while updating", "DeviceRequest-CR:", instance.ObjectMeta.Name)
		return err
	}
	return nil
}

func (r *ReconcileDeviceRequest) FinalizerHandling(
	instance *openebsv1alpha1.DeviceRequest, reqLogger logr.Logger) error {

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		reqLogger.Info("No Deletion Time Stamp set on deviceRequest")
		return nil
	}
	// The object is being deleted
	if containsString(instance.ObjectMeta.Finalizers, FinalizerName) {
		// Finalizer is set, lets handle external dependency
		if err := r.deleteExternalDependency(instance, reqLogger); err != nil {
			reqLogger.Error(err, "Could not delete external dependency", "DeviceRequest-CR:", instance.ObjectMeta.Name)
			return err
		}

		// Remove finalizer from list and update it.
		instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, FinalizerName)
		if err := r.client.Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Could not remove finalizer", "DeviceRequest-CR:", instance.ObjectMeta.Name)
			return err
		}
	}

	return nil
}

func (r *ReconcileDeviceRequest) isDeviceRequestedByThisDeviceRequest(
	instance *openebsv1alpha1.DeviceRequest, item openebsv1alpha1.Device,
	reqLogger logr.Logger) bool {

	if strings.Compare(item.ClaimState.State, ndm.NDMClaimed) != 0 {
		reqLogger.Info("Found device which yet to be claimed")
		return false
	}

	if strings.Compare(item.ClaimRef.Name, instance.ObjectMeta.Name) != 0 {
		reqLogger.Info("ClaimRef Name mismatch")
		return false
	}

	if item.ClaimRef.UID != instance.ObjectMeta.UID {
		reqLogger.Info("DeviceRequest UID mismatch")
		return false
	}

	if strings.Compare(item.ClaimRef.Kind, instance.TypeMeta.Kind) != 0 {
		reqLogger.Info("Kind mismatch")
		return false
	}
	return true
}

func (r *ReconcileDeviceRequest) deleteExternalDependency(
	instance *openebsv1alpha1.DeviceRequest, reqLogger logr.Logger) error {

	reqLogger.Info("Deleting external dependencies for CR:", instance)

	//Get Device list for particular host
	listDVR, err := r.getListofDevicesOnHost(instance, reqLogger)
	if err != nil {
		return err
	}

	// Check if same deviceRequest holding the ObjRef
	for _, item := range listDVR.Items {
		if r.isDeviceRequestedByThisDeviceRequest(instance, item, reqLogger) == false {
			continue
		}

		// Found a device ObjRef with DeviceRequest, Clear
		// ObjRef and mark device unclaimed in etcd
		dvr := item.DeepCopy()
		dvr.ClaimRef = nil
		dvr.ClaimState.State = ndm.NDMUnclaimed
		err := r.client.Update(context.TODO(), dvr)
		if err != nil {
			reqLogger.Error(err, "Error while updating ObjRef", "Device-CR:", dvr.ObjectMeta.Name)
			return err
		}
	}
	return nil
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
