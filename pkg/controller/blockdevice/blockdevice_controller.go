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

package blockdevice

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	openebsv1alpha1 "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/cleaner"
	controllerutil "github.com/openebs/node-disk-manager/pkg/controller/util"
	"github.com/openebs/node-disk-manager/pkg/util"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new BlockDevice Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBlockDevice{client: mgr.GetClient(), scheme: mgr.GetScheme(), recorder: mgr.GetEventRecorderFor("blockdevice-controller")}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("blockdevice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource BlockDevice
	return c.Watch(&source.Kind{Type: &openebsv1alpha1.BlockDevice{}}, &handler.EnqueueRequestForObject{})
}

var _ reconcile.Reconciler = &ReconcileBlockDevice{}

// ReconcileBlockDevice reconciles a BlockDevice object
type ReconcileBlockDevice struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a BlockDevice object and makes changes based on the state read
// and what is in the BlockDevice.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBlockDevice) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the BlockDevice instance
	instance := &openebsv1alpha1.BlockDevice{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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
	case openebsv1alpha1.BlockDeviceReleased:
		klog.V(2).Infof("%s is in Released state", instance.Name)
		jobController := cleaner.NewJobController(r.client, request.Namespace)
		cleanupTracker := &cleaner.CleanupStatusTracker{JobController: jobController}
		bdCleaner := cleaner.NewCleaner(r.client, request.Namespace, cleanupTracker)
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
			err := r.updateBDStatus(openebsv1alpha1.BlockDeviceUnclaimed, instance)
			if err != nil {
				klog.Errorf("Failed to mark %s as Unclaimed: %v", instance.Name, err)
			}
			r.recorder.Eventf(instance, corev1.EventTypeNormal, "BlockDeviceUnclaimed", "BD now marked as Unclaimed")
		} else {
			r.recorder.Eventf(instance, corev1.EventTypeNormal, "BlockDeviceCleanUpInProgress", "CleanUp is in progress")
		}
	case openebsv1alpha1.BlockDeviceClaimed:
		if !util.Contains(instance.GetFinalizers(), controllerutil.BlockDeviceFinalizer) {
			// finalizer is not present, may be a BlockDevice claimed from previous release
			instance.Finalizers = append(instance.Finalizers, controllerutil.BlockDeviceFinalizer)
			err := r.client.Update(context.TODO(), instance)
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

func (r *ReconcileBlockDevice) updateBDStatus(state openebsv1alpha1.DeviceClaimState, instance *openebsv1alpha1.BlockDevice) error {
	instance.Status.ClaimState = state
	return r.client.Update(context.TODO(), instance)
}

// IsReconcileDisabled is used to check if reconciliation is disabled for
// BlockDevice
func IsReconcileDisabled(bd *openebsv1alpha1.BlockDevice) bool {
	return bd.Annotations[ndm.OpenEBSReconcile] == "false"
}
