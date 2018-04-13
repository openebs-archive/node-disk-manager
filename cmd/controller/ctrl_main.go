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
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	crdClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building sp-spc clientset: %s", err.Error())
	}

	host, err := os.Hostname()
	if err != nil {
		glog.Fatalf("Error building sp-spc clientset: %s", err.Error())
	}

	controller := NewController(host, kubeClient, crdClient)

	if err = controller.Run(2, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}
