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
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"

	clientset "github.com/openebs/node-disk-manager/pkg/client/clientset/versioned"
	"github.com/openebs/node-disk-manager/pkg/signals"
)

// Watch sets up the controller, which in-turn
// scans the locally attached disks and push them to etcd
func Watch(kuberconfig string) {
	masterURL := ""
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorf("failed to get k8s Incluster config. %+v", err)
		if kuberconfig == "" {
			glog.Fatalf("kubeconfig is empty")
		} else {
			cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kuberconfig)
			if err != nil {
				glog.Fatalf("Error building kubeconfig: %s", err.Error())
			}
		}
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubeclient: %s", err.Error())
	}

	crdClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building crd client: %s", err.Error())
	}

	host, ret := os.LookupEnv("NODE_NAME")
	if ret != true {
		glog.Fatalf("Error building hostname")
	}

	controller := NewController(host, kubeClient, crdClient)

	if err = controller.Run(2, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}

// DeviceList sets up the controller and calls DevList, which
// queries all the disk resources for this host from etcd and
// prints them to the standard output
func DeviceList(kuberconfig string) {
	masterURL := ""

	cfg, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorf("failed to get k8s Incluster config. %+v", err)
		if kuberconfig == "" {
			glog.Fatalf("kubeconfig is empty")
		} else {
			cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kuberconfig)
			if err != nil {
				glog.Fatalf("Error building kubeconfig: %s", err.Error())
			}
		}
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubeclient: %s", err.Error())
	}

	crdClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building crd client: %s", err.Error())
	}

	host, ret := os.LookupEnv("NODE_NAME")
	if ret != true {
		glog.Fatalf("Error building hostname")
	}

	controller := NewController(host, kubeClient, crdClient)

	if err = controller.DevList(); err != nil {
		glog.Fatalf("Error listing device: %s", err.Error())
	}
}
