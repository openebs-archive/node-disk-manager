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

	// blockDevicehighestTemperature is the highest temperature of the the blockdevice if it is reported
	blockDeviceHighestTemperature *prometheus.GaugeVec

	// blockDeviceHighestTemperatureValid tells whether the highest temperature data is valid
	blockDeviceHighestTemperatureValid *prometheus.GaugeVec

	// blockDevicelowestTemperature is the lowest temperature of the the blockdevice if it is reported
	blockDeviceLowestTemperature *prometheus.GaugeVec

	// blockDeviceLowestTemperatureValid tells whether the lowest temperature data is valid
	blockDeviceLowestTemperatureValid *prometheus.GaugeVec

	// blockDeviceCapacity is capacity of block device
	blockDeviceCapacity *prometheus.GaugeVec

	// blockDeviceTotalReadBytes is the total number of bytes read from the block device
	blockDeviceTotalReadBytes *prometheus.CounterVec

	// blockDeviceTotalWrittenBytes is the total number of bytes written from the block device
	blockDeviceTotalWrittenBytes *prometheus.CounterVec

	// blockDeviceUtilizationRate is utilization rate of the block device
	blockDeviceUtilizationRate *prometheus.GaugeVec

	// blockDevicePercentEnduranceUsed  is percentage of endurance used by a block device
	blockDevicePercentEnduranceUsed *prometheus.GaugeVec

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
		m.blockDeviceHighestTemperatureValid,
		m.blockDeviceLowestTemperatureValid,
		m.blockDeviceCurrentTemperature,
		m.blockDeviceHighestTemperature,
		m.blockDeviceLowestTemperature,
		m.blockDeviceTotalReadBytes,
		m.blockDeviceTotalWrittenBytes,
		m.blockDeviceUtilizationRate,
		m.blockDevicePercentEnduranceUsed,
		m.rejectRequestCount,
		m.errorRequestCount,
	}
}

var labels []string = []string{"blockdevicename", "path", "hostname", "nodename"}

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

// WithBlockDeviceCurrentTemperature declares the metric current temperature
// as a prometheus metric
func (m *Metrics) WithBlockDeviceCurrentTemperature() *Metrics {
	m.blockDeviceCurrentTemperature = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.CollectorType,
			Name:      "block_device_current_temperature_celsius",
			Help:      `Current reported temperature of the blockdevice. -1 if not reported`,
		},
		labels,
	)
	return m
}

// WithBlockDeviceHighestTemperature declares the metric highest temperature
// as a prometheus metric
func (m *Metrics) WithBlockDeviceHighestTemperature() *Metrics {
	m.blockDeviceHighestTemperature = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.CollectorType,
			Name:      "block_device_highest_temperature_celsius",
			Help:      `Highest reported temperature of the blockdevice. -1 if not reported`,
		},
		labels,
	)
	return m
}

// WithBlockDeviceLowestTemperature declares the metric lowest temperature
// as a prometheus metric
func (m *Metrics) WithBlockDeviceLowestTemperature() *Metrics {
	m.blockDeviceLowestTemperature = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.CollectorType,
			Name:      "block_device_lowest_temperature_celsius",
			Help:      `Lowest reported temperature of the blockdevice. -1 if not reported`,
		},
		labels,
	)
	return m
}

// WithBlockDeviceCurrentTemperatureValid declares the metric current temperature valid
// as a prometheus metric
func (m *Metrics) WithBlockDeviceCurrentTemperatureValid() *Metrics {
	m.blockDeviceCurrentTemperatureValid = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.CollectorType,
			Name:      "block_device_current_temperature_valid",
			Help:      `Validity of the current temperature data reported. 0 means not valid, 1 means valid`,
		},
		labels,
	)
	return m
}

// WithBlockDeviceHighestTemperatureValid declares the metric highest temperature valid
// as a prometheus metric
func (m *Metrics) WithBlockDeviceHighestTemperatureValid() *Metrics {
	m.blockDeviceHighestTemperatureValid = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.CollectorType,
			Name:      "block_device_highest_temperature_valid",
			Help:      `Validity of the highest temperature data reported. 0 means not valid, 1 means valid`,
		},
		labels,
	)
	return m
}

