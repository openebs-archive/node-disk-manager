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

package controller

import (
	"strings"

	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/openebs/node-disk-manager/pkg/util"
)

const (
	models = "EphemeralDisk,Virtual_disk"
)

func (c *Controller) GenerateUuid(device *libudevwrapper.UdevDevice) string {
	deviceDetails := device.DiskInfoFromLibudev()
	uid := deviceDetails.Wwn + deviceDetails.Model + deviceDetails.Serial + deviceDetails.Vendor
	idtype := deviceDetails.Type
	model := deviceDetails.Model
	// Virtual disks either have no attributes or they all have
	// the same attributes. Adding hostname in uid so that disks from different
	// nodes can be differentiated. Also, putting devpath in uid so that disks
	// from the same node also can be differentiated.
	// 	On Gke, we have the ID_TYPE property, but still disks will have
	// same attributes. We have to put a special check to handle it and process
	// it like a Virtual disk.
	localDiskModels := make([]string, 0)
	if c.NDMConfig != nil {
		localDiskModels = strings.Split(c.NDMConfig.Data.LocalDiskModels, ",")
	} else {
		localDiskModels = strings.Split(models, ",")
	}
	if len(idtype) == 0 || util.Contains(localDiskModels, model) {
		uid += c.HostName + deviceDetails.Path
	}
	return libudevwrapper.NDMPrefix + util.Hash(uid)
}
