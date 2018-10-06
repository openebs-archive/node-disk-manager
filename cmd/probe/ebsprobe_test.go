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

package probe

import (
	"os"
	"strconv"
	"testing"

	"github.com/openebs/node-disk-manager/cmd/controller"
	"github.com/stretchr/testify/assert"
)

func TestEbsFillDisk(t *testing.T) {
	probe := &ebsProbe{}
	disk := &controller.DiskInfo{}
	tempSysPath := "/tmp"
	os.MkdirAll(tempSysPath+"/queue", 0700)
	file, err := os.Create(tempSysPath + "/queue/hw_sector_size")
	if err != nil {
		t.Fatalf("unable to write file to %s %v", tempSysPath, err)
		return
	}
	file.Write([]byte("10"))

	disk.ProbeIdentifiers.UdevIdentifier = tempSysPath
	disk.Size = strconv.Itoa(10)
	probe.FillDiskDetails(disk)
	assert.Equal(t, disk.Capacity, uint64(100))
}
