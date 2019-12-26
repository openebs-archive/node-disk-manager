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

package smart

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// MetricsData is the prometheus metrics that are exposed by the exporter. This includes
// all the metrics that are available via SMART
// TODO additional smart metrics need to be added here
type MetricsData struct {
	// blockDeviceCurrentTemperatureValid tells whether the current temperature data is valid
	blockDeviceCurrentTemperatureValid *prometheus.GaugeVec
	// blockDeviceTemperature is the temperature of the the blockdevice if it is reported
	blockDeviceCurrentTemperature *prometheus.GaugeVec

	// errors and rejected requests
	rejectRequestCount prometheus.Counter
	errorRequestCount  prometheus.Counter
}

//MetricsLabels are the labels that are available on the prometheus metrics
type MetricsLabels struct {
	UUID     string
	Path     string
	HostName string
	NodeName string
}

// Metrics defines the metrics data along with the labels present on those metrics.
// The collector(currently seachest/smart) used to fetch the metrics is also defined
type Metrics struct {
	CollectorType string
	MetricsData
	MetricsLabels
}

// NewMetrics creates a new Metrics with the given collector type
func NewMetrics(collector string) *Metrics {
	return &Metrics{
		CollectorType: collector,
	}
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

func (m *Metrics) WithBlockDeviceCurrentTemperature() *Metrics {
	m.blockDeviceCurrentTemperature = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.CollectorType,
			Name:      "block_device_current_temperature_celsius",
			Help:      `Current reported temperature of the blockdevice. -1 if not reported`,
		},
		[]string{"blockdevicename", "path", "hostname", "nodename"},
	)
	return m
}

func (m *Metrics) WithBlockDeviceCurrentTemperatureValid() *Metrics {
	m.blockDeviceCurrentTemperatureValid = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.CollectorType,
			Name:      "block_device_current_temperature_valid",
			Help:      `Validity of the current temperature data reported. 0 means not valid, 1 means valid`,
		},
		[]string{"blockdevicename", "path", "hostname", "nodename"},
	)
	return m
}

func (m *Metrics) WithRejectRequest() *Metrics {
	m.rejectRequestCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: m.CollectorType,
			Name:      "reject_request_count",
			Help:      `No. of requests rejected by the exporter`,
		},
	)
	return m
}

func (m *Metrics) WithErrorRequest() *Metrics {
	m.errorRequestCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: m.CollectorType,
			Name:      "error_request_count",
			Help:      `No. of requests errored out by the exporter`,
		})
	return m
}

func (ml *MetricsLabels) WithBlockDeviceUUID(uuid string) *MetricsLabels {
	ml.UUID = uuid
	return ml
}

func (ml *MetricsLabels) WithBlockDevicePath(path string) *MetricsLabels {
	// remove /dev from the device path so that the device path is similar to the
	// path given by node exporter
	ml.Path = strings.ReplaceAll(path, "/dev/", "")
	return ml
}

func (ml *MetricsLabels) WithBlockDeviceHostName(hostName string) *MetricsLabels {
	ml.HostName = hostName
	return ml
}

func (ml *MetricsLabels) WithBlockDeviceNodeName(nodeName string) *MetricsLabels {
	ml.NodeName = nodeName
	return ml
}

func (m *Metrics) SetBlockDeviceCurrentTemperature(currentTemp int16) *Metrics {
	m.blockDeviceCurrentTemperature.WithLabelValues(m.UUID,
		m.Path,
		m.HostName,
		m.NodeName).
		Set(float64(currentTemp))
	return m
}

func (m *Metrics) SetBlockDeviceCurrentTemperatureValid(valid bool) *Metrics {
	m.blockDeviceCurrentTemperatureValid.WithLabelValues(m.UUID,
		m.Path,
		m.HostName,
		m.NodeName).
		Set(getTemperatureValidity(valid))
	return m
}

// getTemperatureValidity converts temperature validity
// flag to a metric
func getTemperatureValidity(isValid bool) float64 {
	if isValid {
		return 1
	}
	return 0
}
