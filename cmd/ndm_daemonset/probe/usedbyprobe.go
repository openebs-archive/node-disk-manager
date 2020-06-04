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
	"os"
	"strings"
	"syscall"

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/blkid"
	"github.com/openebs/node-disk-manager/pkg/spdk"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/openebs/node-disk-manager/pkg/util"

	"k8s.io/klog"
)

type usedbyProbe struct {
	Controller      *controller.Controller
	BlkidIdentifier *blkid.DeviceIdentifier
}

const (
	usedbyProbeConfigKey = "used-by-probe"
	usedbyProbePriority  = 4

	k8sLocalVolumePath1 = "kubernetes.io/local-volume"
	k8sLocalVolumePath2 = "kubernetes.io~local-volume"
	zfsFileSystemLabel  = "zfs_member"
)

var (
	usedbyProbeName  = "used-by probe"
	usedbyProbeState = defaultEnabled
)

var usedbyProbeRegister = func() {
	// Get a controller object
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		klog.Error("unable to configure", usedbyProbeName)
		return
	}
	if ctrl.NDMConfig != nil {
		for _, probeConfig := range ctrl.NDMConfig.ProbeConfigs {
			if probeConfig.Key == usedbyProbeConfigKey {
				usedbyProbeName = probeConfig.Name
				usedbyProbeState = util.CheckTruthy(probeConfig.State)
				break
			}
		}
	}
	newRegisterProbe := &registerProbe{
		priority:   usedbyProbePriority,
		name:       usedbyProbeName,
		state:      usedbyProbeState,
		pi:         &usedbyProbe{Controller: ctrl},
		controller: ctrl,
	}
	// Here we register the used-by probe
	newRegisterProbe.register()
}

func newUsedByProbe(devPath string) *usedbyProbe {
	usedbyProbe := &usedbyProbe{
		BlkidIdentifier: &blkid.DeviceIdentifier{
			DevPath: devPath,
		},
	}
	return usedbyProbe
}

func (sp *usedbyProbe) Start() {}

func (sp *usedbyProbe) FillBlockDeviceDetails(blockDevice *blockdevice.BlockDevice) {
	if blockDevice.DevPath == "" {
		klog.Errorf("device identifier found empty, used-by probe will not fetch information")
		return
	}

	// checking for local PV on the device
	for _, mountPoint := range blockDevice.FSInfo.MountPoint {
		if strings.Contains(mountPoint, k8sLocalVolumePath1) ||
			strings.Contains(mountPoint, k8sLocalVolumePath2) {
			blockDevice.DevUse.InUse = true
			blockDevice.DevUse.UsedBy = blockdevice.LocalPV
			klog.V(4).Infof("device: %s Used by: %s filled by used-by probe", blockDevice.DevPath, blockDevice.DevUse.UsedBy)
			return
		}
	}

	// checking for cstor and zfs localPV
	// we start with the assumption that device has a zfs file system
	lookupZFS := true
	devPath := blockDevice.DevPath

	// if device is not a partition, check whether the zfs partitions are available on the disk
	if blockDevice.DeviceAttributes.DeviceType != libudevwrapper.UDEV_PARTITION {
		path, ok := getBlockDeviceZFSPartition(*blockDevice)
		if ok {
			devPath = path
		} else {
			lookupZFS = false
			klog.V(4).Infof("device: %s is not having any zfs partitions", blockDevice.DevPath)
		}
	}

	// only if lookupZFS is true, we need to check for the zfs filesystem on the disk.
	if lookupZFS {
		usedByProbe := newUsedByProbe(devPath)

		// check for ZFS file system
		fstype := usedByProbe.BlkidIdentifier.GetOnDiskFileSystem()

		if fstype == zfsFileSystemLabel {
			blockDevice.DevUse.InUse = true

			// the disk can either be in use by cstor or zfs local PV
			ok, err := isBlockDeviceInUseByKernel(blockDevice.DevPath)

			if err != nil {
				klog.Errorf("error checking block device: %s: %v", blockDevice.DevPath, err)
			}
			if ok {
				blockDevice.DevUse.UsedBy = blockdevice.ZFSLocalPV
			} else {
				blockDevice.DevUse.UsedBy = blockdevice.CStor
			}
			klog.V(4).Infof("device: %s Used by: %s filled by sampling probe", blockDevice.DevPath, blockDevice.DevUse.UsedBy)
			return
		}
	}

	// create a device identifier for reading the spdk super block from the disk
	spdkIdentifier := &spdk.DeviceIdentifier{
		DevPath: blockDevice.DevPath,
	}

	signature, err := spdkIdentifier.GetSPDKSuperBlockSignature()
	if err != nil {
		klog.Errorf("error reading spdk signature from device: %s, %v", blockDevice.DevPath, err)
	}
	if spdk.IsSPDKSignatureExist(signature) {
		blockDevice.DevUse.InUse = true
		blockDevice.DevUse.UsedBy = blockdevice.Mayastor
		klog.V(4).Infof("device: %s Used by: %s filled by sampling probe", blockDevice.DevPath, blockDevice.DevUse.UsedBy)
		return
	}

	// TODO jiva disk detection
}

// getBlockDeviceZFSPartition is used to get the zfs partition if it exist in a
// given BD
func getBlockDeviceZFSPartition(bd blockdevice.BlockDevice) (string, bool) {

	// check for zfs partitions only if there are 2 partitions on the block device
	if len(bd.DependentDevices.Partitions) != 2 {
		return "", false
	}

	zfsDataPartitionNumber := "1"
	zfsMetaPartitionNumber := "9"

	// to handle cases of devices with nvme drives, the device name will be
	// nvme0n1
	if util.IsMatchRegex(".+[0-9]+$", bd.DevPath) {
		zfsDataPartitionNumber = "p" + zfsDataPartitionNumber
		zfsMetaPartitionNumber = "p" + zfsMetaPartitionNumber
	}

	dataPartition := bd.DevPath + zfsDataPartitionNumber
	metaPartition := bd.DevPath + zfsMetaPartitionNumber

	// check if device has the 2 partitions
	if bd.DependentDevices.Partitions[0] == dataPartition &&
		bd.DependentDevices.Partitions[1] == metaPartition {
		return dataPartition, true
	}
	return "", false
}

// isBlockDeviceInUseByKernel tries to open the device exclusively to check if the device is
// being held by some process. eg: If kernel zfs uses the disk, the open will fail
func isBlockDeviceInUseByKernel(path string) (bool, error) {
	_, err := os.OpenFile(path, os.O_EXCL, 0444)

	if errors.Is(err, syscall.EBUSY) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}
