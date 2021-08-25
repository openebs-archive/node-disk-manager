/*
Copyright 2020 The OpenEBS Authors

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
	"errors"

	"github.com/openebs/node-disk-manager/blockdevice"
	"k8s.io/klog"
)

func (pe *ProbeEvent) changeBlockDevice(bd *blockdevice.BlockDevice, requestedProbes ...string) error {
	pe.Controller.FillBlockDeviceDetails(bd, requestedProbes...)
	if bd.UUID == "" {
		uuid, ok := generateUUID(*bd)
		if !ok {
			klog.Error("could no generate uuid for device. aborting")
			return errors.New("could not indentify device uniquely")
		}
		bd.UUID = uuid
	}
	// add labels to block device that may be helpful for filtering the block device
	// based on some/generic attributes like drive-type, model, vendor etc.
	pe.addBlockDeviceLabels(*bd)
	pe.addBlockDeviceToHierarchyCache(*bd)
	if !pe.Controller.ApplyFilter(bd) {
		return nil
	}
	apiBlockdevice := pe.Controller.NewDeviceInfoFromBlockDevice(bd).ToDevice()
	apiBlockdevice.SetNamespace(pe.Controller.Namespace)
	klog.Info("updating bd: ", apiBlockdevice.GetName())
	return pe.Controller.UpdateBlockDevice(apiBlockdevice, nil)
}
