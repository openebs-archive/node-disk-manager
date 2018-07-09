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

package smart

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectBusType(t *testing.T) {

	tests := map[string]struct {
		input    string
		expected string
	}{
		"Get bus type for devPath /dev/sda":              {input: "/dev/sda", expected: "SCSI"},
		"Get bus type for devPath /dev/hd":               {input: "/dev/hd", expected: "IDE"},
		"Get bus type for devPath /dev/nvme0n0":          {input: "/dev/nvme0n0", expected: "NVMe"},
		"Get bus type for some arbitrary devPath /sa/ds": {input: "/sa/ds", expected: "unknown"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, detectBusType(test.input))
		})
	}
}

func TestIsConditionSatisfied(t *testing.T) {
	errValOnBlankPath := fmt.Errorf("no disk device path given to get the disk details")
	tests := map[string]struct {
		input    string
		expected error
	}{
		"giving empty device path returns error": {input: "", expected: errValOnBlankPath},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, isConditionSatisfied(test.input))
		})
	}
}
