package disk

import (
	"context"
	"fmt"
	"strings"
	"time"

	ndm "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	openebsv1alpha1 "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"

	//corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	//"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"github.com/go-logr/logr"
	"github.com/openebs/node-disk-manager/pkg/common"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_disk")

// If node not see for 10Minute, we treat that node crashed/unreachable.
var expirationeInterval = time.Duration(10 * time.Minute)

// For initial scan of nodes, we use this flag
var initialNodeCheckDone bool

//These two time variables used to control frquency
//of operation which check if node is stale or not.
//Since Reconcilation would be triggered every now and then,
//we need these counters.
var isFirstTimeLivenessCheck bool
var lastNodeLivenessCheckTime time.Time
var nodeLivenessCheckInterval = time.Duration(10 * time.Minute)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Disk Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDisk{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("disk-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Disk
	err = c.Watch(&source.Kind{Type: &openebsv1alpha1.Disk{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDisk{}

// ReconcileDisk reconciles a Disk object
type ReconcileDisk struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Disk object and makes changes based on the state read
// and what is in the Disk.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileDisk) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Disk")

	if !isFirstTimeLivenessCheck {
		lastNodeLivenessCheckTime = time.Now()
		isFirstTimeLivenessCheck = true
	}

	if lastNodeLivenessCheckTime.Add(nodeLivenessCheckInterval).Before(time.Now()) {
		fmt.Println("Node liveness check triggered")
		lastNodeLivenessCheckTime = time.Now()

		//TODO: For now grace this error
		err := r.FindInactiveNodesMarkDiskInactive(reqLogger)
		if err != nil {
			fmt.Println("While checking liveness of nodes, hit err:", err)
		}
	}

	// Fetch the Disk instance
	instance := &openebsv1alpha1.Disk{}
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
	return reconcile.Result{}, nil
}

// Check if any node is inactive/crashed and based on it, update disk CR status "Inactive"
func (r *ReconcileDisk) FindInactiveNodesMarkDiskInactive(reqLogger logr.Logger) error {
	var StaleNodeList []string

	// If this is the first time after crash/reboot,
	// check if we have any stale node.
	//Read all disk CR and prepare a list of nodes.
	if !initialNodeCheckDone {
		nodeList, err := r.FindNodeList(reqLogger)
		if err != nil {
			reqLogger.Info("Error while finding node list")
			return err
		}

		if len(nodeList) != 0 {
			fmt.Println("nodeList:", nodeList)
		}

		initialNodeCheckDone = true
		// Check if any of node is inactive/stale by looking into NodeLiveInfo map
		// which is update after an interval by Node-Disk-Manager Daemonset
		StaleNodeList = common.FindInactiveNodes(expirationeInterval, nodeList)
	} else {
		StaleNodeList = common.FindInactiveNodes(expirationeInterval, nil)
	}

	if len(StaleNodeList) == 0 {
		fmt.Println("No Stale Node found")
		return nil
	}

	fmt.Println("StaleNodeList:", StaleNodeList)

	for _, hostName := range StaleNodeList {
		//Initialize a disk List object.
		listDiskR := &openebsv1alpha1.DiskList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Disk",
				APIVersion: "openebs.io/v1alpha1",
			},
		}

		// Filter disks based on node name
		opts := &client.ListOptions{}
		filter := ndm.NDMHostKey + "=" + hostName
		opts.SetLabelSelector(filter)

		//Fetch disk list
		err := r.client.List(context.TODO(), opts, listDiskR)
		if err != nil {
			reqLogger.Info("Error while getting disk list")
			return err
		}

		//Update disk status to Inactive for all disks owned by this node
		for _, item := range listDiskR.Items {
			if strings.Compare(item.Status.State, ndm.NDMInactive) != 0 {
				dcpyItem := item.DeepCopy()
				dcpyItem.Status.State = ndm.NDMInactive
				err := r.client.Update(context.TODO(), dcpyItem)
				if err != nil {
					fmt.Println("Error updating disk:", item.ObjectMeta.Name)
					return err
				}
			}
		}
	}
	return nil
}

func (r *ReconcileDisk) FindNodeList(reqLogger logr.Logger) ([]string, error) {

	var nodeMap = make(map[string]string)
	var nodeList []string

	//Initialize a disk List object.
	listDiskR := &openebsv1alpha1.DiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Disk",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	opts := &client.ListOptions{}

	//Fetch disk list
	err := r.client.List(context.TODO(), opts, listDiskR)
	if err != nil {
		reqLogger.Info("Error while getting disk list")
		return nil, err
	}

	for _, item := range listDiskR.Items {
		//fmt.Println("Found Host:", item.ObjectMeta.Labels[ndm.NDMHostKey])
		nodeMap[item.ObjectMeta.Labels[ndm.NDMHostKey]] = ""
	}

	for k := range nodeMap {
		//fmt.Println("Key:", k, "Value:", nodeMap[k])
		nodeList = append(nodeList, k)
	}

	return nodeList, nil
}
