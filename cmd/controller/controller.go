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

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
	clientset "github.com/openebs/node-disk-manager/pkg/client/clientset/versioned"
	"github.com/openebs/node-disk-manager/pkg/signals"
	"github.com/openebs/node-disk-manager/pkg/util"
)

const (
	NDMKind     = "Disk"                   // NDMKind is the CR kind.
	NDMVersion  = "openebs.io/v1alpha1"    // NDMVersion is the CR version.
	NDMHostKey  = "kubernetes.io/hostname" // NDMHostKey is host name label prefix.
	NDMActive   = "Active"                 // NDMActive is constant for active resource status.
	NDMInactive = "Inactive"               // NDMInactive is constant for inactive resource status.
)

// ControllerBroadcastChannel is used to send a copy of controller object to each probe.
// Each probe can get the copy of controller struct any time they need to read the channel.
var ControllerBroadcastChannel = make(chan *Controller)

// Controller is the controller implementation for do resources
type Controller struct {
	HostName      string               // HostName is host name in which disk is attached
	KubeClientset kubernetes.Interface // KubeClientset is standard kubernetes clientset
	Clientset     clientset.Interface  // Clientset is clientset for our own API group
	Probes        []*Probe             // Probes are the registered probes like udev/smart
	Mutex         *sync.Mutex          // Mutex is used to lock and unlock Controller
}

// NewController returns a controller pointer for any error case it will return nil
func NewController(kubeconfig string) (*Controller, error) {
	cfg, err := getCfg(kubeconfig)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	crdClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	host, ret := os.LookupEnv("NODE_NAME")
	if ret != true {
		return nil, errors.New("error building hostname")
	}
	mutex := &sync.Mutex{}
	probes := make([]*Probe, 0)
	controller := &Controller{
		HostName:      host,
		KubeClientset: kubeClient,
		Clientset:     crdClient,
		Probes:        probes,
		Mutex:         mutex,
	}
	return controller, nil
}

// Start scans the locally attached disks and push them to etcd.
// This function is called when we execute cli command ndm start.
func Start(kubeconfig string) {
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	controller, err := NewController(kubeconfig)
	if err != nil {
		glog.Fatal(err)
	}
	if err = controller.run(2, stopCh); err != nil {
		glog.Fatalf("error running controller: %s", err.Error())
	}
}

// run broadcasts controller copy to each probe. We are using one single copy of controller
// in our application in that controller object each probe registeres themselves and later
// we can list no of active probe using controller object for that run broadcasts controller
// copy to each probe.
func (c *Controller) run(threadiness int, stopCh <-chan struct{}) error {
	glog.Info("started the controller")
	// sending controller object to each probe. Each probe can get a copy of
	// controller struct anytime only they need to read channel.
	go func() {
		for {
			ControllerBroadcastChannel <- c
		}
	}()
	<-stopCh
	glog.Info("shutting down the controller")
	return nil
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

// DeactivateStaleDiskResource deactivates the stale entry from etcd.
// It gets list of resources which are present in system and queries etcd to get
// list of active resources. One active resource which is present in etcd not in
// system that will be marked as inactive.
func (c *Controller) DeactivateStaleDiskResource(devices []string) {
	listDR, err := c.ListDiskResource()
	if err != nil {
		glog.Error(err)
		return
	}
	for _, item := range listDR.Items {
		if !util.Contains(devices, item.ObjectMeta.Name) {
			c.DeactivateDisk(item)
		}
	}
}

// Lock takes a lock on Controller struct
func (c *Controller) Lock() {
	c.Mutex.Lock()
}

// Unlock unlocks the lock on Controller struct
func (c *Controller) Unlock() {
	c.Mutex.Unlock()
}
