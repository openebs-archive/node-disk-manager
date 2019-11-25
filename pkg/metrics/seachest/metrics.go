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

package seachest

import (
	"strings"

	bd "github.com/openebs/node-disk-manager/blockdevice"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// SeachestName is the name prefix used for all seachest metrics
	SeachestName = "seachest"
)

// Metrics is the prometheus metrics that are exposed by the exporter. This includes
// all the metrics that can be fetched using seachest library
type Metrics struct {
	// blockDeviceCurrentTemperatureValid tells whether the current temperature data is valid
	blockDeviceCurrentTemperatureValid *prometheus.GaugeVec
	// blockDeviceTemperature is the temperature of the the blockdevice if it is reported
	blockDeviceCurrentTemperature *prometheus.GaugeVec

	// errors and rejected requests
	rejectRequestCount prometheus.Counter
	errorRequestCount  prometheus.Counter
}

// NewMetrics creates instance of metrics
func NewMetrics() *Metrics {
	return new(Metrics).
		withBlockDeviceCurrentTemperatureValid().
		withBlockDeviceCurrentTemperature().
		withRejectRequest().
		withErrorRequest()
}

// Collectors lists out all the collectors for which the metrics is exposed
func (m *Metrics) Collectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.blockDeviceCurrentTemperatureValid,
		m.blockDeviceCurrentTemperature,
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

func (m *Metrics) withBlockDeviceCurrentTemperature() *Metrics {
	m.blockDeviceCurrentTemperature = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: SeachestName,
			Name:      "block_device_current_temperature_celsius",
			Help:      `Current reported temperature of the blockdevice. -1 if not reported`,
		},
		[]string{"blockdevicename", "path", "hostname", "nodename"},
	)
	return m
}

func (m *Metrics) withBlockDeviceCurrentTemperatureValid() *Metrics {
	m.blockDeviceCurrentTemperatureValid = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: SeachestName,
			Name:      "block_device_current_temperature_valid",
			Help:      `Validity of the current temperature data reported. 0 means not valid, 1 means valid`,
		},
		[]string{"blockdevicename", "path", "hostname", "nodename"},
	)
	return m
}

func (m *Metrics) withRejectRequest() *Metrics {
	m.rejectRequestCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: SeachestName,
			Name:      "reject_request_count",
			Help:      `No. of requests rejected by the exporter`,
		},
	)
	return m
}

func (m *Metrics) withErrorRequest() *Metrics {
	m.errorRequestCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: SeachestName,
			Name:      "error_request_count",
			Help:      `No. of requests errored out by the exporter`,
		})
	return m
}

// SetMetrics is used to set the seachest metrics onto resepective fields
// of prometheus metrics
func (m *Metrics) SetMetrics(blockDevices []bd.BlockDevice) {
	for _, blockDevice := range blockDevices {
		// remove /dev from the device path so that the device path is similar to the
		// path given by node exporter
		path := strings.ReplaceAll(blockDevice.Path, "/dev/", "")
		m.blockDeviceCurrentTemperature.WithLabelValues(blockDevice.UUID,
			path,
			blockDevice.NodeAttributes[bd.HostName],
			blockDevice.NodeAttributes[bd.NodeName]).
			Set(float64(blockDevice.TemperatureInfo.CurrentTemperature))
		m.blockDeviceCurrentTemperatureValid.WithLabelValues(blockDevice.UUID,
			path,
			blockDevice.NodeAttributes[bd.HostName],
			blockDevice.NodeAttributes[bd.NodeName]).
			Set(getTemperatureValidity(blockDevice.TemperatureInfo.TemperatureDataValid))
	}
}

// getTemperatureValidity converts temperature validity
// flag to a metric
func getTemperatureValidity(isValid bool) float64 {
	if isValid {
		return 1
	}
	return 0
}
