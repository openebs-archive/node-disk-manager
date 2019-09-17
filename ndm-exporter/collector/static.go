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
	"context"
	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/metrics/static"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

// StaticMetricCollector contains the metrics, concurrency handler and client to get the
// static metrics
type StaticMetricCollector struct {
	// Client is the k8s client which will be used to interface with etcd
	Client client.Client

	// concurrency handling
	sync.Mutex
	requestInProgress bool

	// all the exposed metrics
	metrics *static.Metrics
}

// New creates a new instance of StaticMetricCollector which
// implements Collector interface
func New(client client.Client) prometheus.Collector {
	return &StaticMetricCollector{
		Client:  client,
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

	// when a second request comes, and the first one is already in progress,
	// ignore/reject the second request
	mc.Lock()
	if mc.requestInProgress {
		mc.metrics.IncRejectRequestCounter()
		mc.Unlock()
		for _, col := range mc.metrics.ErrorCollectors() {
			col.Collect(ch)
		}
		return
	}

	mc.requestInProgress = true
	mc.Unlock()

	// once a request is processed, set the progress flag to false
	defer mc.setRequestProgressToFalse()

	// get required metric data from etcd
	blockDevices, err := mc.getMetricData()
	if err != nil {
		return
	}

	// set the metric data into the respective fields
	mc.metrics.SetMetrics(blockDevices)

	// collect each metric
	for _, col := range mc.metrics.Collectors() {
		col.Collect(ch)
	}
}

// getMetricData is used to get the metric data from the source
func (mc *StaticMetricCollector) getMetricData() ([]blockdevice.BlockDevice, error) {
	// list all the BDs from etcd
	bdList := &apis.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}
	err := mc.Client.List(context.TODO(), nil, bdList)
	if err != nil {
		glog.Error("error in listing BDs. ", err)
		return nil, err
	}

	// convert the BD apis to BlockDevice struct used by NDM internally
	blockDevices := make([]blockdevice.BlockDevice, 0)
	for _, bd := range bdList.Items {
		blockDevice := blockdevice.BlockDevice{
			NodeAttributes: make(blockdevice.NodeAttribute, 0),
		}
		// copy values from api to BlockDevice struct
		blockDevice.UUID = bd.Name
		blockDevice.NodeAttributes[blockdevice.HostName] = bd.Labels[controller.KubernetesHostNameLabel]
		blockDevice.Path = bd.Spec.Path
		blockDevices = append(blockDevices, blockDevice)
	}
	return blockDevices, nil
}
