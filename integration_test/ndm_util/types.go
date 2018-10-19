package ndmutil

import (
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"strings"
)

// ConfigMap contains configs of probes and filters
type ConfigMap struct {
	ProbeConfigs  []ProbeConfig  `json:"probeconfigs"`  // ProbeConfigs contains configs of Probes
	FilterConfigs []FilterConfig `json:"filterconfigs"` // FilterConfigs contains configs of Filters
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

func GetNDMConfig(fileName string) ConfigMap {
	var configMap ConfigMap
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
	}
	err = yaml.Unmarshal(data, &configMap)
	if err != nil {
		fmt.Println(err)
	}
	yamlString := string(data)
	yamlString = strings.Split(yamlString, "---")[0]
	yamlString = strings.Split(yamlString, "node-disk-manager.config: |")[1]
	yaml.Unmarshal([]byte(yamlString), &configMap)
	return configMap
}
