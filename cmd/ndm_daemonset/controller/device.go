/*
Copyright 2019 OpenEBS Authors.

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
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	//"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
 * DeviceInfo contains details of device which can be converted into api.Device
 * There will be one DeviceInfo created for each physical disk (may change in
 * future). At the end it is converted to Device struct which will be pushed to
 * etcd as a CR of that device.
 */
type DeviceInfo struct {
	HostName           string   // Node's name to which backing disk is attached.
	Uuid               string   // Uuid of backing disk
	Capacity           uint64   // Capacity of device
	Model              string   // Do device have model ??
	Serial             string   // Do device have serial no ??
	Vendor             string   // Vendor of device
	Path               string   // device Path like /dev/sda
	ByIdDevLinks       []string // ByIdDevLinks contains by-id devlinks
	ByPathDevLinks     []string // ByPathDevLinks contains by-path devlinks
	FirmwareRevision   string   // FirmwareRevision is the firmware revision for a disk
	LogicalSectorSize  uint32   // LogicalSectorSize is the Logical size of device sector in bytes
	PhysicalSectorSize uint32   // PhysicalSectorSize is the Physical size of device sector in bytes
	Compliance         string   // Compliance is implemented specifications version i.e. SPC-1, SPC-2, etc
	DeviceType         string   // DeviceType represents the type of backing disk
}

/*
 * NewDeviceInfo returns a pointer of empty DeviceInfo
 * struct which will be field from DiskInfo.
 */
func NewDeviceInfo() *DeviceInfo {
	deviceInfo := &DeviceInfo{}
	return deviceInfo
}

/*
 * ToDevice convert deviceInfo struct to api.Device
 * type which will be pushed to etcd
 */
func (di *DeviceInfo) ToDevice() apis.Device {
	dvr := apis.Device{}
	dvr.Spec = di.getDeviceSpec()
	dvr.ObjectMeta = di.getObjectMeta()
	dvr.TypeMeta = di.getTypeMeta()
	dvr.ClaimState = di.getClaimState()
	dvr.Status = di.getStatus()
	return dvr
}

/*
 * getObjectMeta returns ObjectMeta struct which contains
 * labels and Name of resource. It is used to populate data
 * of Device struct of Device CR.
 */
func (di *DeviceInfo) getObjectMeta() metav1.ObjectMeta {
	/*
		namespace, err := k8sutil.GetWatchNamespace()
		if err != nil && namespace == "" {
			namespace = "default"
		}
	*/
	objectMeta := metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   di.Uuid,
		//Namespace: namespace,
	}
	objectMeta.Labels[NDMHostKey] = di.HostName
	objectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	return objectMeta
}

/*
 * getTypeMeta returns TypeMeta struct which contains
 * resource kind and version. It is used to populate
 * data of Device struct of Device CR.
 */
func (di *DeviceInfo) getTypeMeta() metav1.TypeMeta {
	typeMeta := metav1.TypeMeta{
		Kind:       NDMDeviceKind,
		APIVersion: NDMVersion,
	}
	return typeMeta
}

/*
 * getClaimState returns claimState struct which contains
 * claim state of Device resource. It is used to populate
 * data of Device struct of Device CR.
 */
func (di *DeviceInfo) getClaimState() apis.DeviceClaimState {
	claimState := apis.DeviceClaimState{
		State: NDMUnclaimed,
	}
	return claimState
}

/*
 * getStatus returns DeviceStatus struct which contains
 * state of Device resource. It is used to populate data
 * of Device struct of Device CR.
 */
func (di *DeviceInfo) getStatus() apis.DeviceStatus {
	deviceStatus := apis.DeviceStatus{
		State: NDMActive,
	}
	return deviceStatus
}

/*
 * getDiskSpec returns DiskSpec struct which contains info of device like :
 * - path - /dev/sdb etc.
 * - capacity - (size,logical sector size ...)
 * - devlinks - (by-id , by-path links)
 * It is used to populate data of Device struct of device CR.
 */
func (di *DeviceInfo) getDeviceSpec() apis.DeviceSpec {
	deviceSpec := apis.DeviceSpec{}
	deviceSpec.Path = di.getPath()
	deviceSpec.Details = di.getDeviceDetails()
	deviceSpec.Capacity = di.getDeviceCapacity()
	deviceSpec.DevLinks = di.getDeviceLinks()
	deviceSpec.Partitioned = NDMNotPartitioned
	return deviceSpec
}

/*
 * getPath returns path of the device like (/dev/sda , /dev/sdb ...).
 * It is used to populate data of Device struct of Device CR.
 */
func (di *DeviceInfo) getPath() string {
	return di.Path
}

/*
 * getDeviceDetails returns DeviceDetails struct which contains primary
 * and static info of device resource like model, serial, vendor etc.
 * It is used to populate data of Device struct which of Device CR.
 */
func (di *DeviceInfo) getDeviceDetails() apis.DeviceDetails {
	deviceDetails := apis.DeviceDetails{}
	deviceDetails.Model = di.Model
	deviceDetails.Serial = di.Serial
	deviceDetails.Vendor = di.Vendor
	deviceDetails.FirmwareRevision = di.FirmwareRevision
	deviceDetails.Compliance = di.Compliance
	deviceDetails.DeviceType = di.DeviceType
	return deviceDetails
}

/*
 * getDiskCapacity returns DeviceCapacity struct which contains:
 * -size of disk (in bytes)
 * -logical sector size (in bytes)
 * -physical sector size (in bytes)
 * It is used to populate data of Device struct of Device CR.
 */
func (di *DeviceInfo) getDeviceCapacity() apis.DeviceCapacity {
	capacity := apis.DeviceCapacity{}
	capacity.Storage = di.Capacity
	capacity.LogicalSectorSize = di.LogicalSectorSize
	capacity.PhysicalSectorSize = di.PhysicalSectorSize
	return capacity
}

/*
 * getDiskLinks returns DeviceDevLink struct which contains
 * soft links like by-id ,by-path link. It is used to populate
 * data of Device struct of Device CR.
 */
func (di *DeviceInfo) getDeviceLinks() []apis.DeviceDevLink {
	devLinks := make([]apis.DeviceDevLink, 0)
	if len(di.ByIdDevLinks) != 0 {
		byIdLinks := apis.DeviceDevLink{
			Kind:  "by-id",
			Links: di.ByIdDevLinks,
		}
		devLinks = append(devLinks, byIdLinks)
	}
	if len(di.ByPathDevLinks) != 0 {
		byPathLinks := apis.DeviceDevLink{
			Kind:  "by-path",
			Links: di.ByPathDevLinks,
		}
		devLinks = append(devLinks, byPathLinks)
	}
	return devLinks
}
