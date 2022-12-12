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
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/util/jsonpath"
	"k8s.io/klog/v2"

	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
	bd "github.com/openebs/node-disk-manager/blockdevice"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeviceInfo contains details of blockdevice which can be converted into api.BlockDevice
// There will be one DeviceInfo created for each physical disk (may change in
// future). At the end it is converted to BlockDevice struct which will be pushed to
// etcd as a CR of that blockdevice.
type DeviceInfo struct {
	// NodeAttributes is the attributes of the node to which this block device is attached,
	// like hostname, nodename
	NodeAttributes bd.NodeAttribute
	// Optional labels that can be added to the blockdevice resource
	Labels             map[string]string
	UUID               string   // UUID of backing disk
	Capacity           uint64   // Capacity of blockdevice
	Model              string   // Do blockdevice have model ??
	Serial             string   // Do blockdevice have serial no ??
	Vendor             string   // Vendor of blockdevice
	Path               string   // blockdevice Path like /dev/sda
	ByIdDevLinks       []string // ByIdDevLinks contains by-id devlinks
	ByPathDevLinks     []string // ByPathDevLinks contains by-path devlinks
	FirmwareRevision   string   // FirmwareRevision is the firmware revision for a disk
	LogicalBlockSize   uint32   // LogicalBlockSize is the logical block size of the device in bytes
	PhysicalBlockSize  uint32   // PhysicalBlockSize is the physical block size in bytes
	HardwareSectorSize uint32   // HardwareSectorSize is the hardware sector size in bytes
	Compliance         string   // Compliance is implemented specifications version i.e. SPC-1, SPC-2, etc
	DeviceType         string   // DeviceType represents the type of device, like disk/sparse/partition
	DriveType          string   // DriveType represents the type of backing drive HDD/SSD
	PartitionType      string   // Partition type if the blockdevice is a partition
	FileSystemInfo     FSInfo   // FileSystem info of the blockdevice like FSType and MountPoint
}

// NewDeviceInfo returns a pointer of empty DeviceInfo
// struct which will be field from DiskInfo.
func NewDeviceInfo() *DeviceInfo {
	deviceInfo := &DeviceInfo{}
	return deviceInfo
}

// FSInfo defines the filesystem related information of block device/disk, like mountpoint and
// filesystem
type FSInfo struct {
	FileSystem string // Filesystem on the block device
	MountPoint string // MountPoint of the block device
}

// ToDevice convert deviceInfo struct to api.BlockDevice
// type which will be pushed to etcd
func (di *DeviceInfo) ToDevice(controller *Controller) (apis.BlockDevice, error) {
	blockDevice := apis.BlockDevice{}
	blockDevice.Spec = di.getDeviceSpec()
	blockDevice.ObjectMeta = di.getObjectMeta()
	blockDevice.TypeMeta = di.getTypeMeta()
	blockDevice.Status = di.getStatus()
	err := addBdLabels(&blockDevice, controller)
	if err != nil {
		return blockDevice, fmt.Errorf("error in adding labels to the blockdevice: %v", err)
	}
	return blockDevice, nil
}

