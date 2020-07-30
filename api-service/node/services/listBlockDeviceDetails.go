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

	"github.com/openebs/node-disk-manager/pkg/smart"
	protos "github.com/openebs/node-disk-manager/spec/ndm"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

// ListBlockDeviceDetails gives the details about the disk from SMART
func (n *Node) ListBlockDeviceDetails(ctx context.Context, bd *protos.BlockDevice) (*protos.BlockDeviceDetails, error) {

	klog.Info("Listing Device Details")

	device := smart.Identifier{
		DevPath: bd.Name,
	}
	info, err := device.SCSIBasicDiskInfo()
	if len(err) != 0 {
		klog.Errorf("Error fetching block device details %v", err)
		return nil, status.Errorf(codes.Internal, "Error fetching disk details")
	}
	klog.V(4).Info(info.BasicDiskAttr)
	klog.V(4).Info(info.ATADiskAttr)

	return &protos.BlockDeviceDetails{
		Compliance:       info.BasicDiskAttr.Compliance,
		Vendor:           info.BasicDiskAttr.Vendor,
		Model:            info.BasicDiskAttr.ModelNumber,
		SerialNumber:     info.BasicDiskAttr.SerialNumber,
		FirmwareRevision: info.BasicDiskAttr.FirmwareRevision,
		WWN:              info.BasicDiskAttr.WWN,
		Capacity:         info.BasicDiskAttr.Capacity,
		LBSize:           info.BasicDiskAttr.LBSize,
		PBSize:           info.BasicDiskAttr.PBSize,
		RotationRate:     uint32(info.BasicDiskAttr.RotationRate),
		ATAMajorVersion:  info.ATADiskAttr.ATAMajorVersion,
		ATAMinorVersion:  info.ATADiskAttr.ATAMinorVersion,
		AtaTransport:     info.ATADiskAttr.AtaTransport,
	}, nil

}
