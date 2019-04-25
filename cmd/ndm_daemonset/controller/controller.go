/*
Copyright 2018 OpenEBS Authors.

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

package controller

import (
	"errors"
	"os"
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/pkg/apis"
	"github.com/openebs/node-disk-manager/pkg/signals"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	// FalseString contains string value of false
	FalseString = "false"
	// TrueString contains string value of true
	TrueString = "true"
	// NDMDiskKind is the Disk kind CR.
	NDMDiskKind = "Disk"
	// NDMBlockDeviceKind is the Device kind CR.
	NDMBlockDeviceKind = "BlockDevice"
	// NDMVersion is the CR version.
	NDMVersion = "openebs.io/v1alpha1"
	// NDMHostKey is host name label prefix.
	NDMHostKey = "kubernetes.io/hostname"
	// NDMFree is constant for free/available resource status.
	NDMUnclaimed = "Unclaimed"
	// NDMUsed is constant for in-use resource status.
	NDMClaimed = "Claimed"
	// NDMNotPartioned is used to say blockdevice does not have any partition.
	NDMNotPartitioned = "No"
	// NDMPartitioned is used to say blockdevice has some partitions.
	NDMPartitioned = "Yes"
	// NDMActive is constant for active resource status.
	NDMActive = "Active"
	// NDMInactive is constant for inactive resource status.
	NDMInactive = "Inactive"
	// NDMUnknown is constant for resource unknown satus.
	NDMUnknown = "Unknown"
	// NDMDiskTypeKey specifies the type of disk.
	NDMDiskTypeKey   = "ndm.io/disk-type"
	NDMDeviceTypeKey = "ndm.io/blockdevice-type"
	// NDMUnmanagedDiskKey specifies disk cr should be managed by ndm or not.
	NDMManagedKey = "ndm.io/managed"
)

const (
	// NDMDefaultDiskType will be used to initialize the disk type.
	NDMDefaultDiskType = "disk"
	// NDMDefaultDeviceType will be used to initialize the blockdevice type.
	NDMDefaultDeviceType = "blockdevice"
)

const (
	// CRDRetryInterval is used if CRD is not present.
	CRDRetryInterval = 10 * time.Second
)

// ControllerBroadcastChannel is used to send a copy of controller object to each probe.
// Each probe can get the copy of controller struct any time they need to read the channel.
var ControllerBroadcastChannel = make(chan *Controller)

// Controller is the controller implementation for disk resources
type Controller struct {
	HostName      string               // HostName is host name in which disk is attached
	KubeClientset kubernetes.Interface // KubeClientset is standard kubernetes clientset
	mgr           manager.Manager
	Clientset     client.Client
	NDMConfig     *NodeDiskManagerConfig // NDMConfig contains custom config for ndm
	Mutex         *sync.Mutex            // Mutex is used to lock and unlock Controller
	Filters       []*Filter              // Filters are the registered filters like os disk filter
	Probes        []*Probe               // Probes are the registered probes like udev/smart
}

// NewController returns a controller pointer for any error case it will return nil
func NewController(kubeconfig string) (*Controller, error) {
	controller := &Controller{}
	cfg, err := getCfg(kubeconfig)
	if err != nil {
		return nil, err
	}
	if err := controller.setKubeClient(cfg); err != nil {
		return nil, err
	}
	//if err := controller.setClientSet(cfg); err != nil {
	//	return nil, err
	//}
	if err := controller.setNodeName(); err != nil {
		return nil, err
	}
	controller.SetNDMConfig()
	controller.Filters = make([]*Filter, 0)
	controller.Probes = make([]*Probe, 0)
	controller.Mutex = &sync.Mutex{}

	// Create a new Cmd to provide shared dependencies and start components
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil && namespace == "" {
		namespace = "default"
	} else if err != nil {
		return controller, err
	}

	controller.mgr, err = manager.New(cfg, manager.Options{Namespace: namespace})
	if err != nil {
		return controller, err
	}

	// Setup Scheme for all resources
	if err := apis.AddToScheme(controller.mgr.GetScheme()); err != nil {
		return controller, err
	}

	controller.Clientset, err = client.New(cfg, client.Options{})
	if err != nil {
		return controller, err
	}

	controller.WaitForDiskCRD()
	controller.WaitForBlockDeviceCRD()
	return controller, nil
}

// getCfg returns incluster or out cluster config using
// incluster config or kubeconfig
func getCfg(kubeconfig string) (*rest.Config, error) {
	masterURL := ""
	cfg, err := rest.InClusterConfig()
	if err == nil {
		return cfg, err
	}
	if kubeconfig == "" {
		return nil, errors.New("kubeconfig is empty")
	}
	cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		return nil, err
	}
	return cfg, err
}

// setKubeClient set KubeClientset field in Controller struct
// if it gets kubeClient from cfg else it returns error
func (c *Controller) setKubeClient(cfg *rest.Config) error {
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}
	c.KubeClientset = kubeClient
	return nil
}

/*
// setClientset set Clientset field in Controller struct
// if it gets Clientiset from cfg else it returns error
func (c *Controller) setClientSet(cfg *rest.Config) error {
	crdClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return err
	}
	c.Clientset = crdClient
	return nil
}*/

