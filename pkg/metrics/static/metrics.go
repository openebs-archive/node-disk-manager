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

package static

import (
	"strings"

	bd "github.com/openebs/node-disk-manager/blockdevice"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// NodeNamespace is the namespace used for components on the node.
	// This has been seen as a practice in node exporter.
	NodeNamespace = "node"
)

// Metrics is the prometheus metrics that are exposed by the exporter
type Metrics struct {
	blockDeviceState *prometheus.GaugeVec

	// errors and rejected requests
	rejectRequestCount prometheus.Counter
	errorRequestCount  prometheus.Counter
}

// NewMetrics creates instance of metrics
func NewMetrics() *Metrics {
	return new(Metrics).
		withBlockDeviceState().
		withRejectRequest().
		withErrorRequest()
}

// Collectors lists out all the collectors for which the metrics is exposed
func (m *Metrics) Collectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.blockDeviceState,
		m.rejectRequestCount,
		m.errorRequestCount,
	}
}

// ErrorCollectors lists out all collectors for metrics related to error
func (m *Metrics) ErrorCollectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.rejectRequestCount,
		m.errorRequestCount,
	}
}

// IncRejectRequestCounter increments the reject request error counter
func (m *Metrics) IncRejectRequestCounter() {
	m.rejectRequestCount.Inc()
}

// IncErrorRequestCounter increments the no of requests errored out.
func (m *Metrics) IncErrorRequestCounter() {
	m.errorRequestCount.Inc()
}

func (m *Metrics) withBlockDeviceState() *Metrics {
	m.blockDeviceState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: NodeNamespace,
			Name:      "block_device_state",
			Help:      `State of BlockDevice (0,1,2) = {Active, Inactive, Unknown}`,
		},
		[]string{"blockdevicename", "path", "hostname", "nodename"},
	)
	return m
}

func (m *Metrics) withRejectRequest() *Metrics {
	m.rejectRequestCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: NodeNamespace,
			Name:      "reject_request_count",
			Help:      `No. of requests rejected by the exporter`,
		},
	)
	return m
}

func (m *Metrics) withErrorRequest() *Metrics {
	m.errorRequestCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: NodeNamespace,
			Name:      "error_request_count",
			Help:      `No. of requests errored out by the exporter`,
		})
	return m
}

// SetMetrics is used to set the prometheus metrics to respective fields
func (m *Metrics) SetMetrics(blockDevices []bd.BlockDevice) {
	for _, blockDevice := range blockDevices {
		// do not report metrics for sparse devices
		if blockDevice.DeviceType == bd.SparseBlockDeviceType {
			continue
		}
		// remove /dev from the device path so that the device path is similar to the
		// path given by node exporter
		path := strings.ReplaceAll(blockDevice.Path, "/dev/", "")
		m.blockDeviceState.WithLabelValues(blockDevice.UUID,
			path,
			blockDevice.NodeAttributes[bd.HostName],
			blockDevice.NodeAttributes[bd.NodeName]).
			Set(getState(blockDevice.Status.State))
	}
}

func getState(state string) float64 {
	switch state {
	case bd.Active:
		return 0
	case bd.Inactive:
		return 1
	case bd.Unknown:
		return 2
	}
	// default return unknown state
	return 2
}
