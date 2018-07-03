/*
Copyright 2018 The OpenEBS Author

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

package server

import (
	"net/http"
	"sort"

	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// ListenPort is the port in which HTTP server is running
	ListenPort = ":9090"
	// MetricsPath is the endpoint path of metrics
	MetricsPath = "/metrics"
)

func init() {
	prometheus.MustRegister(metrics.Uptime)
}

// StartHTTPServer is function to start HTTP server.
func StartHTTPServer() error {
	http.Handle(MetricsPath, MetricsMiddleware())
	glog.Info("Starting HTTP server at http://localhost" + ListenPort + MetricsPath + " for metrics.")
	nc, err := metrics.NewNodeCollector()
	if err != nil {
		return err
	}
	glog.Infof("enabled collectors :")
	collectors := []string{}
	for n := range nc.Collectors {
		collectors = append(collectors, n)
	}
	sort.Strings(collectors)
	for _, n := range collectors {
		glog.Infof(" - %s", n)
	}
	err = http.ListenAndServe(ListenPort, nil)
	if err != nil {
		return err
	}
	return nil
}
