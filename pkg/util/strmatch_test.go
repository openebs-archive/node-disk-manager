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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	diskList := make([]string, 0)
	diskList = append(diskList, "Key1")
	diskList = append(diskList, "Key3")
	tests := map[string]struct {
		diskName string
		expected bool
	}{
		"giving a key which is not present in slice": {diskName: "Key0", expected: false},
		"giving a key which is present in slice":     {diskName: "Key3", expected: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, Contains(diskList, test.diskName))
		})
	}
}

func TestContainsIgnoredCase(t *testing.T) {
	diskList := make([]string, 0)
	diskList = append(diskList, "Key1")
	diskList = append(diskList, "Key3")
	tests := map[string]struct {
		diskName string
		expected bool
	}{
		"giving a key which is not present in slice": {diskName: "keY0", expected: false},
		"giving a key which is present in slice":     {diskName: "KEy3", expected: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, ContainsIgnoredCase(diskList, test.diskName))
		})
	}
}

func TestMatchIgnoredCase(t *testing.T) {
	mkList := make([]string, 0)
	mkList = append(mkList, "loop")
	mkList = append(mkList, "/dev/sr0")
	tests := map[string]struct {
		diskPath string
		expected bool
	}{
		"diskPath contains one of the keys ": {diskPath: "/dev/loop0", expected: true},
		"diskPath matches complete key":      {diskPath: "/dev/sr0", expected: true},
		"diskPath does not match any keys":   {diskPath: "/dev/sdb", expected: false},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, MatchIgnoredCase(mkList, test.diskPath))
		})
	}
}

func TestMatchRegex(t *testing.T) {
	tests := map[string]struct {
		regex    string
		str      string
		expected bool
	}{
		"string matches without the need of regex portion": {
			regex:    "/dev/sda(([0-9]*|p[0-9]+))$",
			str:      "/dev/sda",
			expected: true,
		},
		"string matches with regex": {
			regex:    "/dev/sda(([0-9]*|p[0-9]+))$",
			str:      "/dev/sda1",
			expected: true,
		},
		"string does not match the regex": {
			regex:    "/dev/sda(([0-9]*|p[0-9]+))$",
			str:      "/dev/sdaa",
			expected: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, MatchRegex(test.regex, test.str))
		})
	}
}
