package blockdeviceclaim

import (
	"context"
	"fmt"
	"github.com/openebs/node-disk-manager/pkg/util"
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

var log = logf.Log.WithName("controller_deviceclaim")

const (
	// BlockDeviceClaimFinalizer is the finalizer name for the block device claim
	BlockDeviceClaimFinalizer = "blockdeviceclaim.finalizer"
)

// Add creates a new BlockDeviceClaim Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBlockDeviceClaim{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("blockdeviceclaim-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource BlockDeviceClaim
	err = c.Watch(&source.Kind{Type: &openebsv1alpha1.BlockDeviceClaim{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileBlockDeviceClaim{}

// ReconcileBlockDeviceClaim reconciles a BlockDeviceClaim object
type ReconcileBlockDeviceClaim struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a BlockDeviceClaim object and makes changes based on the state read
// and what is in the BlockDeviceClaim.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBlockDeviceClaim) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling BlockDeviceClaim")

	// Fetch the BlockDeviceClaim instance
	instance := &openebsv1alpha1.BlockDeviceClaim{}
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

	// New device claim if Phase is NULL
	if instance.Status.Phase == openebsv1alpha1.BlockDeviceClaimStatusEmpty {
		err := r.claimDeviceForBlockDeviceClaimCR(instance, reqLogger)
		if err != nil {
			//reqLogger.Error(err, "BlockDeviceClaim failed:", instance.ObjectMeta.Name)
			return reconcile.Result{}, err
		}
	} else if instance.Status.Phase == openebsv1alpha1.BlockDeviceClaimStatusPending {

	} else if (instance.Status.Phase == openebsv1alpha1.BlockDeviceClaimStatusDone) ||
		(instance.Status.Phase == openebsv1alpha1.BlockDeviceClaimStatusInvalidCapacity) {

		reqLogger.Info("In process of deleting block device claim")
		err := r.FinalizerHandling(instance, reqLogger)
		if err != nil {
			reqLogger.Error(err, "Finalizer handling failed", "BlockDeviceClaim-CR:", instance.ObjectMeta.Name)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// Capacity check is to check if a valid capacity requested or not
func (r *ReconcileBlockDeviceClaim) getRequestedCapacity(instance *openebsv1alpha1.BlockDeviceClaim,
	reqLogger logr.Logger) (int64, error) {

	resourceCapacity := instance.Spec.Requirements.Requests[openebsv1alpha1.ResourceCapacity]
	// Check if deviceCalim has valid capacity request
	capacity, err := (&resourceCapacity).AsInt64()
	if !err || capacity <= 0 {
		return 0, fmt.Errorf("invalid capacity requested")
	}
	return capacity, nil
}

func (r *ReconcileBlockDeviceClaim) getListofDevicesOnHost(instance *openebsv1alpha1.BlockDeviceClaim,
	reqLogger logr.Logger) (*openebsv1alpha1.BlockDeviceList, error) {

	//Initialize a deviceList object.
	listDVR := &openebsv1alpha1.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
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
		reqLogger.Error(err, "Error getting BlockDeviceList", "BlockDeviceClaim-CR:", instance.ObjectMeta.Name)
		return nil, err
	}

	//Check if listDVR is null or not
	length := len(listDVR.Items)
	if length == 0 {
		err = fmt.Errorf("No blockdevice found with matching criteria")
		return nil, err
	}
	return listDVR, nil
}

func (r *ReconcileBlockDeviceClaim) claimDeviceCR(instance *openebsv1alpha1.BlockDeviceClaim,
	listDVR *openebsv1alpha1.BlockDeviceList, capacity int64, reqLogger logr.Logger) error {

	var driveTypeSpecified bool = false

	// Check if claim is for SSD or HDD
	length := len(instance.Spec.DeviceType)
	if length != 0 {
		reqLogger.Info("DriveType specified in BlockDeviceClaim CR")
		driveTypeSpecified = true
	}

	//fmt.Print("BlockDevice List:", listDVR)

	//Find a blockdevice which is free and have available
	//space more than or equal to requested
	for _, item := range listDVR.Items {
		if (strings.Compare(item.ClaimState.State, ndm.NDMUnclaimed) == 0) &&
			(strings.Compare(item.Status.State, ndm.NDMActive) == 0) &&
			(item.Spec.Capacity.Storage >= uint64(capacity)) {

			if driveTypeSpecified == true {
				if strings.Compare(item.Spec.Details.DeviceType, instance.Spec.DeviceType) != 0 {
					continue
				}
			}

			reqLogger.Info("Found matching blockdevice", "BlockDevice Name:",
				item.ObjectMeta.Name, "BlockDevice Capacity:",
				item.Spec.Capacity.Storage)
			claimRef, err := reference.GetReference(r.scheme, instance)
			if err != nil {
				reqLogger.Error(err, "Unexpected error getting claim reference", "BlockDevice-CR:", instance.ObjectMeta.Name)
				return err
			}

			// Found free blockdevice, mark it claimed, put ClaimRef of BlockDeviceClaim-CR
			dvr := item.DeepCopy()
			dvr.ClaimState.State = ndm.NDMClaimed
			dvr.ClaimRef = claimRef
			err = r.client.Update(context.TODO(), dvr)
			if err != nil {
				reqLogger.Error(err, "Error while updating BlockDevice-CR", "BlockDevice-CR:", dvr.ObjectMeta.Name)
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("No blockdevice found to claim")
}

/*
 * New BlockDeviceClaim CR is created, try to findout blockdevice which is
 * free and has size equal/greater than BlockDeviceClaim request.
 */
func (r *ReconcileBlockDeviceClaim) claimDeviceForBlockDeviceClaimCR(
	instance *openebsv1alpha1.BlockDeviceClaim, reqLogger logr.Logger) error {

	// Get the capacity requested in the claim
	capacity, err := r.getRequestedCapacity(instance, reqLogger)
	if err != nil {
		//Update deviceClaim CR with error string
		instanceCopy := instance.DeepCopy()
		instanceCopy.Status.Phase = openebsv1alpha1.BlockDeviceClaimStatusInvalidCapacity

		err1 := r.client.Delete(context.TODO(), instanceCopy)
		if err1 != nil {
			reqLogger.Error(err1, "Invalid capacity requested, deletion failed", "BlockDeviceClaim-CR:", instance.ObjectMeta.Name)
			return err1
		}
		return err
	}

	//Get BlockDevice list for particular host
	listDVR, err := r.getListofDevicesOnHost(instance, reqLogger)
	if err != nil {
		return err
	}

	err = r.claimDeviceCR(instance, listDVR, capacity, reqLogger)
	if err != nil {
		reqLogger.Error(err, "Error claiming BlockDevice", "BlockDeviceClaim-CR:", instance.ObjectMeta.Name)
		return err
	}

	// Update BlockDeviceClaim CR to show that "claim is done" also set
	// finalizer string on instance
	instance_cpy := instance.DeepCopy()
	instance_cpy.Status.Phase = openebsv1alpha1.BlockDeviceClaimStatusDone
	instance_cpy.ObjectMeta.Finalizers = append(instance_cpy.ObjectMeta.Finalizers, BlockDeviceClaimFinalizer)
	err = r.client.Update(context.TODO(), instance_cpy)
	if err != nil {
		reqLogger.Error(err, "Error while updating", "BlockDeviceClaim-CR:", instance.ObjectMeta.Name)
		return err
	}
	return nil
}

// FinalizerHandling removes the finalizer from the claim resource
func (r *ReconcileBlockDeviceClaim) FinalizerHandling(
	instance *openebsv1alpha1.BlockDeviceClaim, reqLogger logr.Logger) error {

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		reqLogger.Info("No Deletion Time Stamp set on device claim")
		return nil
	}
	// The object is being deleted
	if util.Contains(instance.ObjectMeta.Finalizers, BlockDeviceClaimFinalizer) {
		// Finalizer is set, lets handle external dependency
		if err := r.deleteExternalDependency(instance, reqLogger); err != nil {
			reqLogger.Error(err, "Could not delete external dependency", "BlockDeviceClaim-CR:", instance.ObjectMeta.Name)
			return err
		}

		// Remove finalizer from list and update it.
		instance.ObjectMeta.Finalizers = util.RemoveString(instance.ObjectMeta.Finalizers, BlockDeviceClaimFinalizer)
		if err := r.client.Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Could not remove finalizer", "BlockDeviceClaim-CR:", instance.ObjectMeta.Name)
			return err
		}
	}

	return nil
}

func (r *ReconcileBlockDeviceClaim) isDeviceRequestedByThisDeviceClaim(
	instance *openebsv1alpha1.BlockDeviceClaim, item openebsv1alpha1.BlockDevice,
	reqLogger logr.Logger) bool {

	if strings.Compare(item.ClaimState.State, ndm.NDMClaimed) != 0 {
		reqLogger.Info("Found blockdevice which yet to be claimed")
		return false
	}

	if strings.Compare(item.ClaimRef.Name, instance.ObjectMeta.Name) != 0 {
		reqLogger.Info("ClaimRef Name mismatch")
		return false
	}

	if item.ClaimRef.UID != instance.ObjectMeta.UID {
		reqLogger.Info("BlockDeviceClaim UID mismatch")
		return false
	}

	if strings.Compare(item.ClaimRef.Kind, instance.TypeMeta.Kind) != 0 {
		reqLogger.Info("Kind mismatch")
		return false
	}
	return true
}

func (r *ReconcileBlockDeviceClaim) deleteExternalDependency(
	instance *openebsv1alpha1.BlockDeviceClaim, reqLogger logr.Logger) error {

	reqLogger.Info("Deleting external dependencies for CR:", instance)

	//Get BlockDevice list for particular host
	listDVR, err := r.getListofDevicesOnHost(instance, reqLogger)
	if err != nil {
		return err
	}

	// Check if same deviceclaim holding the ObjRef
	for _, item := range listDVR.Items {
		if r.isDeviceRequestedByThisDeviceClaim(instance, item, reqLogger) == false {
			continue
		}

		// Found a blockdevice ObjRef with BlockDeviceClaim, Clear
		// ObjRef and mark blockdevice unclaimed in etcd
		dvr := item.DeepCopy()
		dvr.ClaimRef = nil
		dvr.ClaimState.State = ndm.NDMUnclaimed
		err := r.client.Update(context.TODO(), dvr)
		if err != nil {
			reqLogger.Error(err, "Error while updating ObjRef", "BlockDevice-CR:", dvr.ObjectMeta.Name)
			return err
		}
	}
	return nil
}
