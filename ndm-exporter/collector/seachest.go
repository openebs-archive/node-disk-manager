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
	smartmetrics "github.com/openebs/node-disk-manager/pkg/metrics/smart"
	"github.com/openebs/node-disk-manager/pkg/seachest"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog"
)

const (
	// SeachestCollectorNamespace is the namespace field in the prometheus metrics when
	// seachest is used to collect the metrics.
	SeachestCollectorNamespace = "seachest"
)

// SeachestCollector contains the metrics, concurrency handler and client to get the
// metrics from seachest
type SeachestCollector struct {
	// Client is the k8s client which will be used to interface with etcd
	Client kubernetes.Client

	// concurrency handling
	sync.Mutex
	requestInProgress bool

	// all metrics collected via seachest
	metrics *smartmetrics.Metrics
}

// SeachestMetricData is the struct which holds the data from seachest library
// corresponding to each blockdevice
type SeachestMetricData struct {
	SeachestIdentifier   *seachest.Identifier
	TempInfo             blockdevice.TemperatureInformation
	Capacity             uint64
	TotalBytesRead       uint64
	TotalBytesWritten    uint64
	DeviceUtilization    float64
	PercentEnduranceUsed float64
}

// NewSeachestMetricCollector creates a new instance of SeachestCollector which
// implements Collector interface
func NewSeachestMetricCollector(c kubernetes.Client) prometheus.Collector {
	klog.V(2).Infof("Seachest Metric Collector initialized")
	sc := &SeachestCollector{
		Client:  c,
		metrics: smartmetrics.NewMetrics(SeachestCollectorNamespace),
	}
	sc.metrics.WithBlockDeviceCurrentTemperature().
		WithBlockDeviceCurrentTemperatureValid().
		WithBlockDeviceHighestTemperatureValid().
		WithBlockDeviceLowestTemperatureValid().
		WithBlockDeviceHighestTemperature().
		WithBlockDeviceLowestTemperature().
		WithBlockDeviceCapacity().
		WithBlockDeviceTotalBytesRead().
		WithBlockDeviceTotalBytesWritten().
		WithBlockDeviceUtilizationRate().
		WithBlockDevicePercentEnduranceUsed().
		WithRejectRequest().
		WithErrorRequest()
	return sc
}

// Describe is the implementation of Describe in prometheus.Collector
func (sc *SeachestCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, col := range sc.metrics.Collectors() {
		col.Describe(ch)
	}
}

// Collect is the implementation of Collect in prometheus.Collector
func (sc *SeachestCollector) Collect(ch chan<- prometheus.Metric) {
	klog.V(4).Info("Starting to collect smartmetrics metrics for a request")

	sc.Lock()
	if sc.requestInProgress {
		klog.V(4).Info("Another request already in progress.")
		sc.metrics.IncRejectRequestCounter()
		sc.Unlock()
		return
	}

	sc.requestInProgress = true
	sc.Unlock()

	// once a request is processed, set the progress flag to false
	defer sc.setRequestProgressToFalse()

	klog.V(4).Info("Setting client for this request.")

	// set the client each time
	if err := sc.Client.InitClient(); err != nil {
		klog.Errorf("error setting client. %v", err)
		sc.metrics.IncErrorRequestCounter()
		sc.collectErrors(ch)
		return
	}

	// get list of blockdevices from etcd
	blockDevices, err := sc.Client.ListBlockDevice()
	if err != nil {
		klog.Errorf("Listing block devices failed %v", err)
		sc.metrics.IncErrorRequestCounter()
		sc.collectErrors(ch)
		return
	}

	klog.V(4).Info("Blockdevices fetched from etcd")

	err = getMetricData(blockDevices)
	if err != nil {
		sc.metrics.IncErrorRequestCounter()
		sc.collectErrors(ch)
		return
	}

	klog.V(4).Infof("metrics data obtained from seachest library")

	sc.setMetricData(blockDevices)

	klog.V(4).Info("Prometheus metrics is set and initializing collection.")

	// collect each metric
	for _, col := range sc.metrics.Collectors() {
		col.Collect(ch)
	}
}

// setRequestProgressToFalse is used to set the progress flag, when a request is
// processed or errored
func (sc *SeachestCollector) setRequestProgressToFalse() {
	sc.Lock()
	sc.requestInProgress = false
	sc.Unlock()
}

// collectErrors collects only the error metrics and set it on the channel
func (sc *SeachestCollector) collectErrors(ch chan<- prometheus.Metric) {
	for _, col := range sc.metrics.ErrorCollectors() {
		col.Collect(ch)
	}
}

