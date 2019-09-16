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