// addBdLabels add labels to block device that may be helpful for filtering the block device
// based on some generic attributes like drive-type, model, vendor etc.
func addBdLabels(bd *apis.BlockDevice, ctrl *Controller) error {
	var JsonPathFields []string

	// get the labels to be added from the configmap
	if ctrl.NDMConfig != nil {
		for _, metaConfig := range ctrl.NDMConfig.MetaConfigs {
			if metaConfig.Key == deviceLabelsKey {
				JsonPathFields = strings.Split(metaConfig.Type, ",")
			}
		}
	}

	if len(JsonPathFields) > 0 {
		for _, jsonPath := range JsonPathFields {
			// Parse jsonpath
			fields, err := RelaxedJSONPathExpression(strings.TrimSpace(jsonPath))
			if err != nil {
				klog.Errorf("Error converting into a valid jsonpath expression: %+v", err)
				return err
			}

			j := jsonpath.New(jsonPath)
			if err := j.Parse(fields); err != nil {
				klog.Errorf("Error parsing jsonpath: %s, error: %+v", fields, err)
				return err
			}

			values, err := j.FindResults(bd)
			if err != nil {
				klog.Errorf("Error finding results for jsonpath: %s, error: %+v", fields, err)
				return err
			}

			valueStrings := []string{}
			var jsonPathFieldValue string

			if len(values) > 0 && len(values[0]) > 0 {
				for arrIx := range values {
					for valIx := range values[arrIx] {
						valueStrings = append(valueStrings, fmt.Sprintf("%v", values[arrIx][valIx].Interface()))
					}
				}

				// convert the string array into a single string
				jsonPathFieldValue = strings.Join(valueStrings, ",")
				jsonPathFieldValue = strings.TrimSuffix(jsonPathFieldValue, ",")

				// the above string formed should adhere to k8s label rules inorder for it to be
				// used as a label value for blockdevice object.
				// Check this link for more info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
				errs := validation.IsValidLabelValue(jsonPathFieldValue)
				if len(errs) > 0 {
					return fmt.Errorf("invalid Label found. Error: {%s}", strings.Join(errs, ","))
				}

				jsonPathFields := strings.Split(jsonPath, ".")
				if len(jsonPathFields) > 0 && jsonPathFields[len(jsonPathFields)-1] != "" && jsonPathFieldValue != "" {
					bd.Labels[NDMLabelPrefix+jsonPathFields[len(jsonPathFields)-1]] = jsonPathFieldValue
				}
			}
		}
		klog.V(4).Infof("successfully added device labels")
	}
	return nil
}

// RelaxedJSONPathExpression attempts to be flexible with JSONPath expressions, it accepts:
//   * metadata.name (no leading '.' or curly braces '{...}'
//   * {metadata.name} (no leading '.')
//   * .metadata.name (no curly braces '{...}')
//   * {.metadata.name} (complete expression)
// And transforms them all into a valid jsonpath expression:
//   {.metadata.name}
// NOTE: This code has been referenced from kubernetes kubectl github repo.
//       Ref: https://github.com/kubernetes/kubectl/blob/caeb9274868c57d8a320014290cc7e3d1bcb9e46/pkg/cmd/get
//      /customcolumn.go#L47
func RelaxedJSONPathExpression(pathExpression string) (string, error) {
	var jsonRegexp = regexp.MustCompile(`^\{\.?([^{}]+)\}$|^\.?([^{}]+)$`)

	if len(pathExpression) == 0 {
		return pathExpression, nil
	}
	submatches := jsonRegexp.FindStringSubmatch(pathExpression)
	if submatches == nil {
		return "", fmt.Errorf("unexpected path string, expected a 'name1.name2' or '.name1.name2' or '{name1.name2}' or '{.name1.name2}'")
	}
	if len(submatches) != 3 {
		return "", fmt.Errorf("unexpected submatch list: %v", submatches)
	}
	var fieldSpec string
	if len(submatches[1]) != 0 {
		fieldSpec = submatches[1]
	} else {
		fieldSpec = submatches[2]
	}
	return fmt.Sprintf("{.%s}", fieldSpec), nil
}

// getObjectMeta returns ObjectMeta struct which contains
// labels and Name of resource. It is used to populate data
// of BlockDevice struct of BlockDevice CR.
func (di *DeviceInfo) getObjectMeta() metav1.ObjectMeta {
	objectMeta := metav1.ObjectMeta{
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
		Name:        di.UUID,
	}
	//objectMeta.Labels[KubernetesHostNameLabel] = di.NodeAttributes[HostNameKey]
	for k, v := range di.NodeAttributes {
		if k == HostNameKey {
			objectMeta.Labels[KubernetesHostNameLabel] = v
		} else {
			objectMeta.Labels[k] = v
		}
	}
	objectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	objectMeta.Labels[NDMManagedKey] = TrueString
	// adding custom labels
	for k, v := range di.Labels {
		objectMeta.Labels[k] = v
	}
	return objectMeta
}

// getTypeMeta returns TypeMeta struct which contains
// resource kind and version. It is used to populate
// data of BlockDevice struct of BlockDevice CR.
func (di *DeviceInfo) getTypeMeta() metav1.TypeMeta {
	typeMeta := metav1.TypeMeta{
		Kind:       NDMBlockDeviceKind,
		APIVersion: NDMVersion,
	}
	return typeMeta
}