// setNodeName set HostName field in Controller struct
// if it gets from env else it returns error
func (c *Controller) setNodeName() error {
	host, ret := os.LookupEnv("NODE_NAME")
	if ret != true {
		return errors.New("error building hostname")
	}
	c.HostName = host
	return nil
}

// WaitForDiskCRD will block till the CRDs are loaded
// into Kubernetes
func (c *Controller) WaitForDiskCRD() {
	for {
		_, err := c.ListDiskResource()
		if err != nil {
			glog.Errorf("Disk CRD is not available yet. Retrying after %v, error: %v", CRDRetryInterval, err)
			time.Sleep(CRDRetryInterval)
			continue
		}
		glog.Info("Disk CRD is available")
		break
	}
}

// WaitForBlockDeviceCRD will block till the CRDs are loaded
// into Kubernetes
func (c *Controller) WaitForBlockDeviceCRD() {
	for {
		_, err := c.ListBlockDeviceResource()
		if err != nil {
			glog.Errorf("BlockDevice CRD is not available yet. Retrying after %v, error: %v", CRDRetryInterval, err)
			time.Sleep(CRDRetryInterval)
			continue
		}
		glog.Info("BlockDevice CRD is available")
		break
	}
}

// Start is called when we execute cli command ndm start.
func (c *Controller) Start() {
	c.InitializeSparseFiles()
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	if err := c.run(2, stopCh); err != nil {
		glog.Fatalf("error running controller: %s", err.Error())
	}
}

// Broadcast Broadcasts controller pointer. We are using one single pointer of controller
// in our application. In that controller pointer each probe and filter registers themselves
// and later we can list no of active probe using controller object.
func (c *Controller) Broadcast() {
	// sending controller object to each probe. Each probe can get a copy of
	// controller struct anytime only they need to read channel.
	go func() {
		for {
			ControllerBroadcastChannel <- c
		}
	}()
}

// run waits until it gets any interrupt signals
func (c *Controller) run(threadiness int, stopCh <-chan struct{}) error {
	glog.Info("started the controller")
	<-stopCh
	glog.Info("changing the state to unknown before shutting down.")
	// Changing the state to unknown before shutting down. Similar as when one pod is
	// running and you stopped kubelet it will make pod status unknown.
	c.MarkDiskStatusToUnknown()
	//TODO: To be called when Device CR will be implemented
	//c.MarkBlockDeviceStatusToUnknown()
	glog.Info("shutting down the controller")
	return nil
}

// Lock takes a lock on Controller struct
func (c *Controller) Lock() {
	c.Mutex.Lock()
}

// Unlock unlocks the lock on Controller struct
func (c *Controller) Unlock() {
	c.Mutex.Unlock()
}
