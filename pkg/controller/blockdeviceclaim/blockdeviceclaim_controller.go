package blockdeviceclaim

import (
	"context"
	"fmt"
	"github.com/openebs/node-disk-manager/pkg/util"
	"k8s.io/api/core/v1"
	"strconv"

	"github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
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
	err = c.Watch(&source.Kind{Type: &v1alpha1.BlockDeviceClaim{}}, &handler.EnqueueRequestForObject{})
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
	instance := &v1alpha1.BlockDeviceClaim{}
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

	reqLogger.Info("BlockDeviceClaim State is:" + string(instance.Status.Phase))
	switch instance.Status.Phase {
	case v1alpha1.BlockDeviceClaimStatusPending:
		fallthrough
	case v1alpha1.BlockDeviceClaimStatusEmpty:
		err := r.claimDeviceForBlockDeviceClaim(instance, reqLogger)
		if err != nil {
			reqLogger.Error(err, "BlockDeviceClaim "+instance.ObjectMeta.Name+" failed")
			return reconcile.Result{}, err
		}
	case v1alpha1.BlockDeviceClaimStatusInvalidCapacity:
		// currently for invalid capacity, the device claim will be deleted
		fallthrough
	case v1alpha1.BlockDeviceClaimStatusDone:
		reqLogger.Info("In process of deleting block device claim")
		err := r.FinalizerHandling(instance, reqLogger)
		if err != nil {
			reqLogger.Error(err, "Finalizer handling failed", "BlockDeviceClaim-CR:", instance.ObjectMeta.Name)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// getRequestedCapacity gets the requested capacity from the BlockDeviceClaim
// It returns an error if the Quantity cannot be parsed
func getRequestedCapacity(list v1.ResourceList) (int64, error) {

	resourceCapacity := list[v1alpha1.ResourceCapacity]
	// Check if deviceCalim has valid capacity request
	capacity, err := (&resourceCapacity).AsInt64()
	if !err || capacity <= 0 {
		return 0, fmt.Errorf("invalid capacity requested, %v", err)
	}
	return capacity, nil
}

// getListofDevicesOnHost gets the list of block devices on the node to which BlockDeviceClaim is made
// TODO:
//  ListBlockDeviceResource in package cmd/ndm_daemonset/controller has the same functionality.
//  Need to merge these 2 functions.
func (r *ReconcileBlockDeviceClaim) getListofDevicesOnHost(instance *v1alpha1.BlockDeviceClaim,
	reqLogger logr.Logger) (*v1alpha1.BlockDeviceList, error) {

	//Initialize a deviceList object.
	listDVR := &v1alpha1.BlockDeviceList{
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
		err = fmt.Errorf("No blockdevice found on the given node")
		return nil, err
	}
	return listDVR, nil
}

// claimDeviceForBlockDeviceClaim is created, try to findout blockdevice which is
// free and has size equal/greater than BlockDeviceClaim request.
func (r *ReconcileBlockDeviceClaim) claimDeviceForBlockDeviceClaim(
	instance *v1alpha1.BlockDeviceClaim, reqLogger logr.Logger) error {

	// Get the capacity requested in the claim
	_, err := getRequestedCapacity(instance.Spec.Requirements.Requests)
	if err != nil {
		//Update deviceClaim CR with error string
		instance.Status.Phase = v1alpha1.BlockDeviceClaimStatusInvalidCapacity
		err1 := r.client.Delete(context.TODO(), instance)
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

	// get matching BD list
	matchingDevices := r.getMatchingBlockDevices(instance, listDVR, reqLogger)

	// select the block device based on resource requirements
	if selectedDevice, ok := selectBlockDevice(matchingDevices, instance.Spec.Requirements.Requests); ok {
		claimRef, err := reference.GetReference(r.scheme, instance)
		if err != nil {
			reqLogger.Error(err, "error getting claim reference", "BlockDevice-CR:", instance.ObjectMeta.Name)
			return err
		}
		selectedDevice.Spec.ClaimRef = claimRef
		selectedDevice.Status.ClaimState = v1alpha1.BlockDeviceClaimed
		err = r.client.Update(context.TODO(), &selectedDevice)
		if err != nil {
			reqLogger.Error(err, "Error while updating BlockDevice-CR", "BlockDevice-CR:", selectedDevice.ObjectMeta.Name)
			return err
		}
		reqLogger.Info("Block Device " + selectedDevice.Name + " claimed")
		instance.Spec.BlockDeviceName = selectedDevice.Name
		instance.Status.Phase = v1alpha1.BlockDeviceClaimStatusDone
	} else {
		reqLogger.Info("Could not find a BlockDevice which satisfies the claim. Changing to pending state")
		instance.Status.Phase = v1alpha1.BlockDeviceClaimStatusPending
	}

	// set finalizer string on instance if claim status is Done
	if instance.Status.Phase == v1alpha1.BlockDeviceClaimStatusDone {
		reqLogger.Info("Added finalizers to BDC:" + instance.Name)
		instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, BlockDeviceClaimFinalizer)
	}
	// Update BlockDeviceClaim CR
	err = r.client.Update(context.TODO(), instance)
	if err != nil {
		reqLogger.Error(err, "Error while updating", "BlockDeviceClaim-CR:", instance.ObjectMeta.Name)
		return err
	}
	return nil
}

// FinalizerHandling removes the finalizer from the claim resource
func (r *ReconcileBlockDeviceClaim) FinalizerHandling(
	instance *v1alpha1.BlockDeviceClaim, reqLogger logr.Logger) error {

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		reqLogger.Info("No Deletion Time Stamp set on device claim")
		return nil
	}
	// The object is being deleted
	if util.Contains(instance.ObjectMeta.Finalizers, BlockDeviceClaimFinalizer) {
		// Finalizer is set, lets handle external dependency
		if err := r.deleteClaimedBlockDevice(instance, reqLogger); err != nil {
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

// isDeviceRequestedByThisDeviceClaim checks whether a claimed block device belongs to the given BDC
func (r *ReconcileBlockDeviceClaim) isDeviceRequestedByThisDeviceClaim(
	instance *v1alpha1.BlockDeviceClaim, item v1alpha1.BlockDevice,
	reqLogger logr.Logger) bool {

	if item.Status.ClaimState != v1alpha1.BlockDeviceClaimed {
		reqLogger.Info("Found blockdevice which yet to be claimed")
		return false
	}

	if item.Spec.ClaimRef.Name != instance.ObjectMeta.Name {
		reqLogger.Info("ClaimRef Name mismatch")
		return false
	}

	if item.Spec.ClaimRef.UID != instance.ObjectMeta.UID {
		reqLogger.Info("BlockDeviceClaim UID mismatch")
		return false
	}

	if item.Spec.ClaimRef.Kind != instance.TypeMeta.Kind {
		reqLogger.Info("Kind mismatch")
		return false
	}
	return true
}

// deleteClaimedBlockDevice unclaims the block device claimed by this BlockDeviceClaim
func (r *ReconcileBlockDeviceClaim) deleteClaimedBlockDevice(
	instance *v1alpha1.BlockDeviceClaim, reqLogger logr.Logger) error {

	reqLogger.Info("Deleting external dependencies for CR:" + instance.Name)

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
		dvr.Spec.ClaimRef = nil
		dvr.Status.ClaimState = v1alpha1.BlockDeviceUnclaimed
		err := r.client.Update(context.TODO(), dvr)
		if err != nil {
			reqLogger.Error(err, "Error while updating ObjRef", "BlockDevice-CR:", dvr.ObjectMeta.Name)
			return err
		}
	}
	return nil
}

// getMatchingBlockDevices returns a list of BDs which match the
// spec given in the BDC
func (r *ReconcileBlockDeviceClaim) getMatchingBlockDevices(
	instance *v1alpha1.BlockDeviceClaim, bdList *v1alpha1.BlockDeviceList, reqLogger logr.Logger) v1alpha1.BlockDeviceList {
	checkDeviceType := false
	if len(instance.Spec.DeviceType) != 0 {
		checkDeviceType = true
	}

	matchingBlockDevices := v1alpha1.BlockDeviceList{}
	for _, bd := range bdList.Items {
		// check whether the block device is unclaimed and active
		if bd.Status.State != ndm.NDMActive ||
			bd.Status.ClaimState != v1alpha1.BlockDeviceUnclaimed {
			continue
		}
		// check device type
		if checkDeviceType && bd.Spec.Details.DeviceType != instance.Spec.DeviceType {
			continue
		}
		matchingBlockDevices.Items = append(matchingBlockDevices.Items, bd)
	}
	reqLogger.Info("No of matching devices based on Spec : " + strconv.Itoa(len(matchingBlockDevices.Items)))
	return matchingBlockDevices
}

// selectBlockDevice selects a single BlockDevice from the list of
// block devices based on the resource requirements
func selectBlockDevice(blockDeviceList v1alpha1.BlockDeviceList, resourceLists v1.ResourceList) (selected v1alpha1.BlockDevice, foundMatch bool) {
	foundMatch = false
	for _, bd := range blockDeviceList.Items {
		if matchResourceRequirements(bd, resourceLists) {
			selected = *bd.DeepCopy()
			foundMatch = true
			break
		}
	}
	return
}

// matchResourceRequirements matches a block device with a ResourceList
func matchResourceRequirements(bd v1alpha1.BlockDevice, list v1.ResourceList) bool {
	capacity, _ := getRequestedCapacity(list)
	return bd.Spec.Capacity.Storage >= uint64(capacity)
}
