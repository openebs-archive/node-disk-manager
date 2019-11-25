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
	"fmt"
	"sync"

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/db/kubernetes"
	seachestmetrics "github.com/openebs/node-disk-manager/pkg/metrics/seachest"
	"github.com/openebs/node-disk-manager/pkg/seachest"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog"
)

// SeachestMetricCollector contains the metrics, concurrency handler and client to get the
// metrics from seachest
type SeachestMetricCollector struct {
	// Client is the k8s client which will be used to interface with etcd
	Client kubernetes.Client

	// concurrency handling
	sync.Mutex
	requestInProgress bool

	// all exposed metrics from seachest
	metrics *seachestmetrics.Metrics
}

// SeachestMetricData is the struct which holds the data from seachest library
// correspnding to each blockdevice
type SeachestMetricData struct {
	SeachestIdentifier *seachest.Identifier
	TempInfo           blockdevice.TemperatureInformation
}

// NewSeachestMetricCollector creates a new instance of SeachestMetricCollector which
// implements Collector interface
func NewSeachestMetricCollector(c kubernetes.Client) prometheus.Collector {
	klog.V(2).Infof("Seachest Metric Collector initialized")
	return &SeachestMetricCollector{
		Client:  c,
		metrics: seachestmetrics.NewMetrics(),
	}
}

// Describe is the implementation of Describe in prometheus.Collector
func (mc *SeachestMetricCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, col := range mc.metrics.Collectors() {
		col.Describe(ch)
	}
}

// Collect is the implementation of Collect in prometheus.Collector
func (mc *SeachestMetricCollector) Collect(ch chan<- prometheus.Metric) {
	klog.V(4).Info("Starting to collect seachestmetrics metrics for a request")

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

	klog.V(4).Info("Blockdevices fetched from etcd")

	err = GetMetricData(blockDevices)
	if err != nil {
		mc.metrics.IncErrorRequestCounter()
		mc.collectErrors(ch)
		return
	}

	klog.V(4).Infof("metrics data obtained from seachest library")

	mc.metrics.SetMetrics(blockDevices)

	klog.V(4).Info("Prometheus metrics is set and initializing collection.")

	// collect each metric
	for _, col := range mc.metrics.Collectors() {
		col.Collect(ch)
	}
}

// setRequestProgressToFalse is used to set the progress flag, when a request is
// processed or errored
func (mc *SeachestMetricCollector) setRequestProgressToFalse() {
	mc.Lock()
	mc.requestInProgress = false
	mc.Unlock()
}

// collectErrors collects only the error metrics and set it on the channel
func (mc *SeachestMetricCollector) collectErrors(ch chan<- prometheus.Metric) {
	for _, col := range mc.metrics.ErrorCollectors() {
		col.Collect(ch)
	}
}

// GetMetricData gets the seachest metrics for each blockdevice and fills it in the blockdevice struct
func GetMetricData(bds []blockdevice.BlockDevice) error {
	var err error
	ok := false
	for i, bd := range bds {
		sc := SeachestMetricData{
			SeachestIdentifier: &seachest.Identifier{
				DevPath: bd.Path,
			},
		}
		err = sc.GetSeachestData()
		if err != nil {
			klog.Errorf("fetching seachest data for %s failed. %v", bd.Path, err)
			continue
		}
		ok = true
		bds[i].TemperatureInfo = sc.TempInfo
	}
	if !ok {
		return fmt.Errorf("getting seachest metrics for the blockdevices failed")
	} else {
		return nil
	}
}

// GetSeachestData fetches the data for a blockdevice using the seachest library from the disk.
func (sc *SeachestMetricData) GetSeachestData() error {
	driveInfo, err := sc.SeachestIdentifier.SeachestBasicDiskInfo()
	if err != 0 {
		klog.Errorf("error fetching basic disk info using seachest. %s", seachest.SeachestErrors(err))
		return fmt.Errorf("error getting seachest data for metrics. %s", seachest.SeachestErrors(err))
	}
	sc.TempInfo.TemperatureDataValid = sc.SeachestIdentifier.GetTemperatureDataValidStatus(driveInfo)
	sc.TempInfo.CurrentTemperature = sc.SeachestIdentifier.GetCurrentTemperature(driveInfo)

	return nil
}