// WithBlockDeviceLowestTemperatureValid declares the metric lowest temperature valid
// as a prometheus metric
func (m *Metrics) WithBlockDeviceLowestTemperatureValid() *Metrics {
	m.blockDeviceLowestTemperatureValid = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.CollectorType,
			Name:      "block_device_lowest_temperature_valid",
			Help:      `Validity of the lowest temperature data reported. 0 means not valid, 1 means valid`,
		},
		labels,
	)
	return m
}

// WithBlockDeviceCapacity declares the blockdevice capacity
func (m *Metrics) WithBlockDeviceCapacity() *Metrics {
	m.blockDeviceCapacity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.CollectorType,
			Name:      "block_device_capacity_bytes",
			Help:      `Capacity of the block device in bytes`,
		},
		labels,
	)
	return m
}

// WithBlockDeviceTotalBytesRead declares the total number of bytes read by a block device
func (m *Metrics) WithBlockDeviceTotalBytesRead() *Metrics {
	m.blockDeviceTotalReadBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.CollectorType,
			Name:      "block_device_total_read_bytes",
			Help:      `total number of bytes read by a block device in bytes `,
		},
		labels,
	)
	return m
}

// WithBlockDeviceTotalBytesWritten declares the total number of bytes written by a block device
func (m *Metrics) WithBlockDeviceTotalBytesWritten() *Metrics {
	m.blockDeviceTotalWrittenBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.CollectorType,
			Name:      "block_device_total_written_bytes",
			Help:      `total number of bytes written by a block device in bytes `,
		},
		labels,
	)
	return m
}

// WithBlockDeviceUtilizationRate declares the utilization rate of a block device
func (m *Metrics) WithBlockDeviceUtilizationRate() *Metrics {
	m.blockDeviceUtilizationRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.CollectorType,
			Name:      "block_device_utilization_rate_percent",
			Help:      `Ratio of actual workload to manufacturer's designed workload for the device `,
		},
		labels,
	)
	return m
}

// WithBlockDevicePercentEnduranceUsed declares the percentage of endurance used by a block device
func (m *Metrics) WithBlockDevicePercentEnduranceUsed() *Metrics {
	m.blockDevicePercentEnduranceUsed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.CollectorType,
			Name:      "block_device_endurance_used_percent",
			Help:      `Estimate of the percentage of the device life that has been used `,
		},
		labels,
	)
	return m
}

// WithRejectRequest declares the reject request count metric
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

// WithErrorRequest declares the error request count metric
func (m *Metrics) WithErrorRequest() *Metrics {
	m.errorRequestCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: m.CollectorType,
			Name:      "error_request_count",
			Help:      `No. of requests errored out by the exporter`,
		})
	return m
}

// WithBlockDeviceUUID sets the blockdevice UUID to the metric label
func (ml *MetricsLabels) WithBlockDeviceUUID(uuid string) *MetricsLabels {
	ml.UUID = uuid
	return ml
}

// WithBlockDevicePath sets the blockdevice path to the metric label
func (ml *MetricsLabels) WithBlockDevicePath(path string) *MetricsLabels {
	// remove /dev from the device path so that the device path is similar to the
	// path given by node exporter
	ml.Path = strings.ReplaceAll(path, "/dev/", "")
	return ml
}

// WithBlockDeviceHostName sets the blockdevice hostname to the metric label
func (ml *MetricsLabels) WithBlockDeviceHostName(hostName string) *MetricsLabels {
	ml.HostName = hostName
	return ml
}

// WithBlockDeviceNodeName sets the blockdevice nodename to the metric label
func (ml *MetricsLabels) WithBlockDeviceNodeName(nodeName string) *MetricsLabels {
	ml.NodeName = nodeName
	return ml
}

