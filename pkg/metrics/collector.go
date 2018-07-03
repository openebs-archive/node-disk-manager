/*
Copyright 2015 The Prometheus Authors.
Copyright 2018 The OpenEBS Authors.

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

package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace       = "ndm"
	defaultEnabled  = true
	defaultDisabled = false
)

var (
	factories      = make(map[string]func() (Collector, error))
	collectorState = make(map[string]*bool)
)

// Collector is the interface a collector(eg- diskstat collector) has to implement it.
type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Update(ch chan<- prometheus.Metric) error
}

type typedDesc struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}

// nodeCollector implements the prometheus.Collector interface.
// 1. diskstats collector collects iostats of disks like total read/write completed,
// average read/write size, read/write latency, total read/write size.
type nodeCollector struct {
	Collectors map[string]Collector
}

func registerCollector(collector string, isDefaultEnabled bool, factory func() (Collector, error)) {
	collectorState[collector] = &isDefaultEnabled
	factories[collector] = factory
}

// NewNodeCollector creates a new NodeCollector
func NewNodeCollector() (*nodeCollector, error) {
	f := make(map[string]bool)
	collectors := make(map[string]Collector)
	for key, enabled := range collectorState {
		if *enabled {
			collector, err := factories[key]()
			if err != nil {
				return nil, err
			}
			if len(f) == 0 || f[key] {
				collectors[key] = collector
			}
		}
	}
	return &nodeCollector{Collectors: collectors}, nil
}

// Describe implements the prometheus.Collector interface.
func (n nodeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collectorUptimeDesc
}

// Collect implements the prometheus.Collector interface.
func (n nodeCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(n.Collectors))
	for name, c := range n.Collectors {
		go func(name string, c Collector) {
			execute(name, c, ch)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

// execute ...
func execute(name string, c Collector, ch chan<- prometheus.Metric) {
	c.Update(ch)
	duration := time.Since(StartingTime)
	ch <- prometheus.MustNewConstMetric(collectorUptimeDesc, prometheus.GaugeValue, duration.Seconds(), name)
}
