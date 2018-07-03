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
	"fmt"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsMiddleware is middleware function to update metrics
func MetricsMiddleware() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.Uptime.Set(time.Now().Sub(metrics.StartingTime).Seconds())
		nc, err := metrics.NewNodeCollector()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Couldn't create %s", err)))
			return
		}

		registry := prometheus.NewRegistry()
		err = registry.Register(nc)
		if err != nil {
			glog.Errorln("Couldn't register collector:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Couldn't register collector: %s", err)))
			return
		}

		gatherers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
			registry,
		}
		h := promhttp.InstrumentMetricHandler(
			registry,
			promhttp.HandlerFor(gatherers,
				promhttp.HandlerOpts{
					ErrorHandling: promhttp.ContinueOnError,
				}),
		)
		h.ServeHTTP(w, r)
	})
}
