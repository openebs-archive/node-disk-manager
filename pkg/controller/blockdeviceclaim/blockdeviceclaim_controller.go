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
	"github.com/openebs/node-disk-manager/db/kubernetes"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	controllerutil "github.com/openebs/node-disk-manager/pkg/controller/util"
	"github.com/openebs/node-disk-manager/pkg/select/blockdevice"
	"github.com/openebs/node-disk-manager/pkg/select/verify"
	"github.com/openebs/node-disk-manager/pkg/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new BlockDeviceClaim Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBlockDeviceClaim{client: mgr.GetClient(), scheme: mgr.GetScheme(), recorder: mgr.GetRecorder("blockdeviceclaim-operator")}
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
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a BlockDeviceClaim object and makes changes based on the state read
// and what is in the BlockDeviceClaim.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBlockDeviceClaim) Reconcile(request reconcile.Request) (reconcile.Result, error) {
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

	// check if reconciliation is disabled for this resource
	if IsReconcileDisabled(instance) {
		return reconcile.Result{}, nil
	}

	switch instance.Status.Phase {
	case apis.BlockDeviceClaimStatusPending:
		fallthrough
	case apis.BlockDeviceClaimStatusEmpty:
		klog.Infof("BDC %s claim phase is: %s", instance.Name, instance.Status.Phase)
		// claim the BD only if deletion time stamp is not set.
		// since BDC can now have multiple finalizers, we should not claim a
		// BD if its deletiontime stamp is set.
		if instance.DeletionTimestamp.IsZero() {
			err := r.claimDeviceForBlockDeviceClaim(instance)
			if err != nil {
				klog.Errorf("%s failed to claim: %v", instance.Name, err)
				return reconcile.Result{}, err
			}
		}
		// TODO @akhilerm this phase should be moved out from ClaimPhase and will be a reason for not claiming
	case apis.BlockDeviceClaimStatusInvalidCapacity:
		// currently for invalid capacity, the BDC will remain in that state
		klog.Infof("%s claim phase is: %s", instance.Name, instance.Status.Phase)
	case apis.BlockDeviceClaimStatusDone:
		err := r.FinalizerHandling(instance)
		if err != nil {
			klog.Errorf("Finalizer handling failed for %s: %v", instance.Name, err)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// claimDeviceForBlockDeviceClaim is created, try to determine blockdevice which is
// free and has size equal/greater than BlockDeviceClaim request.
func (r *ReconcileBlockDeviceClaim) claimDeviceForBlockDeviceClaim(instance *apis.BlockDeviceClaim) error {

	config := blockdevice.NewConfig(&instance.Spec, r.client)

	// check for capacity only in auto selection
	if !config.ManualSelection {
		// perform verification of the claim, like capacity
		// Get the capacity requested in the claim
		_, err := verify.GetRequestedCapacity(instance.Spec.Resources.Requests)
		if err != nil {
			r.recorder.Eventf(instance, corev1.EventTypeWarning, "InvalidCapacity", err.Error())
			//Update deviceClaim CR with error string
			instance.Status.Phase = apis.BlockDeviceClaimStatusInvalidCapacity
			err1 := r.updateClaimStatus(instance.Status.Phase, instance)
			if err1 != nil {
				klog.Errorf("%s requested an invalid capacity: %v", instance.Name, err1)
				return err1
			}
			return err
		}
	}

	// create selector from the label selector given in BDC spec.
	selector := generateSelector(*instance)

	// get list of block devices.
	bdList, err := r.getListofDevices(selector)
	if err != nil {
		return err
	}

	selectedDevice, err := config.Filter(bdList)
	if err != nil {
		klog.Errorf("Error selecting device for %s: %v", instance.Name, err)
		r.recorder.Eventf(instance, corev1.EventTypeWarning, "SelectionFailed", err.Error())
		instance.Status.Phase = apis.BlockDeviceClaimStatusPending
	} else {
		instance.Spec.BlockDeviceName = selectedDevice.Name
		instance.Status.Phase = apis.BlockDeviceClaimStatusDone
		err = r.claimBlockDevice(selectedDevice, instance)
		if err != nil {
			return err
		}
		r.recorder.Eventf(selectedDevice, corev1.EventTypeNormal, "BlockDeviceClaimed", "BlockDevice claimed by %v", instance.Name)
		r.recorder.Eventf(instance, corev1.EventTypeNormal, "BlockDeviceClaimed", "BlockDevice: %v claimed", instance.Spec.BlockDeviceName)
	}

	err = r.updateClaimStatus(instance.Status.Phase, instance)
	if err != nil {
		return err
	}

	return nil
}

// FinalizerHandling removes the finalizer from the claim resource
func (r *ReconcileBlockDeviceClaim) FinalizerHandling(instance *apis.BlockDeviceClaim) error {

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		return nil
	}
	// The object is being deleted
	// Check if the BDC has only NDM finalizer present. If yes, it means that the BDC
	// was deleted by the owner itself, and NDM can proceed with releasing the BD and
	// removing the NDM finalizer
	if len(instance.ObjectMeta.Finalizers) == 1 &&
		util.Contains(instance.ObjectMeta.Finalizers, controllerutil.BlockDeviceClaimFinalizer) {
		// Finalizer is set, lets handle external dependency
		if err := r.releaseClaimedBlockDevice(instance); err != nil {
			klog.Errorf("Error releasing claimed block device %s from %s: %v",
				instance.Spec.BlockDeviceName, instance.Name, err)
			return err
		}
		r.recorder.Eventf(instance, corev1.EventTypeNormal, "BlockDeviceReleased", "BlockDevice: %v is released", instance.Spec.BlockDeviceName)

		// Remove finalizer from list and update it.
		instance.ObjectMeta.Finalizers = util.RemoveString(instance.ObjectMeta.Finalizers, controllerutil.BlockDeviceClaimFinalizer)
		if err := r.client.Update(context.TODO(), instance); err != nil {
			klog.Errorf("Error removing finalizer from %s", instance.Name)
			r.recorder.Eventf(instance, corev1.EventTypeWarning, "UpdateOperationFailed", "Unable to remove Finalizer, due to error: %v", err.Error())
			return err
		}
	}

	return nil
}

func (r *ReconcileBlockDeviceClaim) updateClaimStatus(phase apis.DeviceClaimPhase,
	instance *apis.BlockDeviceClaim) error {
	switch phase {
	case apis.BlockDeviceClaimStatusDone:
		instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, controllerutil.BlockDeviceClaimFinalizer)
		r.recorder.Eventf(instance, corev1.EventTypeNormal, "BlockDeviceClaimBound", "BlockDeviceClaim is bound to %v", instance.Spec.BlockDeviceName)

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
	instance *apis.BlockDeviceClaim, item apis.BlockDevice) bool {

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
	instance *apis.BlockDeviceClaim) error {

	klog.Infof("Releasing claimed block device %s from %s", instance.Spec.BlockDeviceName, instance.Name)

	//Get BlockDevice list on all nodes
	//empty selector is used to select everything.
	selector := &v1.LabelSelector{}
	bdList, err := r.getListofDevices(selector)
	if err != nil {
		return err
	}

	// Check if same deviceclaim holding the ObjRef
	var claimedBd *apis.BlockDevice
	for _, item := range bdList.Items {
		// Found a blockdevice ObjRef with BlockDeviceClaim, Clear
		// ObjRef and mark blockdevice released in etcd
		if r.isDeviceRequestedByThisDeviceClaim(instance, item) {
			claimedBd = &item
			break
		}
	}
	dvr := claimedBd.DeepCopy()
	dvr.Spec.ClaimRef = nil
	dvr.Status.ClaimState = apis.BlockDeviceReleased

	err = r.client.Update(context.TODO(), dvr)
	if err != nil {
		klog.Errorf("Error updating ClaimRef of %s: %v", dvr.Name, err)
		return err
	}
	r.recorder.Eventf(dvr, corev1.EventTypeNormal, "BlockDeviceCleanUpInProgress", "Released from BDC: %v", instance.Name)

	return nil
}

// claimBlockDevice is used to claim the passed on blockdevice
func (r *ReconcileBlockDeviceClaim) claimBlockDevice(bd *apis.BlockDevice, instance *apis.BlockDeviceClaim) error {
	claimRef, err := reference.GetReference(r.scheme, instance)
	if err != nil {
		return fmt.Errorf("error getting claim reference for BDC:%s, %v", instance.ObjectMeta.Name, err)
	}
	// add finalizer to BlockDevice to prevent accidental deletion of BD
	bd.Finalizers = append(bd.Finalizers, controllerutil.BlockDeviceFinalizer)
	bd.Spec.ClaimRef = claimRef
	bd.Status.ClaimState = apis.BlockDeviceClaimed
	err = r.client.Update(context.TODO(), bd)
	if err != nil {
		return fmt.Errorf("error while updating BD:%s, %v", bd.ObjectMeta.Name, err)
	}
	klog.Infof("%s claimed by %s", bd.Name, instance.Name)
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
func (r *ReconcileBlockDeviceClaim) getListofDevices(selector *v1.LabelSelector) (*apis.BlockDeviceList, error) {

	//Initialize a deviceList object.
	listBlockDevice := &apis.BlockDeviceList{
		TypeMeta: v1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	opts := &client.ListOptions{}

	if sel, err := v1.LabelSelectorAsSelector(selector); err != nil {
		// if conversion of selector errors out, the list call will be errored
		return nil, err
	} else {
		opts.LabelSelector = sel
	}

	//Fetch deviceList with matching criteria
	err := r.client.List(context.TODO(), opts, listBlockDevice)
	if err != nil {
		return nil, err
	}

	return listBlockDevice, nil
}

// IsReconcileDisabled is used to check if reconciliation is disabled for
// BlockDeviceClaim
func IsReconcileDisabled(bdc *apis.BlockDeviceClaim) bool {
	return bdc.Annotations[ndm.OpenEBSReconcile] == "false"
}

// generateSelector creates the label selector for BlockDevices from
// the BlockDeviceClaim spec
func generateSelector(bdc apis.BlockDeviceClaim) *v1.LabelSelector {
	var hostName string
	// get the hostname
	if len(bdc.Spec.HostName) != 0 {
		hostName = bdc.Spec.HostName
	}
	// the hostname in NodeAttribute will override the hostname in spec, since spec.hostName
	// will be deprecated shortly
	if len(bdc.Spec.BlockDeviceNodeAttributes.HostName) != 0 {
		hostName = bdc.Spec.BlockDeviceNodeAttributes.HostName
	}

	// the hostname label is added into the user given list of labels. If the user hasn't
	// given any selector, then the selector object is initialized.
	selector := bdc.Spec.Selector.DeepCopy()
	if selector == nil {
		selector = &v1.LabelSelector{}
	}
	if selector.MatchLabels == nil {
		selector.MatchLabels = make(map[string]string)
	}

	// if any hostname is provided, add it to selector
	if len(hostName) != 0 {
		selector.MatchLabels[kubernetes.KubernetesHostNameLabel] = hostName
	}
	return selector
}
