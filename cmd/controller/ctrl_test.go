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

package controller_test

import (
	"github.com/openebs/node-disk-manager/pkg/udev"
	"os/exec"
	"strings"
	"testing"
)

func TestDevList(t *testing.T) {
	udevices := udev.ListDevices()
	if udevices == nil {
		t.Fatal("not able to list the devices")
	}

	disks, err := exec.Command("lsblk", "-r", "-d", "-n", "-o", "name").Output()
	if err != nil {
		t.Fatal(err)
	}

	for _, device := range strings.Split(string(disks[:len(disks)-1]), "\n") {
		var udevice *udev.Udevice
		for _, udevice = range udevices {
			if "/dev/"+device == udevice.Devnode() {
				break
			}
		}
		if "/dev/"+device != udevice.Devnode() {
			t.Fatal("deivce not detected by udev", "/dev/"+device)
		}
	}
}
