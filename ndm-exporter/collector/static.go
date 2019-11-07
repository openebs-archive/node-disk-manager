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

package collector

import (
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/db/kubernetes"
	"github.com/openebs/node-disk-manager/pkg/metrics/static"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog"
	"sync"
)

// StaticMetricCollector contains the metrics, concurrency handler and client to get the
// static metrics
type StaticMetricCollector struct {
	// Client is the k8s client which will be used to interface with etcd
	Client kubernetes.Client

	// concurrency handling
	sync.Mutex
	requestInProgress bool

	// all the exposed metrics
	metrics *static.Metrics
}

// NewStaticMetricCollector creates a new instance of StaticMetricCollector which
// implements Collector interface
func NewStaticMetricCollector(c kubernetes.Client) prometheus.Collector {
	return &StaticMetricCollector{
		Client:  c,
		metrics: static.NewMetrics(),
	}
}

// setRequestProgressToFalse is used to set the progress flag, when a request is
// processed or errored
func (mc *StaticMetricCollector) setRequestProgressToFalse() {
	mc.Lock()
	mc.requestInProgress = false
	mc.Unlock()
}

// Describe is the implementation of Describe in prometheus.Collector
func (mc *StaticMetricCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, col := range mc.metrics.Collectors() {
		col.Describe(ch)
	}
}

// Collect is the implementation of Collect in prometheus.Collector
func (mc *StaticMetricCollector) Collect(ch chan<- prometheus.Metric) {

	klog.V(4).Info("Starting to collect metrics for a request")

	// when a second request comes, and the first one is already in progress,
	// ignore/reject the second request
	mc.Lock()
	if mc.requestInProgress {
		klog.V(4).Info("Another request already in progress.")
		mc.metrics.IncRejectRequestCounter()
		mc.Unlock()
		return
	}

	mc.requestInProgress = true
	mc.Unlock()

	// once a request is processed, set the progress flag to false
	defer mc.setRequestProgressToFalse()

	klog.V(4).Info("Setting client for this request.")

	// set the client each time
	if err := mc.Client.InitClient(); err != nil {
		klog.Errorf("error setting client. %v", err)
		mc.metrics.IncErrorRequestCounter()
		mc.collectErrors(ch)
		return
	}

	// get list of blockdevices from etcd
	blockDevices, err := mc.Client.ListBlockDevice()
	if err != nil {
		mc.metrics.IncErrorRequestCounter()
		mc.collectErrors(ch)
		return
	}

	klog.V(4).Info("Metric data fetched from etcd")

	// set the metric data into the respective fields
	mc.metrics.SetMetrics(blockDevices)

	klog.V(4).Info("Prometheus metrics is set and initializing collection.")

	// collect each metric
	for _, col := range mc.metrics.Collectors() {
		col.Collect(ch)
	}
}

// collectErrors collects only the error metrics and set it on the channel
func (mc *StaticMetricCollector) collectErrors(ch chan<- prometheus.Metric) {
	for _, col := range mc.metrics.ErrorCollectors() {
		col.Collect(ch)
	}
}
