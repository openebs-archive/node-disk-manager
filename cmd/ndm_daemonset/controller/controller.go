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
	"os"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/pkg/util"
)

const (
	// FalseString contains string value of false
	FalseString = "false"
	// TrueString contains string value of true
	TrueString = "true"
	// NDMBlockDeviceKind is the Device kind CR.
	NDMBlockDeviceKind = "BlockDevice"
	// kubernetesLabelPrefix is the prefix for k8s labels
	kubernetesLabelPrefix = "kubernetes.io/"
	// openEBSLabelPrefix is the label prefix for openebs labels
	openEBSLabelPrefix = "openebs.io/"
	// HostNameKey is the key for hostname
	HostNameKey = "hostname"
	// LabelTypeNode is the key present in an ENV variable(containing comma separated string value)
	LabelTypeNode = "node-attributes"
	// NodeNameKey is the node name label prefix
	NodeNameKey = "nodename"
	// KubernetesHostNameLabel is the hostname label used by k8s
	KubernetesHostNameLabel = kubernetesLabelPrefix + HostNameKey
	// NDMVersion is the CR version.
	NDMVersion = openEBSLabelPrefix + "v1alpha1"
	// reconcileKey is the key used for enable/disable of reconciliation
	reconcileKey = "reconcile"
	// OpenEBSReconcile is used in annotation to check whether CR is to be reconciled or not
	OpenEBSReconcile = openEBSLabelPrefix + reconcileKey
	// NDMNotPartitioned is used to say blockdevice does not have any partition.
	NDMNotPartitioned = "No"
	// NDMPartitioned is used to say blockdevice has some partitions.
	NDMPartitioned = "Yes"
	// NDMActive is constant for active resource status.
	NDMActive = "Active"
	// NDMInactive is constant for inactive resource status.
	NDMInactive = "Inactive"
	// NDMUnknown is constant for resource unknown status.
	NDMUnknown = "Unknown"
	// NDMDeviceTypeKey specifies the block device type
	NDMDeviceTypeKey = NDMLabelPrefix + "blockdevice-type"
	// NDMManagedKey specifies blockdevice cr should be managed by ndm or not.
	NDMManagedKey = NDMLabelPrefix + "managed"
	// nodeLabelsKey is the meta config key name for adding node labels
	nodeLabelsKey = "node-labels"
	// deviceLabelsKey is the meta config key name for adding device labels
	deviceLabelsKey = "device-labels"
	// NDMLabelPrefix is the label prefix for ndm labels
	NDMLabelPrefix = "ndm.io/"
	// NDMZpoolName specifies the zpool name
	NDMZpoolName = NDMLabelPrefix + "zpool-name"
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

// NDMOptions defines the options to run the NDM daemon
type NDMOptions struct {
	ConfigFilePath string
	// holds the slice of feature gates.
	FeatureGate []string
}

// Controller is the controller implementation for disk resources
type Controller struct {
	config *rest.Config // config is the generated config using kubeconfig/incluster config
	// Namespace is the namespace in which NDM is installed
	Namespace string
	// Clientset is the client used to interface with API server
	Clientset client.Client
	NDMConfig *NodeDiskManagerConfig // NDMConfig contains custom config for ndm
	Mutex     *sync.Mutex            // Mutex is used to lock and unlock Controller
	Filters   []*Filter              // Filters are the registered filters like os disk filter
	Probes    []*Probe               // Probes are the registered probes like udev/smart
	// NodeAttribute is a map of various attributes of the node in which this daemon is running.
	// The attributes can be hostname, nodename, zone, failure-domain etc
	NodeAttributes map[string]string
	// BDHierarchy stores the hierarchy of devices on this node
	BDHierarchy blockdevice.Hierarchy
}

// NewController returns a controller pointer for any error case it will return nil
func NewController() (*Controller, error) {
	controller := &Controller{}
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	controller.config = cfg

	// get the namespace in which NDM is installed
	ns, err := getNamespace()
	if err != nil {
		return controller, err
	}
	controller.Namespace = ns

	mgr, err := manager.New(controller.config, manager.Options{Namespace: controller.Namespace, MetricsBindAddress: "0"})
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

	controller.WaitForBlockDeviceCRD()
	return controller, nil
}

// SetControllerOptions sets the various attributes and options
// on the controller
func (c *Controller) SetControllerOptions(opts NDMOptions) error {
	// set the config for running NDM daemon
	c.SetNDMConfig(opts)

	c.Filters = make([]*Filter, 0)
	c.Probes = make([]*Probe, 0)
	c.NodeAttributes = make(map[string]string, 0)
	c.Mutex = &sync.Mutex{}
	if err := c.setNodeAttributes(); err != nil {
		return err
	}
	return nil
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

	// set the node labels
	if err = c.setNodeLabels(); err != nil {
		return fmt.Errorf("unable to set node attributes:%v", err)
	}
	return nil
}

// setNodeLabels set NodeAttribute field in Controller struct
// from the labels in node object
func (c *Controller) setNodeLabels() error {
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
	if hostName, ok := node.Labels[KubernetesHostNameLabel]; !ok || hostName == "" {
		c.NodeAttributes[HostNameKey] = nodeName
	} else {
		c.NodeAttributes[HostNameKey] = hostName
	}

	var labelPattern []string

	// Get the list of node label patterns to be added from the configmap
	if c.NDMConfig != nil {
		for _, metaConfig := range c.NDMConfig.MetaConfigs {
			if metaConfig.Key == nodeLabelsKey {
				labelPattern = strings.Split(metaConfig.Pattern, ",")
			}
		}
	}

	if len(labelPattern) > 0 {
		// Add only those node labels that matches the pattern specified in the
		// node-labels meta config
		for key, value := range node.Labels {
			for _, pattern := range labelPattern {
				if util.IsMatchRegex(pattern, key) {
					if value != "" {
						c.NodeAttributes[key] = value
						break
					}
				}
			}
		}
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

// WaitForBlockDeviceCRD will block till the CRDs are loaded
// into Kubernetes
func (c *Controller) WaitForBlockDeviceCRD() {
	for {
		_, err := c.ListBlockDeviceResource(false)
		if err != nil {
			klog.Errorf("BlockDevice CRD is not available yet. Retrying after %v, error: %v", CRDRetryInterval, err)
			time.Sleep(CRDRetryInterval)
			_, err := c.newClientSet()
			if err != nil {
				klog.Errorf("unable to set clientset field in controller struct, Error: %v", err)
			}
			continue
		}
		klog.Info("BlockDevice CRD is available")
		break
	}
}

// Start is called when we execute cli command ndm start.
func (c *Controller) Start() {
	c.InitializeSparseFiles()
	// set up signals so we handle the first shutdown signal gracefully
	ctx := signals.SetupSignalHandler()
	if err := c.run(2, ctx); err != nil {
		klog.Fatalf("error running controller: %s", err.Error())
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
func (c *Controller) run(threadiness int, ctx context.Context) error {
	klog.Info("started the controller")

	if ctx.Err() != nil {
		return ctx.Err()
	}
	<-ctx.Done()
	// Changing the state to unknown before shutting down. Similar as when one pod is
	// running and you stopped kubelet it will make pod status unknown.
	c.MarkBlockDeviceStatusToUnknown()
	klog.Info("shutting down the controller")
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
