## Write a new Collector for NDM Exporter
To write a new collector certain steps have to be followed.

1. Create a new file in `ndm-exporter/collector/my-collector.go`. This file will include the struct which will implement prometheus collector interface.
```
type MyMetricCollector struct {
	Client kubernetes.Client

	sync.Mutex
	requestInProgress bool

	metrics *seachestmetrics.Metrics
}
```
The fields are 
- `Client` : used to interface with the database from which blockdevice information can be obtained
- `sync.Mutex` and `requestInProgress` : to aquire lock on the collector so that at a time only a single request is being handled
- `metrics` : The actual prometheus metrics fields. This struct will have the metrics that are exposed.

2. A `Collect()` and `Describe` method need to be implemented for this struct.
The `Describe()` will list out all the metrics, and `Collect()` will have the logic to fetch the metrics. All metrics fetched will be first
converted into a `BlockDevice` struct, which is used to represent a blockdevice in NDM.

3. Create a new package for your new metrics and add one file. `pkg/metrics/mymetrics/metrics.go`. This fill will contain the struct for your metrics.
```
type Metrics struct {
	mymetric1 *prometheus.GaugeVec
	mymetric2 *prometheus.GaugeVec

	rejectRequestCount prometheus.Counter
	errorRequestCount  prometheus.Counter
}
```
`mymetric1` and `mymetric2` are the metrics that will be exposed along with the rejected requests count and errored request count. 
Requests are rejected if a request is already in progress. Request errors when an error occurs during the collection of metrics

`Metrics` struct should have a `SetMetric(bd []BlockDevice)` method. This method will set all the metrics into the `Metric` struct from the `BlockDevice` object.

4. Once your new collector is implemented, it should be registered with the exporter. NDM has 2 types of exporter, one running at cluster
level and other at node level. register your collector with the exporter, depending on where you need the collector to be run.
`prometheus.MustRegister(myCollector1, myCollector2)`