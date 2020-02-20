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
	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	openebsv1alpha1 "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/cleaner"
	//corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_device")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new BlockDevice Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBlockDevice{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("blockdevice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource BlockDevice
	err = c.Watch(&source.Kind{Type: &openebsv1alpha1.BlockDevice{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileBlockDevice{}

// ReconcileBlockDevice reconciles a BlockDevice object
type ReconcileBlockDevice struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a BlockDevice object and makes changes based on the state read
// and what is in the BlockDevice.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBlockDevice) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
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
		jobController := cleaner.NewJobController(r.client, request.Namespace)
		cleanupTracker := &cleaner.CleanupStatusTracker{JobController: jobController}
		bdCleaner := cleaner.NewCleaner(r.client, request.Namespace, cleanupTracker)
		ok, err := bdCleaner.Clean(instance)
		if err != nil {
			reqLogger.Error(err, "error while cleaning")
			break
		}
		if ok {
			err := r.updateBDStatus(openebsv1alpha1.BlockDeviceUnclaimed, instance)
			if err != nil {
				reqLogger.Error(err, "marking blockdevice "+instance.Name+" as Unclaimed failed")
			}
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileBlockDevice) updateBDStatus(state openebsv1alpha1.DeviceClaimState, instance *openebsv1alpha1.BlockDevice) error {
	instance.Status.ClaimState = state
	err := r.client.Update(context.TODO(), instance)
	if err != nil {
		return err
	}
	return nil
}

// IsReconcileDisabled is used to check if reconciliation is disabled for
// BlockDevice
func IsReconcileDisabled(bd *openebsv1alpha1.BlockDevice) bool {
	return bd.Annotations[ndm.OpenEBSReconcile] == "false"
}
