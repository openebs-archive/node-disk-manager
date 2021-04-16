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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	controllerutil "github.com/openebs/node-disk-manager/controllers/util"
	"github.com/openebs/node-disk-manager/pkg/cleaner"
	"github.com/openebs/node-disk-manager/pkg/util"
)

// BlockDeviceReconciler reconciles a BlockDevice object
type BlockDeviceReconciler struct {
	Client   client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=openebs.io,resources=blockdevices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openebs.io,resources=blockdevices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openebs.io,resources=blockdevices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *BlockDeviceReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	// Fetch the BlockDevice instance
	instance := &apis.BlockDevice{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
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

	// check if the this block device need to reconciled
	if IsReconcileDisabled(instance) {
		return reconcile.Result{}, nil
	}

	switch instance.Status.ClaimState {
	case apis.BlockDeviceReleased:
		klog.V(2).Infof("%s is in Released state", instance.Name)
		jobController := cleaner.NewJobController(r.Client, request.Namespace)
		cleanupTracker := &cleaner.CleanupStatusTracker{JobController: jobController}
		bdCleaner := cleaner.NewCleaner(r.Client, request.Namespace, cleanupTracker)
		ok, err := bdCleaner.Clean(instance)
		if err != nil {
			klog.Errorf("Error while cleaning %s: %v", instance.Name, err)
			r.recorder.Eventf(instance, corev1.EventTypeWarning, "BlockDeviceCleanUp", "CleanUp unsuccessful, due to error: %v", err)
			break
		}
		if ok {
			r.recorder.Eventf(instance, corev1.EventTypeNormal, "BlockDeviceReleased", "CleanUp Completed")
			// remove the finalizer string from BlockDevice resource
			instance.Finalizers = util.RemoveString(instance.Finalizers, controllerutil.BlockDeviceFinalizer)
			klog.Infof("Cleanup completed for %s", instance.Name)
			err := r.updateBDStatus(apis.BlockDeviceUnclaimed, instance)
			if err != nil {
				klog.Errorf("Failed to mark %s as Unclaimed: %v", instance.Name, err)
			}
			r.recorder.Eventf(instance, corev1.EventTypeNormal, "BlockDeviceUnclaimed", "BD now marked as Unclaimed")
		} else {
			r.recorder.Eventf(instance, corev1.EventTypeNormal, "BlockDeviceCleanUpInProgress", "CleanUp is in progress")
		}
	case apis.BlockDeviceClaimed:
		if !util.Contains(instance.GetFinalizers(), controllerutil.BlockDeviceFinalizer) {
			// finalizer is not present, may be a BlockDevice claimed from previous release
			instance.Finalizers = append(instance.Finalizers, controllerutil.BlockDeviceFinalizer)
			err := r.Client.Update(context.TODO(), instance)
			if err != nil {
				klog.Errorf("Error updating finalizer on %s: %v", instance.Name, err)
			}
			klog.Infof("%s updated with %s finalizer", instance.Name, controllerutil.BlockDeviceFinalizer)
			r.recorder.Eventf(instance, corev1.EventTypeNormal, "BlockDeviceClaimed", "BD Claimed, and finalizer added")
		}
		// if finalizer is already present. do nothing
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BlockDeviceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apis.BlockDevice{}).
		Owns(&apis.BlockDevice{}).
		Complete(r)
}

func (r *BlockDeviceReconciler) updateBDStatus(state apis.DeviceClaimState, instance *apis.BlockDevice) error {
	instance.Status.ClaimState = state
	err := r.Client.Update(context.TODO(), instance)
	if err != nil {
		return err
	}
	return nil
}

// IsReconcileDisabled is used to check if reconciliation is disabled for
// BlockDevice
func IsReconcileDisabled(bd *apis.BlockDevice) bool {
	return bd.Annotations[ndm.OpenEBSReconcile] == "false"
}
