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

	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ListenPort  string = ":9090"
	MetricsPath string = "/metrics"
)

func init(){
	prometheus.MustRegister(metrics.Uptime)
}

func StartHttpServer() error {
	http.Handle(MetricsPath, MetricsMiddleware(promhttp.Handler()))
	glog.Info("Starting HTTP server at http://localhost" + ListenPort + MetricsPath + " for metrics.")
	err := http.ListenAndServe(ListenPort, nil)
	if err != nil {
		glog.Error(err)
		return err
	}
	return nil
}
