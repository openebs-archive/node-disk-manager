package static

import (
	"github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	NodeNamespace = "node"
)

type metrics struct {
	blockDeviceState *prometheus.GaugeVec

	// errors and rejected requests
	rejectRequestCount prometheus.Counter
}

func newMetrics() *metrics {
	return new(metrics)
}

func (m *metrics) withBlockDeviceState() *metrics {
	m.blockDeviceState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: NodeNamespace,
			Name:      "block_device_state",
			Help:      `State of BlockDevice (0,1,2) = {Active, Inactive, Unknown}`,
		},
		[]string{"Name", "Path", "HostName"},
	)
	return m
}

func (m *metrics) withRejectRequest() *metrics {
	m.rejectRequestCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: NodeNamespace,
			Name:      "reject_request_count",
			Help:      `No. of requests rejected by the exporter`,
		},
	)
	return m
}

func getState(state v1alpha1.BlockDeviceState) float64 {
	switch state {
	case v1alpha1.BlockDeviceActive:
		return 0
	case v1alpha1.BlockDeviceInactive:
		return 1
	case v1alpha1.BlockDeviceUnknown:
		return 2
	}
	return 0
}
