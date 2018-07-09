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

package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMsignificantBit(t *testing.T) {
	tests := map[string]struct {
		inputnumber uint
		expected    int
	}{
		"input 0 should return 0":           {inputnumber: 0, expected: 0},
		"input 1 should return 1":           {inputnumber: 1, expected: 0},
		"input 100 should return 100":       {inputnumber: 100, expected: 6},
		"input 100100 should return 100100": {inputnumber: 100100, expected: 16},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, MSignificantBit(test.inputnumber))
		})
	}
}

func TestCollect(t *testing.T) {
	collector := NewErrorCollector()
	inputKey := "ScsiInqErr"
	inputval := fmt.Errorf("error in scsi inquiry command")

	tests := map[string]struct {
		inputErrorVal error
		expected      bool
	}{
		"giving nil as an input":      {inputErrorVal: nil, expected: false},
		"giving an error as an input": {inputErrorVal: inputval, expected: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, collector.Collect(inputKey, test.inputErrorVal))
		})
	}
}

func TestError(t *testing.T) {
	collector := NewErrorCollector()
	errorMap := make(map[string]error)
	errorMap["ScsiInqErr"] = fmt.Errorf("Error in scsi Inquiry")
	collector.Collect("ScsiInqErr", fmt.Errorf("Error in scsi Inquiry"))

	tests := map[string]struct {
		expected map[string]error
	}{
		"calling Error func should return error collected by collect": {expected: errorMap},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, collector.Error())
		})
	}
}

func TestNewErrorCollector(t *testing.T) {
	expectedVal := &ErrorCollector{errors: map[string]error{}}
	tests := map[string]struct {
		expected *ErrorCollector
	}{
		"calling NewErrorCollector returns reference to errorCollector struct": {expected: expectedVal},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, NewErrorCollector())
		})
	}
}
