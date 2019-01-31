package deviceclaim

import (
	"context"
	"fmt"
	"strings"

	openebsv1alpha1 "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/go-logr/logr"
	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_deviceclaim")

const (
	FinalizerName = "deviceclaim.finalizer"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new DeviceClaim Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDeviceClaim{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("deviceclaim-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource DeviceClaim
	err = c.Watch(&source.Kind{Type: &openebsv1alpha1.DeviceClaim{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner DeviceClaim
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &openebsv1alpha1.DeviceClaim{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDeviceClaim{}

// ReconcileDeviceClaim reconciles a DeviceClaim object
type ReconcileDeviceClaim struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a DeviceClaim object and makes changes based on the state read
// and what is in the DeviceClaim.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileDeviceClaim) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling DeviceClaim")

	// Fetch the DeviceClaim instance
	instance := &openebsv1alpha1.DeviceClaim{}
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

	// New deviceclaim if Phase is NULL
	if instance.Status.Phase == openebsv1alpha1.DeviceClaimStatusEmpty {
		err := r.claimDeviceForDeviceClaimCR(instance, reqLogger)
		if err != nil {
			reqLogger.Info("DeviceClaim failed")
			return reconcile.Result{}, err
		}
	} else if instance.Status.Phase == openebsv1alpha1.DeviceClaimStatusPending {

	} else if (instance.Status.Phase == openebsv1alpha1.DeviceClaimStatusDone) ||
		(instance.Status.Phase == openebsv1alpha1.DeviceClaimStatusInvalidCapacity) {

		reqLogger.Info("In process of deleting deviceClaim")
		err := r.FinalizerHandling(instance, reqLogger)
		if err != nil {
			reqLogger.Error(err, "Finalizer handling failed")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

/*
 * New DeviceClaim CR is created, try to findout device which is
 * free and has size equal/greater than DeviceClaim request.
 */
func (r *ReconcileDeviceClaim) claimDeviceForDeviceClaimCR(
	instance *openebsv1alpha1.DeviceClaim, reqLogger logr.Logger) error {

	var driveTypeSpecified bool = false
	// Check if deviceCalim has valid capacity request
	if instance.Spec.Capacity <= 0 {
		err1 := fmt.Errorf("Invalid Capacity requested")
		reqLogger.Error(err1, "Capacity requested is invalid")

		//Update deviceClaim CR to with error string
		instance_cpy := instance.DeepCopy()
		instance_cpy.Status.Phase = openebsv1alpha1.DeviceClaimStatusInvalidCapacity
		instance_cpy.ObjectMeta.Finalizers = append(instance_cpy.ObjectMeta.Finalizers, FinalizerName)
		err := r.client.Update(context.TODO(), instance_cpy)
		if err != nil {
			reqLogger.Error(err, "Invalid Capacity error update to DeviceClaimCR failed")
		}
		return err1
	}

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
	reqLogger.Info("Filter string", "filter:", filter, "instance:", instance)

	opts.SetLabelSelector(filter)

	//Fetch deviceList with matching criteria
	err := r.client.List(context.TODO(), opts, listDVR)
	if err != nil {
		reqLogger.Info("Error while getting device list")
		return err
	}

	//Check if listDVR is null or not
	length := len(listDVR.Items)
	if length == 0 {
		reqLogger.Info("No device found with matching criteria")
		return fmt.Errorf("No device found with matching criteria")
	}

	length = len(instance.Spec.DriveType)
	if length != 0 {
		reqLogger.Info("DriveType specified in DeviceClaim CR")
		driveTypeSpecified = true
	}

	//Find a device which is free and have available
	//space more than or equal to requested
	for _, item := range listDVR.Items {
		if (strings.Compare(item.ClaimState.State, ndm.NDMUnclaimed) == 0) &&
			(item.Spec.Capacity.Storage >= instance.Spec.Capacity) {

			if driveTypeSpecified == true {
				if strings.Compare(item.Spec.Details.DriveType, instance.Spec.DriveType) != 0 {
					continue
				}
			}

			reqLogger.Info("Found matching device", "Device Name:",
				item.ObjectMeta.Name, "Device Capacity:",
				item.Spec.Capacity.Storage)

			//Found a device, claim it, update ClaimInfo into etcd
			dvr := item.DeepCopy()
			dvr.Claim.APIVersion = instance.TypeMeta.APIVersion
			dvr.Claim.Kind = instance.TypeMeta.Kind
			dvr.Claim.Name = instance.ObjectMeta.Name
			dvr.Claim.DeviceClaimUID = instance.ObjectMeta.UID
			dvr.ClaimState.State = ndm.NDMClaimed
			err := r.client.Update(context.TODO(), dvr)
			if err != nil {
				reqLogger.Info("Error while updating device CR")
				return err
			}

			//Update deviceClaim CR to show that device claim happened
			instance_cpy := instance.DeepCopy()
			instance_cpy.Status.Phase = openebsv1alpha1.DeviceClaimStatusDone
			instance_cpy.ObjectMeta.Finalizers = append(instance_cpy.ObjectMeta.Finalizers, FinalizerName)
			err = r.client.Update(context.TODO(), instance_cpy)
			if err != nil {
				reqLogger.Info("Error while updating deviceClaim CR")
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("No device found to claim")
}

func (r *ReconcileDeviceClaim) FinalizerHandling(
	instance *openebsv1alpha1.DeviceClaim, reqLogger logr.Logger) error {

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		reqLogger.Info("No Deletion Time Stamp set on deviceClaim")
		return nil
	}
	// The object is being deleted
	if containsString(instance.ObjectMeta.Finalizers, FinalizerName) {
		// Finalizer is set, lets handle external dependency
		if err := r.deleteExternalDependency(instance, reqLogger); err != nil {
			reqLogger.Error(err, "Could not delete external dependency")
			return err
		}

		// Remove finalizer from list and update it.
		instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, FinalizerName)
		if err := r.client.Update(context.Background(), instance); err != nil {
			reqLogger.Error(err, "Finalizer could not be removed")
			return err
		}
	}

	// Finalizer has finished
	return nil
}

func (r *ReconcileDeviceClaim) isDeviceClaimedByThisDeviceClaim(
	instance *openebsv1alpha1.DeviceClaim, item openebsv1alpha1.Device,
	reqLogger logr.Logger) bool {

	if strings.Compare(item.ClaimState.State, ndm.NDMClaimed) != 0 {
		reqLogger.Info("Found device which yet to be claimed")
		return false
	}

	if strings.Compare(item.Claim.Name, instance.ObjectMeta.Name) != 0 {
		reqLogger.Info("Claim Name mismatch")
		return false
	}

	if item.Claim.DeviceClaimUID != instance.ObjectMeta.UID {
		reqLogger.Info("DeviceClaimUID mismatch")
		return false
	}

	if strings.Compare(item.Claim.Kind, instance.TypeMeta.Kind) != 0 {
		reqLogger.Info("Kind mismatch")
		return false
	}
	return true
}

func (r *ReconcileDeviceClaim) deleteExternalDependency(
	instance *openebsv1alpha1.DeviceClaim, reqLogger logr.Logger) error {
	reqLogger.Info("Deleting external dependencies for CR:", instance)

	//Initialize an deviceList object.
	listDVR := &openebsv1alpha1.DeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Device",
			APIVersion: "openebs.io/v1alpha1",
		},
	}
	//Set filter option, filter based on hostname/node
	opts := &client.ListOptions{}
	filter := ndm.NDMHostKey + "=" + instance.Spec.HostName
	reqLogger.Info("Filter string", "filter:", filter, "instance:", instance)
	if err := opts.SetLabelSelector(filter); err != nil {
		reqLogger.Error(err, "Error in SetLabelSelector")
		return err
	}

	//Fetch deviceList with matching criteria
	err := r.client.List(context.TODO(), opts, listDVR)
	if err != nil {
		reqLogger.Error(err, "Error while getting device list")
		return err
	}

	//Check if listDVR is null or not
	length := len(listDVR.Items)
	if length == 0 {
		reqLogger.Info("No device found with matching criteria")
		return fmt.Errorf("No device found with matching criteria")
	}

	// Check Claim and check if Claim.DeviceClaimUID
	// is matching with instance.ObjectMeta.UID
	for _, item := range listDVR.Items {
		if r.isDeviceClaimedByThisDeviceClaim(instance, item, reqLogger) == false {
			continue
		}

		//Found a device, clear claim, mark it
		//unclaimed and update ClaimInfo into etcd
		dvr := item.DeepCopy()
		dvr.Claim.APIVersion = ""
		dvr.Claim.Kind = ""
		dvr.Claim.Name = ""
		dvr.Claim.DeviceClaimUID = ""
		dvr.ClaimState.State = ndm.NDMUnclaimed
		err := r.client.Update(context.TODO(), dvr)
		if err != nil {
			reqLogger.Error(err, "Error while updating Claim of device CR")
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
