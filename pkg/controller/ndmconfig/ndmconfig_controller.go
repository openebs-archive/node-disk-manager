package ndmconfig

import (
	"context"
	"fmt"
	"strings"

	openebsv1alpha1 "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	//"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"github.com/go-logr/logr"
	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_ndmconfig")

const (
	FinalizerName = metav1.FinalizerDeleteDependents
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new NdmConfig Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNdmConfig{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("ndmconfig-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource NdmConfig
	err = c.Watch(&source.Kind{Type: &openebsv1alpha1.NdmConfig{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileNdmConfig{}

// ReconcileNdmConfig reconciles a NdmConfig object
type ReconcileNdmConfig struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a NdmConfig object and makes changes based on the state read
// and what is in the NdmConfig.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNdmConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling NdmConfig")

	// Fetch the NdmConfig instance
	instance := &openebsv1alpha1.NdmConfig{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Error(err, "Could not found ndmconfig instance", "ndmconfig-CR", request.Name)
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.Status.Phase == openebsv1alpha1.NdmConfigPhaseInit {
		reqLogger.Info("ndmconfigInitHandler getting executed")
		err = r.ndmconfigInitHandler(instance, reqLogger)
		if err != nil {
			reqLogger.Error(err, "Error while handing Init", "ndmconfig-CR:", instance.ObjectMeta.Name)
			return reconcile.Result{}, err
		}
	} else if instance.Status.Phase == openebsv1alpha1.NdmConfigPhaseDone {
		if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
			reqLogger.Info("Deletion Time Stamp set on ndmconfig-CR")
			reqLogger.Info("ndmconfigDoneHandler getting executed")
			err = r.ndmconfigDeleteHandler(instance, reqLogger)
			if err != nil {
				reqLogger.Error(err, "Error while handing Delete", "ndmconfig-CR:", instance.ObjectMeta.Name)
				return reconcile.Result{}, err
			}
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileNdmConfig) ndmconfigInitHandler(instance *openebsv1alpha1.NdmConfig,
	reqLogger logr.Logger) error {

	err := r.updateFinalizeronNDMDaemonSetPod(instance, true, reqLogger)
	if err != nil {
		reqLogger.Error(err, "Error while updating finalizer on DaemonSet pod")
		return err
	}

	instance_cpy := instance.DeepCopy()
	instance_cpy.ObjectMeta.Finalizers = append(instance_cpy.ObjectMeta.Finalizers, FinalizerName)
	instance_cpy.Status.Phase = openebsv1alpha1.NdmConfigPhaseDone
	err = r.client.Update(context.TODO(), instance_cpy)

	if err != nil {
		reqLogger.Error(err, "Error while updating", "ndmconfig-CR:", instance.ObjectMeta.Name)
		return err
	}

	err = r.ndmconfigReconcileDiskHandler(instance, ndm.NDMActive, reqLogger)
	if err != nil {
		return err
	}

	err = r.ndmconfigReconcileDeviceHandler(instance, ndm.NDMActive, reqLogger)
	return err
}

func (r *ReconcileNdmConfig) ndmconfigDeleteHandler(instance *openebsv1alpha1.NdmConfig,
	reqLogger logr.Logger) error {

	err := r.ndmconfigReconcileDiskHandler(instance, ndm.NDMUnknown, reqLogger)
	if err != nil {
		return err
	}

	err = r.ndmconfigReconcileDeviceHandler(instance, ndm.NDMUnknown, reqLogger)
	if err != nil {
		return err
	}

	instance_cpy := instance.DeepCopy()
	// Remove finalizer from list and update it.
	instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, FinalizerName)
	err = r.client.Update(context.TODO(), instance_cpy)
	if err != nil {
		reqLogger.Error(err, "Error while updating", "ndmconfig-CR:", instance.ObjectMeta.Name)
		return err
	}

	err = r.updateFinalizeronNDMDaemonSetPod(instance, false, reqLogger)
	if err != nil {
		reqLogger.Error(err, "Error while updating finalizer on DaemonSet pod")
	}
	return err
}

func (r *ReconcileNdmConfig) ndmconfigReconcileDiskHandler(instance *openebsv1alpha1.NdmConfig, state string,
	reqLogger logr.Logger) error {

	//Get Disk list for particular host
	listDiskR, err := r.getListofDisksOnHost(instance, reqLogger)
	if err != nil {
		return err
	}

	// Mark all disks Active
	for _, item := range listDiskR.Items {

		if item.Status.State == state {
			continue
		}

		diskcpyR := item.DeepCopy()
		diskcpyR.Status.State = state
		err := r.client.Update(context.TODO(), diskcpyR)
		if err != nil {
			reqLogger.Error(err, "Error while updating state", "Disk-CR:", diskcpyR.ObjectMeta.Name, "State", state)
			return err
		}
	}
	return nil
}

func (r *ReconcileNdmConfig) getListofDisksOnHost(instance *openebsv1alpha1.NdmConfig,
	reqLogger logr.Logger) (*openebsv1alpha1.DiskList, error) {

	//Initialize a diskList object.
	listDiskR := &openebsv1alpha1.DiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Disk",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	//Set filter option, in our case we are filtering based on hostname/node
	opts := &client.ListOptions{}
	hostName := instance.ObjectMeta.Labels[ndm.NDMHostKey]
	filter := ndm.NDMHostKey + "=" + hostName
	//reqLogger.Info("Filter string", "filter:", filter, "instance:", instance)

	opts.SetLabelSelector(filter)

	//Fetch diskList with matching criteria
	err := r.client.List(context.TODO(), opts, listDiskR)
	if err != nil {
		reqLogger.Error(err, "Error getting DiskList", "ndmconfig-CR:", instance.ObjectMeta.Name)
		return nil, err
	}

	//Check if listDVR is null or not
	length := len(listDiskR.Items)
	if length == 0 {
		err = fmt.Errorf("No disk found with matching criteria")
		return nil, err
	}
	return listDiskR, nil
}

func (r *ReconcileNdmConfig) ndmconfigReconcileDeviceHandler(instance *openebsv1alpha1.NdmConfig, state string,
	reqLogger logr.Logger) error {
	//Get Device list for particular host
	listDVR, err := r.getListofDevicesOnHost(instance, reqLogger)
	if err != nil {
		return err
	}

	// Mark all device as Active
	for _, item := range listDVR.Items {

		if item.Status.State == state {
			continue
		}

		dvr := item.DeepCopy()
		dvr.Status.State = state
		err := r.client.Update(context.TODO(), dvr)
		if err != nil {
			reqLogger.Error(err, "Error while updating state", "Device-CR:", dvr.ObjectMeta.Name, "State", state)
			return err
		}
	}
	return nil
}

func (r *ReconcileNdmConfig) getListofDevicesOnHost(instance *openebsv1alpha1.NdmConfig,
	reqLogger logr.Logger) (*openebsv1alpha1.DeviceList, error) {

	//Initialize a deviceList object.
	listDVR := &openebsv1alpha1.DeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Device",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	//Set filter option, in our case we are filtering based on hostname/node
	opts := &client.ListOptions{}
	hostName := instance.ObjectMeta.Labels[ndm.NDMHostKey]
	filter := ndm.NDMHostKey + "=" + hostName
	//reqLogger.Info("Filter string", "filter:", filter, "instance:", instance)

	opts.SetLabelSelector(filter)

	//Fetch deviceList with matching criteria
	err := r.client.List(context.TODO(), opts, listDVR)
	if err != nil {
		reqLogger.Error(err, "Error getting DeviceList", "ndmconfig-CR:", instance.ObjectMeta.Name)
		return nil, err
	}

	//Check if listDVR is null or not
	length := len(listDVR.Items)
	if length == 0 {
		err = fmt.Errorf("No device found with matching criteria")
		return nil, err
	}
	return listDVR, nil
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

func (r *ReconcileNdmConfig) updateFinalizeronNDMDaemonSetPod(instance *openebsv1alpha1.NdmConfig, is_add bool,
	reqLogger logr.Logger) error {

	var ns string
	name := strings.TrimPrefix(instance.Name, ndm.NDMConfigPreFix)
	if instance.ObjectMeta.Namespace == "" {
		ns = "default"
	} else {
		ns = instance.ObjectMeta.Namespace
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}

	err := r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Error(err, "Could not found pod", "Pod:", name)
			return err
		}
		reqLogger.Error(err, "Error while getting pod", "Pod:", name)
		return err
	}

	pod_cpy := pod.DeepCopy()
	if is_add == true {
		pod_cpy.ObjectMeta.Finalizers = append(pod_cpy.ObjectMeta.Finalizers, FinalizerName)
	} else {
		pod_cpy.ObjectMeta.Finalizers = removeString(pod_cpy.ObjectMeta.Finalizers, FinalizerName)
	}
	err = r.client.Update(context.TODO(), pod_cpy)
	if err != nil {
		reqLogger.Error(err, "Error while setting finalizer on pod", "Pod:", name)
	}
	return err
}
