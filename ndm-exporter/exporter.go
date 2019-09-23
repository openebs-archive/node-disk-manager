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

package ndm_exporter

import (
	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/db/kubernetes"
	"github.com/openebs/node-disk-manager/ndm-exporter/collector"
	"github.com/openebs/node-disk-manager/pkg/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Exporter contains the options for starting the exporter along with
// clients to retrieve the metrics data
type Exporter struct {
	Client kubernetes.Client
	Mode   string
	Server server.Server
}

const (
	// ClusterLevel is the cluster mode operation of the exporter
	ClusterLevel = "cluster"
	// NodeLevel is the node level mode operation of the exporter
	NodeLevel = "node"
	// Port is the default port on which to start http server
	Port = ":9100"
	// MetricsPath is the endpoint at which metrics will be available
	MetricsPath = "/metrics"
)

// Run starts the exporter, depending on the mode of startup of the exporter
func (e *Exporter) Run() error {
	var err error

	// get the kube config
	cfg, err := config.GetConfig()
	if err != nil {
		glog.Errorf("error getting config. %v", err)
		return err
	}

	glog.V(2).Info("Client config created.")

	// generate a new client object
	e.Client, err = kubernetes.New(cfg)
	if err != nil {
		glog.Errorf("error creating client from config. %v", err)
		return err
	}

	glog.V(2).Info("K8s Client generated using the config.")

	// register the scheme for the APIs
	if err = e.Client.RegisterAPI(); err != nil {
		glog.Errorf("error registering scheme. %v", err)
		return err
	}

	glog.V(2).Info("APIs registered.")

	switch e.Mode {
	case ClusterLevel:
		err = e.runClusterExporter()
	case NodeLevel:
		err = e.runNodeExporter()
	}

	if err != nil {
		glog.Error("error in running exporter")
		return err
	}

	// set handler for server to prometheus handler
	e.Server.Handler = promhttp.Handler()

	// start the server
	if err = e.Server.Start(); err != nil {
		glog.Error("error in running exporter")
		return err
	}
	return nil
}

// runClusterExporter starts the cluster level ndm exporter
func (e *Exporter) runClusterExporter() error {
	glog.Info("Starting cluster level exporter . . .")

	// create instance of a new static collector and register it.
	c := collector.NewStaticMetricCollector(e.Client)
	prometheus.MustRegister(c)

	return nil
}

// runNodeExporter starts the node level ndm exporter
func (e *Exporter) runNodeExporter() error {
	glog.Info("Starting node level exporter . . .")
	// TODO code for node exporter
	return nil
}
