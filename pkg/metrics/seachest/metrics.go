package seachest

import (
	"strings"

	bd "github.com/openebs/node-disk-manager/blockdevice"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	SeachestName = "seachest"
)

type Metrics struct {
	// blockDeviceCurrentTemperatureValid tells whether the current temperature data is valid
	blockDeviceCurrentTemperatureValid *prometheus.GaugeVec
	// blockDeviceTemperature is the temperature of the the blockdevice if it is reported
	blockDeviceCurrentTemperature *prometheus.GaugeVec

	// errors and rejected requests
	rejectRequestCount prometheus.Counter
	errorRequestCount  prometheus.Counter
}

func NewMetrics() *Metrics {
	return new(Metrics).
		withBlockDeviceCurrentTemperatureValid().
		withBlockDeviceCurrentTemperature().
		withRejectRequest().
		withErrorRequest()
}

func (m *Metrics) Collectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.blockDeviceCurrentTemperatureValid,
		m.blockDeviceCurrentTemperature,
		m.rejectRequestCount,
		m.errorRequestCount,
	}
}

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
			Name:      "block_device_current_temperature",
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

func getTemperatureValidity(isValid bool) float64 {
	if isValid {
		return 1
	} else {
		return 0
	}
}
