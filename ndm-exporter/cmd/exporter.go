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

package cmd

import (
	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/ndm-exporter/collector/static"
	"github.com/openebs/node-disk-manager/pkg/apis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Exporter struct {
	Client client.Client
	ExporterOptions
}

const (
	ClusterLevel = "cluster"
	NodeLevel    = "node"
	Port         = ":8080"
	MetricsPath  = "/metrics"
)

type ExporterOptions struct {
	// Mode to run the exporter. Can be either cluster or node
	Mode string

	// Port to listen
	Port string

	// MetricsPath to query for the metrics
	MetricsPath string
}

func (e *Exporter) Run() error {
	var err error
	switch e.Mode {
	case ClusterLevel:
		err = runClusterExporter()
	case NodeLevel:
		err = runNodeExporter()
	}

	if err != nil {
		glog.Error("error in running exporter")
		return err
	}

	// starting prometheus http handler
	http.Handle(e.MetricsPath, promhttp.Handler())
	err = http.ListenAndServe(e.Port, nil)
	if err != nil {
		glog.Errorf("error in starting http server. %v", err)
		return err
	}
	glog.Info("Server started on %s at %s endpoint", e.Port, e.MetricsPath)
	return nil
}

func runClusterExporter() error {
	glog.Info("Starting cluster level exporter . . .")

	// get the kube config
	cfg, err := config.GetConfig()
	if err != nil {
		glog.Errorf("error getting config", err)
		return err
	}

	// generate the client
	k8sClient, err := client.New(cfg, client.Options{})
	if err != nil {
		glog.Errorf("error creating k8s client", err)
		return err
	}

	// generate a manager. This is required for registering the APIs
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		glog.Errorf("error creating manager", err)
		return err
	}

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		glog.Errorf("error registering apis", err)
		return err
	}

	// create instance of a new static collector and register it.
	c := static.New(k8sClient)
	prometheus.MustRegister(c)

	return nil
}

func runNodeExporter() error {
	glog.Info("Starting node level exporter . . .")
	// TODO code for node exporter
	return nil
}
