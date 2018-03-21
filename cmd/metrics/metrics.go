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
package metrics

import (
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ListenPort   string
	MetricsPath  string
	startingTime time.Time
	// Variable for generate uptime metrics
	Uptime       = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "Uptime",
		Help: "Uptime of node disk manager",
	})
)

func init() {
	ListenPort = ":9090"
	MetricsPath = "/metrics"
	startingTime = time.Now()
	prometheus.MustRegister(Uptime)
	Uptime.Set(time.Now().Sub(startingTime).Seconds())
}

func StartHttpServer() error {
	glog.Info("Started HTTP server for /metrics endpont")
	http.Handle(MetricsPath, metricsMiddleware(promhttp.Handler()))
	err := http.ListenAndServe(ListenPort, nil)
	if err != nil {
		return err
	}
	return nil
}

func metricsMiddleware(promHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Uptime.Set(time.Now().Sub(startingTime).Seconds())
		promHandler.ServeHTTP(w, r)
	})
}
