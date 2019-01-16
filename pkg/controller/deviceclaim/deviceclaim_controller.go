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

	//	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	//	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"github.com/go-logr/logr"
	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_deviceclaim")

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
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
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

	} else if instance.Status.Phase == openebsv1alpha1.DeviceClaimStatusDone {

	}

	return reconcile.Result{}, nil
}

/*
 * New DeviceClaim CR is created, try to findout device which is
 * free and has size equal/greater than DeviceClaim request.
 */
func (r *ReconcileDeviceClaim) claimDeviceForDeviceClaimCR(
	instance *openebsv1alpha1.DeviceClaim, reqLogger logr.Logger) error {

	//Initialize an deviceList object.
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

	//Find a device which is free and have available
	//space more than or equal to requested
	for _, item := range listDVR.Items {
		if (strings.Compare(item.Status.State, ndm.NDMFree) == 0) &&
			(item.Spec.Capacity.Storage >= instance.Spec.Capacity) {

			reqLogger.Info("Found matching device", "Device Name:",
				item.ObjectMeta.Name, "Device Capacity:",
				item.Spec.Capacity.Storage)

			//Found a device, claim it, update ClaimInfo into etcd
			dvr := item.DeepCopy()
			dvr.Claim.APIVersion = instance.TypeMeta.APIVersion
			dvr.Claim.Kind = instance.TypeMeta.Kind
			dvr.Claim.Name = instance.ObjectMeta.Name
			dvr.Claim.DeviceClaimUID = instance.ObjectMeta.UID
			dvr.Status.State = ndm.NDMUsed
			err := r.client.Update(context.TODO(), dvr)
			if err != nil {
				reqLogger.Info("Error while updating device CR")
				return err
			}

			//Update deviceClaim CR to show that device claim happened
			instance_cpy := instance.DeepCopy()
			instance_cpy.Status.Phase = openebsv1alpha1.DeviceClaimStatusDone
			err = r.client.Update(context.TODO(), instance_cpy)
			if err != nil {
				reqLogger.Info("Error while updating deviceClaim CR")
				return err
			}
			return nil
		}
	}
	return err
}
