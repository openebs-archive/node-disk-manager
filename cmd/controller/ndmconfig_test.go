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
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetNDMConfig(t *testing.T) {
	fakeConfigFilePath := "/tmp/fakendm.config"
	defaultConfigFilePath := ConfigFilePath

	// test case : when file not present
	fakeController := &Controller{}
	fakeController.SetNDMConfig()
	if fakeController.NDMConfig != nil {
		t.Error("NDMConfig should nil for invalid file path")
	}
	// test case for invalid json
	ConfigFilePath = fakeConfigFilePath
	fileContent := []byte(`{
        "probeconfigs": [
            {
            "key": "udev-probe",
            "name": "udev probe",
            "state": "true"
            },
        ]
    }`)
	err := ioutil.WriteFile(fakeConfigFilePath, fileContent, 0644)
	if err != nil {
		t.Fatal(err)
	}
	if fakeController.NDMConfig != nil {
		t.Error("NDMConfig should nil for invalid file path")
	}
	// test case for valid json
	fileContent = []byte(`{
    "probeconfigs": [
        {
        "key": "udev-probe",
        "name": "udev probe",
        "state": "true"
        }
    ],
    "filterconfigs": [
        {
        "key": "os-disk-exclude-filter",
        "name": "os disk exclude filter",
        "state": "true"
        }
    ]
}`)
	expectedProbeConfig := ProbeConfig{
		Key:   "udev-probe",
		Name:  "udev probe",
		State: "true",
	}
	expectedFilterConfig := FilterConfig{
		Key:   "os-disk-exclude-filter",
		Name:  "os disk exclude filter",
		State: "true",
	}
	expectedNDMConfig := NodeDiskManagerConfig{
		FilterConfigs: make([]FilterConfig, 0),
		ProbeConfigs:  make([]ProbeConfig, 0),
	}
	expectedNDMConfig.FilterConfigs = append(expectedNDMConfig.FilterConfigs, expectedFilterConfig)
	expectedNDMConfig.ProbeConfigs = append(expectedNDMConfig.ProbeConfigs, expectedProbeConfig)
	ConfigFilePath = fakeConfigFilePath
	err = ioutil.WriteFile(fakeConfigFilePath, fileContent, 0644)
	if err != nil {
		t.Fatal(err)
	}
	fakeController.SetNDMConfig()
	assert.Equal(t, expectedNDMConfig, *fakeController.NDMConfig)
	os.Remove(fakeConfigFilePath)
	ConfigFilePath = defaultConfigFilePath
}

func TestController_SetNDMConfig_yaml(t *testing.T) {
	fakeConfigFilePath := "/tmp/fakendm-yaml.config"
	writeTestYaml(t, fakeConfigFilePath)
	defer os.Remove(fakeConfigFilePath)

	ConfigFilePath = fakeConfigFilePath

	ctrl := &Controller{}
	ctrl.SetNDMConfig()

	assert.NotNil(t, ctrl.NDMConfig)

	assert.Len(t, ctrl.NDMConfig.ProbeConfigs, 1)
	expectedProbeConfig := ProbeConfig{Key: "udev-probe", Name: "udev probe", State: "true"}
	assert.Equal(t, expectedProbeConfig, ctrl.NDMConfig.ProbeConfigs[0])

	assert.Len(t, ctrl.NDMConfig.FilterConfigs, 1)
	expectedFilterConfig := FilterConfig{Key: "os-disk-exclude-filter", Name: "os disk exclude filter", State: "true", Include: "", Exclude: "/,/etc/hosts,/boot"}
	assert.Equal(t, expectedFilterConfig, ctrl.NDMConfig.FilterConfigs[0])
}

func writeTestYaml(t *testing.T, fpath string) {
	data := `
probeconfigs:
  - key: udev-probe
    name: udev probe
    state: true
filterconfigs:
  - key: os-disk-exclude-filter
    name: os disk exclude filter
    state: true
    include: ""
    exclude: /,/etc/hosts,/boot
`

	err := ioutil.WriteFile(fpath, []byte(data), 0644)
	assert.NoError(t, err)
}