// getStatus returns DeviceStatus struct which contains
// state of BlockDevice resource. It is used to populate data
// of BlockDevice struct of BlockDevice CR.
func (di *DeviceInfo) getStatus() apis.DeviceStatus {
	deviceStatus := apis.DeviceStatus{
		ClaimState: apis.BlockDeviceUnclaimed,
		State:      NDMActive,
	}
	return deviceStatus
}

// getDiskSpec returns DiskSpec struct which contains info of blockdevice like :
// - path - /dev/sdb etc.
// - capacity - (size,logical sector size ...)
// - devlinks - (by-id , by-path links)
// It is used to populate data of BlockDevice struct of blockdevice CR.
func (di *DeviceInfo) getDeviceSpec() apis.DeviceSpec {
	deviceSpec := apis.DeviceSpec{}
	deviceSpec.NodeAttributes.NodeName = di.NodeAttributes[NodeNameKey]
	deviceSpec.Path = di.getPath()
	deviceSpec.Details = di.getDeviceDetails()
	deviceSpec.Capacity = di.getDeviceCapacity()
	deviceSpec.DevLinks = di.getDeviceLinks()
	deviceSpec.Partitioned = NDMNotPartitioned
	deviceSpec.FileSystem = di.FileSystemInfo.getFileSystemInfo()
	return deviceSpec
}

// getPath returns path of the blockdevice like (/dev/sda , /dev/sdb ...).
// It is used to populate data of BlockDevice struct of BlockDevice CR.
func (di *DeviceInfo) getPath() string {
	return di.Path
}

// getDeviceDetails returns DeviceDetails struct which contains primary
// and static info of blockdevice resource like model, serial, vendor etc.
// It is used to populate data of BlockDevice struct which of BlockDevice CR.
func (di *DeviceInfo) getDeviceDetails() apis.DeviceDetails {
	deviceDetails := apis.DeviceDetails{}
	deviceDetails.Model = di.Model
	deviceDetails.Serial = di.Serial
	deviceDetails.Vendor = di.Vendor
	deviceDetails.FirmwareRevision = di.FirmwareRevision
	deviceDetails.Compliance = di.Compliance
	deviceDetails.DeviceType = di.DeviceType
	deviceDetails.DriveType = di.DriveType
	deviceDetails.LogicalBlockSize = di.LogicalBlockSize
	deviceDetails.PhysicalBlockSize = di.PhysicalBlockSize
	deviceDetails.HardwareSectorSize = di.HardwareSectorSize

	return deviceDetails
}

// getDiskCapacity returns DeviceCapacity struct which contains:
// -size of disk (in bytes)
// -logical sector size (in bytes)
// -physical sector size (in bytes)
// It is used to populate data of BlockDevice struct of BlockDevice CR.
func (di *DeviceInfo) getDeviceCapacity() apis.DeviceCapacity {
	capacity := apis.DeviceCapacity{}
	capacity.Storage = di.Capacity
	capacity.LogicalSectorSize = di.LogicalBlockSize
	capacity.PhysicalSectorSize = di.PhysicalBlockSize
	return capacity
}

// getDiskLinks returns DeviceDevLink struct which contains
// soft links like by-id ,by-path link. It is used to populate
// data of BlockDevice struct of BlockDevice CR.
func (di *DeviceInfo) getDeviceLinks() []apis.DeviceDevLink {
	devLinks := make([]apis.DeviceDevLink, 0)
	if len(di.ByIdDevLinks) != 0 {
		byIDLinks := apis.DeviceDevLink{
			Kind:  "by-id",
			Links: di.ByIdDevLinks,
		}
		devLinks = append(devLinks, byIDLinks)
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

func (fs *FSInfo) getFileSystemInfo() apis.FileSystemInfo {
	fsInfo := apis.FileSystemInfo{}
	fsInfo.Type = fs.FileSystem
	fsInfo.Mountpoint = fs.MountPoint
	return fsInfo
}
