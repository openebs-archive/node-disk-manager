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

package services

import (
	"context"
	"strings"

	"github.com/openebs/node-disk-manager/api-service/node"
	"github.com/openebs/node-disk-manager/api/v1alpha1"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/sysfs"
	"github.com/openebs/node-disk-manager/pkg/util"
	protos "github.com/openebs/node-disk-manager/spec/ndm"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

// Node helps in using types defined in package Node
type Node struct {
	node.Node
}

// NewNode returns an instance of type Node
func NewNode() *Node {
	return &Node{}
}

type disks []protos.BlockDevice

// AllBlockDevices contains all the relationships and device types
type AllBlockDevices struct {
	Parents    []string
	Partitions []string
	LVMs       []string
	RAIDs      []string
	Holders    []string
	Slaves     []string
	Loops      []string
	Sparse     []string
}

var all AllBlockDevices // This variable would contain all devices found in the node and their relationships

// ConfigFilePath refers to the config file for ndm
const ConfigFilePath = "/host/node-disk-manager.config"

// ListBlockDevices returns the block devices and their relationships
func (n *Node) ListBlockDevices(ctx context.Context, null *protos.Null) (*protos.BlockDevices, error) {
	klog.Info("Listing block devices")

	ctrl, err := controller.NewController()
	if err != nil {
		klog.Errorf("Error creating a controller %v", err)
		return nil, status.Errorf(codes.NotFound, "Namespace not found")
	}

	err = ctrl.SetControllerOptions(controller.NDMOptions{ConfigFilePath: ConfigFilePath})
	if err != nil {
		klog.Errorf("Error setting config to controller %v", err)
		return nil, status.Errorf(codes.Internal, "Error setting config to controller")
	}

	blockDeviceList, err := ctrl.ListBlockDeviceResource(false)
	if err != nil {
		klog.Errorf("Error listing block devices %v", err)
		return nil, status.Errorf(codes.Internal, "Error fetching list of disks")
	}

	if len(blockDeviceList.Items) == 0 {
		klog.V(4).Info("No items found")
	}

	blockDevices := make([]*protos.BlockDevice, 0)

	err = GetAllTypes(blockDeviceList)
	if err != nil {
		klog.Errorf("Error fetching Parent disks %v", err)
	}

	for _, name := range all.Parents {

		blockDevices = append(blockDevices, &protos.BlockDevice{
			Name:       name,
			Type:       "Disk",
			Partitions: FilterPartitions(name, all.Partitions),
		})
	}

	for _, name := range all.LVMs {

		blockDevices = append(blockDevices, &protos.BlockDevice{
			Name: name,
			Type: "LVM",
		})
	}

	for _, name := range all.RAIDs {

		blockDevices = append(blockDevices, &protos.BlockDevice{
			Name: name,
			Type: "RAID",
		})
	}

	for _, name := range all.Loops {

		blockDevices = append(blockDevices, &protos.BlockDevice{
			Name:       name,
			Type:       "Loop",
			Partitions: FilterPartitions(name, all.Partitions),
		})
	}

	for _, name := range all.Sparse {

		blockDevices = append(blockDevices, &protos.BlockDevice{
			Name: name,
			Type: "Sparse",
		})
	}

	return &protos.BlockDevices{
		Blockdevices: blockDevices,
	}, nil
}

// GetAllTypes updates the list of all block devices found on nodes and their relationships
func GetAllTypes(BL *v1alpha1.BlockDeviceList) error {
	ParentDeviceNames := make([]string, 0)
	HolderDeviceNames := make([]string, 0)
	SlaveDeviceNames := make([]string, 0)
	PartitionNames := make([]string, 0)
	LoopNames := make([]string, 0)
	SparseNames := make([]string, 0)
	LVMNames := make([]string, 0)
	RAIDNames := make([]string, 0)

	for _, bd := range BL.Items {
		klog.V(4).Infof("Device %v of type %v ", bd.Spec.Path, bd.Spec.Details.DeviceType)

		if bd.Spec.Details.DeviceType == "sparse" {
			SparseNames = append(SparseNames, bd.Spec.Path)
			continue
		}

		// GetDependents should not be called on sparse devices, hence this block comes later.
		sysfsDevice, err := sysfs.NewSysFsDeviceFromDevPath(bd.Spec.Path)
		if err != nil {
			klog.Errorf("could not get sysfs device for %s, err: %v", bd.Spec.Path, err)
			continue
		}
		depDevices, err := sysfsDevice.GetDependents()
		if err != nil {
			klog.Errorf("Error fetching dependents of the disk name: %v, err: %v", bd.Spec.Path, err)
			continue
		}

		if bd.Spec.Details.DeviceType == "loop" {
			LoopNames = append(LoopNames, bd.Spec.Path)
			PartitionNames = append(PartitionNames, depDevices.Partitions...)
			continue
		}

		// This will run when GPTbasedUUID is enabled
		if bd.Spec.Details.DeviceType == "partition" {
			// We add the partition only if it doesn't already exist
			PartitionNames = util.AddUniqueStringtoSlice(PartitionNames, bd.Spec.Path)
			// We add the parent if it doesn't already exist
			ParentDeviceNames = util.AddUniqueStringtoSlice(ParentDeviceNames, depDevices.Parent)
			// Since partitions can also be holders
			HolderDeviceNames = append(HolderDeviceNames, depDevices.Holders...)

			// Since we found it's a partition and got it's dependents, we can continue with next device

			continue
		}

		// This will run when GPTbasedUUID is disabled
		if bd.Spec.Details.DeviceType == "disk" {
			// We add the parent if it doesn't exist
			ParentDeviceNames = util.AddUniqueStringtoSlice(ParentDeviceNames, bd.Spec.Path)
			// Since there isn't a way we come across this BD again, we can add partitions all at once without checking they exist already
			PartitionNames = append(PartitionNames, depDevices.Partitions...)
			continue
		}

		if bd.Spec.Details.DeviceType == "lvm" {
			// Add the lvm if it doesn't already exist
			LVMNames = util.AddUniqueStringtoSlice(LVMNames, bd.Spec.Path)
			// if we encounter a lvm say dm-0, we add it's slaves(sda1, sdb1)
			SlaveDeviceNames = append(SlaveDeviceNames, depDevices.Slaves...)
			// if we encounter a lvm say dm-1 which is a partition of dm-0, then dm-0 would be a holder of dm-1
			HolderDeviceNames = append(HolderDeviceNames, depDevices.Holders...)
			continue
		}

		if strings.Contains(bd.Spec.Details.DeviceType, "raid") {
			// Add the RAID if it doesn't already exist
			RAIDNames = util.AddUniqueStringtoSlice(RAIDNames, bd.Spec.Path)
			// if we encounter a RAID device md-0, we add it's slaves(sda1, sdb1)
			SlaveDeviceNames = append(SlaveDeviceNames, depDevices.Slaves...)
			// if we encounter a raid say md-1 which is a partition of md-0, then md-0 would be a holder of md-1
			HolderDeviceNames = append(HolderDeviceNames, depDevices.Holders...)
			continue
		}

	}
	klog.V(4).Infof("Parent Devices found are: %v", ParentDeviceNames)
	klog.V(4).Infof("Partitions  found are: %v", PartitionNames)

	klog.V(4).Infof("LVM found are: %v", LVMNames)
	klog.V(4).Infof("RAID disks found are: %v", RAIDNames)

	klog.V(4).Infof("Holder Devices found are: %v", HolderDeviceNames)
	klog.V(4).Infof("Slaves found are: %v", SlaveDeviceNames)

	klog.V(4).Infof("Loop Devices found are: %v", LoopNames)
	klog.V(4).Infof("Sparse disks found are: %v", SparseNames)

	all.Parents = ParentDeviceNames
	all.Partitions = PartitionNames
	all.Holders = HolderDeviceNames
	all.Slaves = SlaveDeviceNames
	all.LVMs = LVMNames
	all.RAIDs = RAIDNames
	all.Loops = LoopNames
	all.Sparse = SparseNames
	return nil

}

// FilterPartitions gets the name of the partitions given a block device.
// Given a disk name /dev/sdb and slice of partition names : ["/dev/sdb1", "/dev/sdb2", "/dev/sdc1"],
//it should return ["/dev/sdb1", "/dev/sdb2"]
func FilterPartitions(name string, pns []string) []string {
	fpns := make([]string, 0)

	if len(pns) == 0 {
		pns = append(pns, name)
		return pns
	}

	for _, pn := range pns {
		if strings.Contains(pn, name) {
			fpns = append(fpns, pn)
		}
	}

	return fpns
}
