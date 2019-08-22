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
	"context"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"os"
	"sync"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
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
	// Kubernetes label prefix
	KubernetesLabelPrefix = "kubernetes.io/"
	// OpenEBS label prefix
	OpenEBSLabelPrefix = "openebs.io/"
	// HostNameKey is the key for hostname
	HostNameKey = "hostname"
	// NodeNameKey is the node name label prefix
	NodeNameKey = "nodename"
	// KubernetesHostNameLabel is the hostname label used by k8s
	KubernetesHostNameLabel = KubernetesLabelPrefix + HostNameKey
	// NDMVersion is the CR version.
	NDMVersion = OpenEBSLabelPrefix + "v1alpha1"
	// NDMNotPartitioned is used to say blockdevice does not have any partition.
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
	NDMDiskTypeKey = "ndm.io/disk-type"
	// NDMDeviceTypeKey specifies the block device type
	NDMDeviceTypeKey = "ndm.io/blockdevice-type"
	// NDMManagedKey specifies disk cr should be managed by ndm or not.
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

// Namespace is the namespace in which NDM is installed
var Namespace string

// ControllerBroadcastChannel is used to send a copy of controller object to each probe.
// Each probe can get the copy of controller struct any time they need to read the channel.
var ControllerBroadcastChannel = make(chan *Controller)

// Controller is the controller implementation for disk resources
type Controller struct {
	config *rest.Config // config is the generated config using kubeconfig/incluster config
	// Clientset is the client used to interface with API server
	Clientset client.Client
	NDMConfig *NodeDiskManagerConfig // NDMConfig contains custom config for ndm
	Mutex     *sync.Mutex            // Mutex is used to lock and unlock Controller
	Filters   []*Filter              // Filters are the registered filters like os disk filter
	Probes    []*Probe               // Probes are the registered probes like udev/smart
	// NodeAttribute is a map of various attributes of the node in which this daemon is running.
	// The attributes can be hostname, nodename, zone, failure-domain etc
	NodeAttributes map[string]string
}

// NewController returns a controller pointer for any error case it will return nil
func NewController(kubeconfig string) (*Controller, error) {
	controller := &Controller{}
	cfg, err := getCfg(kubeconfig)
	if err != nil {
		return nil, err
	}
	controller.config = cfg

	mgr, err := manager.New(controller.config, manager.Options{Namespace: Namespace})
	if err != nil {
		return controller, err
	}

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		return controller, err
	}

	_, err = controller.newClientSet()
	if err != nil {
		return controller, err
	}

	controller.SetNDMConfig()
	controller.Filters = make([]*Filter, 0)
	controller.Probes = make([]*Probe, 0)
	controller.Mutex = &sync.Mutex{}

	// get the namespace in which NDM is installed
	Namespace, err = getNamespace()
	if err != nil {
		return controller, err
	}

	if err := controller.setNodeAttributes(); err != nil {
		return nil, err
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

// newClientSet set Clientset field in Controller struct
// if it gets Client from config. It returns the generated
// client, else it returns error
func (c *Controller) newClientSet() (client.Client, error) {
	clientSet, err := client.New(c.config, client.Options{})
	if err != nil {
		return nil, err
	}
	c.Clientset = clientSet
	return clientSet, nil
}

func (c *Controller) setNodeAttributes() error {
	// sets the node name label
	nodeName, err := getNodeName()
	if err != nil {
		return fmt.Errorf("unable to set node attributes: %v", err)
	}
	c.NodeAttributes[NodeNameKey] = nodeName

	// set the hostname label
	if err = c.setHostName(); err != nil {
		return fmt.Errorf("unable to set node attributes:%v", err)
	}
	return nil
}

// setHostName set NodeAttribute field in Controller struct
// from the labels in node object
func (c *Controller) setHostName() error {
	nodeName := c.NodeAttributes[NodeNameKey]
	// get the node object and fetch the hostname label from the
	// node object
	node := &v1.Node{}
	err := c.Clientset.Get(context.TODO(), client.ObjectKey{Namespace: "", Name: nodeName}, node)
	if err != nil {
		return err
	}

	// if the label is not present, or hostname is an empty string,
	// use nodename as hostname
	if hostName, ok := node.Labels[HostNameKey]; !ok || hostName == "" {
		c.NodeAttributes[HostNameKey] = nodeName
	} else {
		c.NodeAttributes[HostNameKey] = hostName
	}
	return nil
}

// getNodeName gets the node name from env, else
// returns an error
func getNodeName() (string, error) {
	nodeName, ok := os.LookupEnv("NODE_NAME")
	if !ok {
		return "", errors.New("error getting node name")
	}
	return nodeName, nil
}

// getNamespace get Namespace from env, else it returns error
func getNamespace() (string, error) {
	ns, ok := os.LookupEnv("NAMESPACE")
	if !ok {
		return "", errors.New("error getting namespace")
	}
	return ns, nil
}

// WaitForDiskCRD will block till the CRDs are loaded
// into Kubernetes
func (c *Controller) WaitForDiskCRD() {
	for {
		_, err := c.ListDiskResource()
		if err != nil {
			glog.Errorf("Disk CRD is not available yet. Retrying after %v, error: %v", CRDRetryInterval, err)
			time.Sleep(CRDRetryInterval)
			c.newClientSet()
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
			c.newClientSet()
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
	c.MarkBlockDeviceStatusToUnknown()
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
