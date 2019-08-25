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

	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/select/blockdevice"
	"github.com/openebs/node-disk-manager/pkg/select/verify"
	"github.com/openebs/node-disk-manager/pkg/util"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/reference"

	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_deviceclaim")

const (
	// BlockDeviceClaimFinalizer is the finalizer name for the block device claim
	BlockDeviceClaimFinalizer = "openebs.io/bdc-protection"
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
	err = c.Watch(&source.Kind{Type: &apis.BlockDeviceClaim{}}, &handler.EnqueueRequestForObject{})
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

	// Fetch the BlockDeviceClaim instance
	instance := &apis.BlockDeviceClaim{}
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

	switch instance.Status.Phase {
	case apis.BlockDeviceClaimStatusPending:
		fallthrough
	case apis.BlockDeviceClaimStatusEmpty:
		reqLogger.Info("BlockDeviceClaim State is:" + string(instance.Status.Phase))
		// claim the BD only if deletion time stamp is not set.
		// since BDC can now have multiple finalizers, we should not claim a
		// BD if its deletiontime stamp is set.
		if instance.DeletionTimestamp.IsZero() {
			err := r.claimDeviceForBlockDeviceClaim(instance, reqLogger)
			if err != nil {
				reqLogger.Error(err, "BlockDeviceClaim "+instance.ObjectMeta.Name+" failed")
				return reconcile.Result{}, err
			}
		}
	case apis.BlockDeviceClaimStatusInvalidCapacity:
		// currently for invalid capacity, the BDC will remain in that state
		reqLogger.Info("BlockDeviceClaim State is:" + string(instance.Status.Phase))
	case apis.BlockDeviceClaimStatusDone:
		err := r.FinalizerHandling(instance, reqLogger)
		if err != nil {
			reqLogger.Error(err, "Finalizer handling failed", "BlockDeviceClaim-CR:", instance.ObjectMeta.Name)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// claimDeviceForBlockDeviceClaim is created, try to findout blockdevice which is
// free and has size equal/greater than BlockDeviceClaim request.
func (r *ReconcileBlockDeviceClaim) claimDeviceForBlockDeviceClaim(
	instance *apis.BlockDeviceClaim, reqLogger logr.Logger) error {

	config := blockdevice.NewConfig(&instance.Spec, r.client)

	// check for capacity only in auto selection
	if !config.ManualSelection {
		// perform verification of the claim, like capacity
		// Get the capacity requested in the claim
		_, err := verify.GetRequestedCapacity(instance.Spec.Resources.Requests)
		if err != nil {
			//Update deviceClaim CR with error string
			instance.Status.Phase = apis.BlockDeviceClaimStatusInvalidCapacity
			err1 := r.updateClaimStatus(instance.Status.Phase, instance)
			if err1 != nil {
				reqLogger.Error(err1, "Invalid Capacity requested")
				return err1
			}
			return err
		}
	}

	//select block device from list of devices.
	bdList, err := r.getListofDevices(instance.Spec.HostName, config.ManualSelection)
	if err != nil {
		return err
	}

	selectedDevice, err := config.Filter(bdList)
	if err != nil {
		reqLogger.Error(err, "Error selecting device")
		instance.Status.Phase = apis.BlockDeviceClaimStatusPending
	} else {
		instance.Spec.BlockDeviceName = selectedDevice.Name
		instance.Status.Phase = apis.BlockDeviceClaimStatusDone
		err = r.claimBlockDevice(selectedDevice, instance, reqLogger)
		if err != nil {
			return err
		}
	}

	err = r.updateClaimStatus(instance.Status.Phase, instance)
	if err != nil {
		return err
	}
	return nil
}

// FinalizerHandling removes the finalizer from the claim resource
func (r *ReconcileBlockDeviceClaim) FinalizerHandling(
	instance *apis.BlockDeviceClaim, reqLogger logr.Logger) error {

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		return nil
	}
	// The object is being deleted
	// Check if the BDC has only NDM finalizer present. If yes, it means that the BDC
	// was deleted by the owner itself, and NDM can proceed with releasing the BD and
	// removing the NDM finalizer
	if len(instance.ObjectMeta.Finalizers) == 1 &&
		util.Contains(instance.ObjectMeta.Finalizers, BlockDeviceClaimFinalizer) {
		// Finalizer is set, lets handle external dependency
		if err := r.releaseClaimedBlockDevice(instance, reqLogger); err != nil {
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

func (r *ReconcileBlockDeviceClaim) updateClaimStatus(phase apis.DeviceClaimPhase,
	instance *apis.BlockDeviceClaim) error {
	switch phase {
	case apis.BlockDeviceClaimStatusDone:
		instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, BlockDeviceClaimFinalizer)
	}
	// Update BlockDeviceClaim CR
	err := r.client.Update(context.TODO(), instance)
	if err != nil {
		return fmt.Errorf("error updating status of BDC : %s, %v", instance.ObjectMeta.Name, err)
	}

	return nil
}

// isDeviceRequestedByThisDeviceClaim checks whether a claimed block device belongs to the given BDC
func (r *ReconcileBlockDeviceClaim) isDeviceRequestedByThisDeviceClaim(
	instance *apis.BlockDeviceClaim, item apis.BlockDevice,
	reqLogger logr.Logger) bool {

	if item.Status.ClaimState != apis.BlockDeviceClaimed {
		return false
	}

	if item.Spec.ClaimRef.Name != instance.ObjectMeta.Name {
		return false
	}

	if item.Spec.ClaimRef.UID != instance.ObjectMeta.UID {
		return false
	}

	if item.Spec.ClaimRef.Kind != instance.TypeMeta.Kind {
		return false
	}
	return true
}

// releaseClaimedBlockDevice releases the block device claimed by this BlockDeviceClaim
func (r *ReconcileBlockDeviceClaim) releaseClaimedBlockDevice(
	instance *apis.BlockDeviceClaim, reqLogger logr.Logger) error {

	reqLogger.Info("Deleting external dependencies for CR:" + instance.Name)

	manualSelection := false
	if instance.Spec.BlockDeviceName != "" {
		manualSelection = true
	}

	//Get BlockDevice list for particular host
	listDVR, err := r.getListofDevices(instance.Spec.HostName, manualSelection)
	if err != nil {
		return err
	}

	// Check if same deviceclaim holding the ObjRef
	for _, item := range listDVR.Items {
		if !r.isDeviceRequestedByThisDeviceClaim(instance, item, reqLogger) {
			continue
		}

		// Found a blockdevice ObjRef with BlockDeviceClaim, Clear
		// ObjRef and mark blockdevice released in etcd
		dvr := item.DeepCopy()
		dvr.Spec.ClaimRef = nil
		dvr.Status.ClaimState = apis.BlockDeviceReleased
		err := r.client.Update(context.TODO(), dvr)
		if err != nil {
			reqLogger.Error(err, "Error while updating ObjRef", "BlockDevice-CR:", dvr.ObjectMeta.Name)
			return err
		}
	}
	return nil
}

// claimBlockDevice is used to claim the passed on blockdevice
func (r *ReconcileBlockDeviceClaim) claimBlockDevice(bd *apis.BlockDevice,
	instance *apis.BlockDeviceClaim, reqLogger logr.Logger) error {
	claimRef, err := reference.GetReference(r.scheme, instance)
	if err != nil {
		return fmt.Errorf("error getting claim reference for BDC:%s, %v", instance.ObjectMeta.Name, err)
	}
	bd.Spec.ClaimRef = claimRef
	bd.Status.ClaimState = apis.BlockDeviceClaimed
	err = r.client.Update(context.TODO(), bd)
	if err != nil {
		return fmt.Errorf("error while updating BD:%s, %v", bd.ObjectMeta.Name, err)
	}
	reqLogger.Info("Block Device " + bd.Name + " claimed")
	return nil
}

// GetBlockDevice get block device resource from etcd
func (r *ReconcileBlockDeviceClaim) GetBlockDevice(name string) (*apis.BlockDevice, error) {
	bd := &apis.BlockDevice{}
	err := r.client.Get(context.TODO(),
		client.ObjectKey{Namespace: "", Name: name}, bd)

	if err != nil {
		return nil, err
	}
	return bd, nil
}

// getListofDevices gets the list of block devices on the node to which BlockDeviceClaim is made
// TODO:
//  ListBlockDeviceResource in package cmd/ndm_daemonset/controller has the same functionality.
//  Need to merge these 2 functions.
func (r *ReconcileBlockDeviceClaim) getListofDevices(hostName string, ManualSelection bool) (*apis.BlockDeviceList, error) {

	//Initialize a deviceList object.
	listBlockDevice := &apis.BlockDeviceList{
		TypeMeta: v1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	opts := &client.ListOptions{}
	// Set filter option, in our case we are filtering based on hostname/node
	// Filter is only applied when its auto selection. In Manual selection, all
	// blockdevices are listed.
	// TODO for manual selection, instead of listing all BDs, only get the BD with given name
	if !ManualSelection {
		filter := ndm.KubernetesHostNameLabel + "=" + hostName
		opts.SetLabelSelector(filter)
	}

	//Fetch deviceList with matching criteria
	err := r.client.List(context.TODO(), opts, listBlockDevice)
	if err != nil {
		return nil, err
	}

	return listBlockDevice, nil
}