// getMetricData gets the seachest metrics for each blockdevice and fills it in the blockdevice struct
func getMetricData(bds []blockdevice.BlockDevice) error {
	var err error
	ok := false
	for i, bd := range bds {
		// do not report metrics for sparse devices
		if bd.DeviceAttributes.DeviceType == blockdevice.SparseBlockDeviceType {
			continue
		}
		sc := SeachestMetricData{
			SeachestIdentifier: &seachest.Identifier{
				DevPath: bd.DevPath,
			},
		}
		err = sc.getSeachestData()
		if err != nil {
			klog.Errorf("fetching seachest data for %s failed. %v", bd.DevPath, err)
			continue
		}
		ok = true

		bds[i].SMARTInfo.TemperatureInfo = sc.TempInfo
		bds[i].Capacity.Storage = sc.Capacity
		bds[i].SMARTInfo.TotalBytesRead = sc.TotalBytesRead
		bds[i].SMARTInfo.TotalBytesWritten = sc.TotalBytesWritten
		bds[i].SMARTInfo.UtilizationRate = sc.DeviceUtilization
		bds[i].SMARTInfo.PercentEnduranceUsed = sc.PercentEnduranceUsed

	}
	if !ok {
		return fmt.Errorf("getting seachest metrics for the blockdevices failed")
	}
	return nil
}

// getSeachestData fetches the data for a blockdevice using the seachest library from the disk.
func (sc *SeachestMetricData) getSeachestData() error {
	driveInfo, err := sc.SeachestIdentifier.SeachestBasicDiskInfo()
	if err != 0 {
		klog.Errorf("error fetching basic disk info using seachest. %s", seachest.SeachestErrors(err))
		return fmt.Errorf("error getting seachest data for metrics. %s", seachest.SeachestErrors(err))
	}

	sc.TempInfo.CurrentTemperatureDataValid = sc.SeachestIdentifier.GetTemperatureDataValidStatus(driveInfo)
	sc.TempInfo.CurrentTemperature = sc.SeachestIdentifier.GetCurrentTemperature(driveInfo)
	sc.TempInfo.LowestTemperature = sc.SeachestIdentifier.GetLowestTemperature(driveInfo)
	sc.TempInfo.HighestTemperature = sc.SeachestIdentifier.GetHighestTemperature(driveInfo)
	sc.TempInfo.HighestTemperatureDataValid = sc.SeachestIdentifier.GetHighestValid(driveInfo)
	sc.TempInfo.LowestTemperatureDataValid = sc.SeachestIdentifier.GetLowestValid(driveInfo)
	sc.Capacity = sc.SeachestIdentifier.GetCapacity(driveInfo)
	sc.TotalBytesRead = sc.SeachestIdentifier.GetTotalBytesRead(driveInfo)
	sc.TotalBytesWritten = sc.SeachestIdentifier.GetTotalBytesWritten(driveInfo)
	sc.DeviceUtilization = sc.SeachestIdentifier.GetDeviceUtilizationRate(driveInfo)
	sc.PercentEnduranceUsed = sc.SeachestIdentifier.GetPercentEnduranceUsed(driveInfo)

	klog.V(4).Infof("Device is : %v", sc.SeachestIdentifier.DevPath)
	klog.V(4).Infof("Current temperature is %v", sc.TempInfo.CurrentTemperature)
	klog.V(4).Infof("Lowest temperature is %v", sc.TempInfo.LowestTemperature)
	klog.V(4).Infof("Highest temperature is %v", sc.TempInfo.HighestTemperature)
	klog.V(4).Infof("Capacity is %v", sc.Capacity)
	klog.V(4).Infof("Total bytes read is %v", sc.TotalBytesRead)
	klog.V(4).Infof("Total bytes written is %v", sc.TotalBytesWritten)
	klog.V(4).Infof("Device utilization rate is %v", sc.DeviceUtilization)
	klog.V(4).Infof("Endurance used is %v", sc.PercentEnduranceUsed)

	return nil
}

// setMetricData sets the SMART metric data collected using seachest onto
// the prometheus metrics
func (sc *SeachestCollector) setMetricData(blockdevices []blockdevice.BlockDevice) {
	for _, bd := range blockdevices {
		// sets the label values
		sc.metrics.WithBlockDeviceUUID(bd.UUID).
			WithBlockDevicePath(bd.DevPath).
			WithBlockDeviceHostName(bd.NodeAttributes[blockdevice.HostName]).
			WithBlockDeviceNodeName(bd.NodeAttributes[blockdevice.NodeName])
		// sets the metrics
		sc.metrics.SetBlockDeviceCurrentTemperature(bd.SMARTInfo.TemperatureInfo.CurrentTemperature).
			SetBlockDeviceHighestTemperature(bd.SMARTInfo.TemperatureInfo.HighestTemperature).
			SetBlockDeviceLowestTemperature(bd.SMARTInfo.TemperatureInfo.LowestTemperature).
			SetBlockDeviceCurrentTemperatureValid(bd.SMARTInfo.TemperatureInfo.CurrentTemperatureDataValid).
			SetBlockDeviceHighestTemperatureValid(bd.SMARTInfo.TemperatureInfo.HighestTemperatureDataValid).
			SetBlockDeviceLowestTemperatureValid(bd.SMARTInfo.TemperatureInfo.LowestTemperatureDataValid).
			SetBlockDeviceCapacity(bd.Capacity.Storage).
			SetBlockDeviceUtilizationRate(bd.SMARTInfo.UtilizationRate).
			SetBlockDeviceTotalBytesRead(bd.SMARTInfo.TotalBytesRead).
			SetBlockDeviceTotalBytesWritten(bd.SMARTInfo.TotalBytesWritten).
			SetBlockDevicePercentEnduranceUsed(bd.SMARTInfo.PercentEnduranceUsed)

	}
}
