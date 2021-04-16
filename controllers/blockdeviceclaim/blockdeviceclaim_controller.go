/*
Copyright 2021.

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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	controllerutil "github.com/openebs/node-disk-manager/controllers/util"
	"github.com/openebs/node-disk-manager/db/kubernetes"
	"github.com/openebs/node-disk-manager/pkg/select/blockdevice"
	"github.com/openebs/node-disk-manager/pkg/select/verify"
	"github.com/openebs/node-disk-manager/pkg/util"
)

// BlockDeviceClaimReconciler reconciles a BlockDeviceClaim object
type BlockDeviceClaimReconciler struct {
	Client   client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=openebs.io,resources=blockdeviceclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openebs.io,resources=blockdeviceclaims/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openebs.io,resources=blockdeviceclaims/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *BlockDeviceClaimReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("blockdeviceclaim", request.NamespacedName)

	// your logic here
	// Fetch the BlockDeviceClaim instance

	instance := &apis.BlockDeviceClaim{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
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
	case apis.BlockDeviceClaimStatusInvalidCapacity:
		// migrating state to Pending if in InvalidCapacity state.
		// The InvalidCapacityState is deprecated and pending will be used.
		// InvalidCapacity will be the reason for why the BDC is in Pending state.
		instance.Status.Phase = apis.BlockDeviceClaimStatusPending
		err := r.updateClaimStatus(apis.BlockDeviceClaimStatusPending, instance)
		if err != nil {
			klog.Errorf("error in updating phase to pending from invalid capacity for %s: %v", instance.Name, err)
		}
		klog.Infof("%s claim phase is: %s", instance.Name, instance.Status.Phase)
	case apis.BlockDeviceClaimStatusDone:
		err := r.FinalizerHandling(instance)
		if err != nil {
			klog.Errorf("Finalizer handling failed for %s: %v", instance.Name, err)
			return reconcile.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// claimDeviceForBlockDeviceClaim is created, try to determine blockdevice which is
// free and has size equal/greater than BlockDeviceClaim request.
func (r *BlockDeviceClaimReconciler) claimDeviceForBlockDeviceClaim(instance *apis.BlockDeviceClaim) error {

	config := blockdevice.NewConfig(&instance.Spec, r.Client)

	// check for capacity only in auto selection
	if !config.ManualSelection {
		// perform verification of the claim, like capacity
		// Get the capacity requested in the claim
		_, err := verify.GetRequestedCapacity(instance.Spec.Resources.Requests)
		if err != nil {
			r.recorder.Eventf(instance, corev1.EventTypeWarning, "InvalidCapacity", "Invalid Capacity requested")
			//Update deviceClaim CR with pending status
			instance.Status.Phase = apis.BlockDeviceClaimStatusPending
			err1 := r.updateClaimStatus(instance.Status.Phase, instance)
			if err1 != nil {
				klog.Errorf("%s requested an invalid capacity: %v", instance.Name, err1)
				return err1
			}
			klog.Infof("%s set to Pending due to invalid capacity request", instance.Name)
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

	return r.updateClaimStatus(instance.Status.Phase, instance)
}

// FinalizerHandling removes the finalizer from the claim resource
func (r *BlockDeviceClaimReconciler) FinalizerHandling(instance *apis.BlockDeviceClaim) error {

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
		if err := r.Client.Update(context.TODO(), instance); err != nil {
			klog.Errorf("Error removing finalizer from %s", instance.Name)
			r.recorder.Eventf(instance, corev1.EventTypeWarning, "UpdateOperationFailed", "Unable to remove Finalizer, due to error: %v", err.Error())
			return err
		}
	}

	return nil
}

func (r *BlockDeviceClaimReconciler) updateClaimStatus(phase apis.DeviceClaimPhase,
	instance *apis.BlockDeviceClaim) error {
	switch phase {
	case apis.BlockDeviceClaimStatusDone:
		instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, controllerutil.BlockDeviceClaimFinalizer)
		r.recorder.Eventf(instance, corev1.EventTypeNormal, "BlockDeviceClaimBound", "BlockDeviceClaim is bound to %v", instance.Spec.BlockDeviceName)

	}
	// Update BlockDeviceClaim CR
	err := r.Client.Update(context.TODO(), instance)
	if err != nil {
		return fmt.Errorf("error updating status of BDC : %s, %v", instance.ObjectMeta.Name, err)
	}

	return nil
}

// isDeviceRequestedByThisDeviceClaim checks whether a claimed block device belongs to the given BDC
func (r *BlockDeviceClaimReconciler) isDeviceRequestedByThisDeviceClaim(
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
func (r *BlockDeviceClaimReconciler) releaseClaimedBlockDevice(
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
	for i := range bdList.Items {
		// Found a blockdevice ObjRef with BlockDeviceClaim, Clear
		// ObjRef and mark blockdevice released in etcd
		if r.isDeviceRequestedByThisDeviceClaim(instance, bdList.Items[i]) {
			claimedBd = &bdList.Items[i]
			break
		}
	}
	// This case occurs when a claimed BD is manually deleted by removing the finalizer.
	// If this check is not performed, the NDM operator will continuously crash, because it
	// will try to release a non existent BD.
	if claimedBd == nil {
		r.recorder.Eventf(instance, corev1.EventTypeWarning, "BlockDeviceNotFound", "BlockDevice %s not found for releasing", instance.Spec.BlockDeviceName)
		klog.Errorf("could not find blockdevice for claim: %s", instance.Name)
		return fmt.Errorf("blockdevice: %s not found for releasing from bdc: %s", instance.Spec.BlockDeviceName, instance.Name)
	}

	dvr := claimedBd.DeepCopy()
	dvr.Spec.ClaimRef = nil
	dvr.Status.ClaimState = apis.BlockDeviceReleased

	err = r.Client.Update(context.TODO(), dvr)
	if err != nil {
		klog.Errorf("Error updating ClaimRef of %s: %v", dvr.Name, err)
		return err
	}
	r.recorder.Eventf(dvr, corev1.EventTypeNormal, "BlockDeviceCleanUpInProgress", "Released from BDC: %v", instance.Name)

	return nil
}

// claimBlockDevice is used to claim the passed on blockdevice
func (r *BlockDeviceClaimReconciler) claimBlockDevice(bd *apis.BlockDevice, instance *apis.BlockDeviceClaim) error {
	claimRef, err := reference.GetReference(r.Scheme, instance)
	if err != nil {
		return fmt.Errorf("error getting claim reference for BDC:%s, %v", instance.ObjectMeta.Name, err)
	}
	// add finalizer to BlockDevice to prevent accidental deletion of BD
	bd.Finalizers = append(bd.Finalizers, controllerutil.BlockDeviceFinalizer)
	bd.Spec.ClaimRef = claimRef
	bd.Status.ClaimState = apis.BlockDeviceClaimed
	err = r.Client.Update(context.TODO(), bd)
	if err != nil {
		return fmt.Errorf("error while updating BD:%s, %v", bd.ObjectMeta.Name, err)
	}
	klog.Infof("%s claimed by %s", bd.Name, instance.Name)
	return nil
}

// GetBlockDevice get block device resource from etcd
func (r *BlockDeviceClaimReconciler) GetBlockDevice(name string) (*apis.BlockDevice, error) {
	bd := &apis.BlockDevice{}
	err := r.Client.Get(context.TODO(),
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
func (r *BlockDeviceClaimReconciler) getListofDevices(selector *v1.LabelSelector) (*apis.BlockDeviceList, error) {

	//Initialize a deviceList object.
	listBlockDevice := &apis.BlockDeviceList{
		TypeMeta: v1.TypeMeta{
			Kind:       apis.BlockDeviceResourceKind,
			APIVersion: apis.GroupVersion.Version,
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
	err := r.Client.List(context.TODO(), listBlockDevice, opts)
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

func (r *BlockDeviceClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&apis.BlockDeviceClaim{}).
		Owns(&apis.BlockDeviceClaim{}).
		Complete(r)
}
