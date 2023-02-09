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

package controller

import (
	"encoding/json"
	"io/ioutil"

	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	// DefaultConfigFilePath is the default path at which config is present inside
	// container
	DefaultConfigFilePath = "/host/node-disk-manager.config"
)

// NodeDiskManagerConfig contains configs of probes and filters
type NodeDiskManagerConfig struct {
	ProbeConfigs  []ProbeConfig  `json:"probeconfigs"`  // ProbeConfigs contains configs of Probes
	FilterConfigs []FilterConfig `json:"filterconfigs"` // FilterConfigs contains configs of Filters
	// TagConfigs contains configs for tags
	TagConfigs []TagConfig `json:"tagconfigs"`
	// MetaConfig contains configs for device labels
	MetaConfigs []MetaConfig `json:"metaconfigs"`
}

// ProbeConfig contains configs of Probe
type ProbeConfig struct {
	Key   string `json:"key"`   // Key is key for each Probe
	Name  string `json:"name"`  // Name is name of Probe
	State string `json:"state"` // State is state of Probe
}

// FilterConfig contains configs of Filter
type FilterConfig struct {
	Key     string `json:"key"`     // Key is key for each Filter
	Name    string `json:"name"`    // Name is name of Filer
	State   string `json:"state"`   // State is state of Filter
	Include string `json:"include"` // Include contains , separated values which we want to include for filter
	Exclude string `json:"exclude"` // Exclude contains , separated values which we want to exclude for filter
}

type TagConfig struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Pattern string `json:"pattern"`
	TagName string `json:"tag"`
}

type MetaConfig struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Pattern string `json:"pattern"`
}

// SetNDMConfig sets config for probes and filters which user provides via configmap. If
// no configmap present then ndm will load default config for each probes and filters.
func (c *Controller) SetNDMConfig(opts NDMOptions) {
	data, err := ioutil.ReadFile(opts.ConfigFilePath)
	if err != nil {
		c.NDMConfig = nil
		klog.Error("unable to set ndm config : ", err)
		return
	}

	var ndmConfig NodeDiskManagerConfig
	if json.Valid(data) {
		err = json.Unmarshal(data, &ndmConfig)
	} else {
		err = yaml.Unmarshal(data, &ndmConfig)
	}
	if err != nil {
		c.NDMConfig = nil
		klog.Error("unable to set ndm config : ", err)
		return
	}

	c.NDMConfig = &ndmConfig
}
