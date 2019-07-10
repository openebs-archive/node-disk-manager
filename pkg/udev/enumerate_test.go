/*
Copyright 2018 OpenEBS Authors.

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

package udev

import (
	"testing"
)

func TestAddSubsystemFilter(t *testing.T) {
	newUdev, err := NewUdev()
	if err != nil {
		t.Fatal(err)
	}
	defer newUdev.UnrefUdev()
	newUdevEnumerate, err := newUdev.NewUdevEnumerate()
	if err != nil {
		t.Fatal(err)
	}
	defer newUdevEnumerate.UnrefUdevEnumerate()
	err = newUdevEnumerate.AddSubsystemFilter("block")
	if err != nil {
		t.Error("error should be nil for successful subsystem filter")
	}
}

func TestScanDevices(t *testing.T) {
	newUdev, err := NewUdev()
	if err != nil {
		t.Fatal(err)
	}
	defer newUdev.UnrefUdev()
	newUdevEnumerate, err := newUdev.NewUdevEnumerate()
	if err != nil {
		t.Fatal(err)
	}
	defer newUdevEnumerate.UnrefUdevEnumerate()
	err = newUdevEnumerate.AddSubsystemFilter("block")
	if err != nil {
		t.Error("error should be nil for successful subsystem filter")
	}
	err = newUdevEnumerate.ScanDevices()
	if err != nil {
		t.Error("error should be nil for successful scan device")
	}
}
