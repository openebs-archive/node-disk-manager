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
	"context"
	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

type MetricCollector struct {
	// Client is the k8s client which will be used to interface with etcd
	Client client.Client

	// concurrency handling
	sync.Mutex
	requestInProgress bool

	// all the exposed metrics
	*metrics
}

func New(client client.Client) prometheus.Collector {
	return &MetricCollector{
		Client: client,
		metrics: newMetrics().
			withBlockDeviceState().
			withRejectRequest(),
	}
}

func (mc *MetricCollector) setRequestProgressToFalse() {
	mc.Lock()
	mc.requestInProgress = false
	mc.Unlock()
}

func (mc *MetricCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, col := range mc.collectors() {
		col.Describe(ch)
	}
}

func (mc *MetricCollector) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		mc.blockDeviceState,
		mc.rejectRequestCount,
	}
}

func (mc *MetricCollector) Collect(ch chan<- prometheus.Metric) {

	// when a second request comes, and the first one is already in progress,
	// ignore/reject the second request
	mc.Lock()
	if mc.requestInProgress {
		mc.rejectRequestCount.Inc()
		mc.Unlock()
		mc.rejectRequestCount.Collect(ch)
		return
	}

	mc.requestInProgress = true
	mc.Unlock()

	// once a request is processed, set the progress flag to false
	defer mc.setRequestProgressToFalse()

	// get required metric data from etcd
	bdList, err := mc.getMetricData()
	if err != nil {
		return
	}

	// set the metric data into the respective fields
	mc.setMetricData(bdList)

	// collect each metric
	for _, col := range mc.collectors() {
		col.Collect(ch)
	}
}

// getMetricData is used to get the metric data from the source
func (mc *MetricCollector) getMetricData() ([]v1alpha1.BlockDevice, error) {
	bdlist := &v1alpha1.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}
	err := mc.Client.List(context.TODO(), nil, bdlist)
	if err != nil {
		glog.Error("error in listing BDs", err)
		return nil, err
	}
	return bdlist.Items, nil
}

// setMetricData sets the metric data into respective fields
func (mc *MetricCollector) setMetricData(bdList []v1alpha1.BlockDevice) {
	for _, bd := range bdList {
		mc.blockDeviceState.WithLabelValues(bd.Name, bd.Spec.Path, bd.Labels[controller.KubernetesHostNameLabel]).Set(getState(bd.Status.State))
	}
}
