/*
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNodeCollector(t *testing.T) {
	nc, err := NewNodeCollector()
	if err != nil {
		t.Error(err)
	}
	fakeCollectors := []string{"diskstats"}
	collectors := []string{}
	for n := range nc.Collectors {
		collectors = append(collectors, n)
	}
	assert.Equal(t, fakeCollectors, collectors)
}

func TestRegisterCollector(t *testing.T) {
	registerCollector("diskstats", defaultEnabled, NewDiskstatsCollector)
	assert.Equal(t, true, *collectorState["diskstats"])
	registerCollector("diskstats", defaultDisabled, NewDiskstatsCollector)
	assert.Equal(t, false, *collectorState["diskstats"])
}