// SetBlockDeviceCurrentTemperature sets the current temperature value to the metric
func (m *Metrics) SetBlockDeviceCurrentTemperature(currentTemp int16) *Metrics {
	m.blockDeviceCurrentTemperature.WithLabelValues(m.UUID,
		m.Path,
		m.HostName,
		m.NodeName,
	).
		Set(float64(currentTemp))
	return m
}

// SetBlockDeviceHighestTemperature sets the highest temperature value to the metric
func (m *Metrics) SetBlockDeviceHighestTemperature(highTemp int16) *Metrics {
	m.blockDeviceHighestTemperature.WithLabelValues(m.UUID,
		m.Path,
		m.HostName,
		m.NodeName,
	).
		Set(float64(highTemp))
	return m
}

// SetBlockDeviceLowestTemperature sets the lowest temperature value to the metric
func (m *Metrics) SetBlockDeviceLowestTemperature(lowTemp int16) *Metrics {
	m.blockDeviceLowestTemperature.WithLabelValues(m.UUID,
		m.Path,
		m.HostName,
		m.NodeName,
	).
		Set(float64(lowTemp))
	return m
}

// SetBlockDeviceCurrentTemperatureValid sets the validity of the exposed current
// temperature metrics
func (m *Metrics) SetBlockDeviceCurrentTemperatureValid(valid bool) *Metrics {
	m.blockDeviceCurrentTemperatureValid.WithLabelValues(m.UUID,
		m.Path,
		m.HostName,
		m.NodeName,
	).
		Set(getTemperatureValidity(valid))
	return m
}

// SetBlockDeviceHighestTemperatureValid sets the validity of the exposed highest
// temperature metrics
func (m *Metrics) SetBlockDeviceHighestTemperatureValid(valid bool) *Metrics {
	m.blockDeviceCurrentTemperatureValid.WithLabelValues(m.UUID,
		m.Path,
		m.HostName,
		m.NodeName,
	).
		Set(getTemperatureValidity(valid))
	return m
}

// SetBlockDeviceLowestTemperatureValid sets the validity of the exposed lowest
// temperature metrics
func (m *Metrics) SetBlockDeviceLowestTemperatureValid(valid bool) *Metrics {
	m.blockDeviceCurrentTemperatureValid.WithLabelValues(m.UUID,
		m.Path,
		m.HostName,
		m.NodeName,
	).
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

// SetBlockDeviceCapacity sets the current block device capacity value to the metric
func (m *Metrics) SetBlockDeviceCapacity(capacity uint64) *Metrics {
	m.blockDeviceCapacity.WithLabelValues(m.UUID,
		m.Path,
		m.HostName,
		m.NodeName,
	).
		Set(float64(capacity))
	return m
}

// SetBlockDeviceTotalBytesRead sets the total bytes read value to the metric
func (m *Metrics) SetBlockDeviceTotalBytesRead(size uint64) *Metrics {
	m.blockDeviceTotalReadBytes.WithLabelValues(m.UUID,
		m.Path,
		m.HostName,
		m.NodeName,
	)
	return m
}

// SetBlockDeviceTotalBytesWritten sets the total bytes written value to the metric
func (m *Metrics) SetBlockDeviceTotalBytesWritten(size uint64) *Metrics {
	m.blockDeviceTotalWrittenBytes.WithLabelValues(m.UUID,
		m.Path,
		m.HostName,
		m.NodeName,
	)
	return m
}

// SetBlockDeviceUtilizationRate sets the utilization rate value to the metric
func (m *Metrics) SetBlockDeviceUtilizationRate(size float64) *Metrics {
	m.blockDeviceUtilizationRate.WithLabelValues(m.UUID,
		m.Path,
		m.HostName,
		m.NodeName,
	).
		Set(float64(size))
	return m
}

// SetBlockDevicePercentEnduranceUsed sets the percentage of endurance used by a block device to the metric
func (m *Metrics) SetBlockDevicePercentEnduranceUsed(size float64) *Metrics {
	m.blockDevicePercentEnduranceUsed.WithLabelValues(m.UUID,
		m.Path,
		m.HostName,
		m.NodeName,
	).
		Set(float64(size))
	return m
}
