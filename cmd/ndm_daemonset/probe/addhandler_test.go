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
	"context"
	"testing"

	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/db/kubernetes"
	"github.com/openebs/node-disk-manager/pkg/util"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestAddBlockDeviceToHierarchyCache(t *testing.T) {
	tests := map[string]struct {
		cache     blockdevice.Hierarchy
		bd        blockdevice.BlockDevice
		wantCache blockdevice.Hierarchy
		wantOk    bool
	}{
		"empty cache": {
			cache: make(blockdevice.Hierarchy),
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
			},
			wantCache: map[string]blockdevice.BlockDevice{
				"/dev/sda": {
					Identifier: blockdevice.Identifier{
						DevPath: "/dev/sda",
					},
				},
			},
			wantOk: false,
		},
		"cache with same device already existing": {
			cache: map[string]blockdevice.BlockDevice{
				"/dev/sda": {
					Identifier: blockdevice.Identifier{
						DevPath: "/dev/sda",
					},
				},
			},
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
					SysPath: "/sys/class/block/sda",
				},
			},
			wantCache: map[string]blockdevice.BlockDevice{
				"/dev/sda": {
					Identifier: blockdevice.Identifier{
						SysPath: "/sys/class/block/sda",
						DevPath: "/dev/sda",
					},
				},
			},
			wantOk: true,
		},
		"cache with different device existing": {
			cache: map[string]blockdevice.BlockDevice{
				"/dev/sda": {
					Identifier: blockdevice.Identifier{
						DevPath: "/dev/sda",
					},
				},
			},
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sdb",
				},
			},
			wantCache: map[string]blockdevice.BlockDevice{
				"/dev/sda": {
					Identifier: blockdevice.Identifier{
						DevPath: "/dev/sda",
					},
				},
				"/dev/sdb": {
					Identifier: blockdevice.Identifier{
						DevPath: "/dev/sdb",
					},
				},
			},
			wantOk: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			pe := &ProbeEvent{
				Controller: &controller.Controller{
					BDHierarchy: tt.cache,
				},
			}
			gotOk := pe.addBlockDeviceToHierarchyCache(tt.bd)
			assert.Equal(t, tt.wantCache, pe.Controller.BDHierarchy)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}

func TestDeviceInUseByMayastor(t *testing.T) {
	tests := map[string]struct {
		bd        blockdevice.BlockDevice
		bdAPIList *apis.BlockDeviceList
		want      bool
		wantErr   bool
	}{
		"device not in use": {
			bd: blockdevice.BlockDevice{
				DevUse: blockdevice.DeviceUsage{
					InUse: false,
				},
			},
			want:    true,
			wantErr: false,
		},
		"device in use, but not by mayastor": {
			bd: blockdevice.BlockDevice{
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.LocalPV,
				},
			},
			want:    true,
			wantErr: false,
		},
		"device in use by mayastor": {
			bd: blockdevice.BlockDevice{
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.Mayastor,
				},
			},
			want:    false,
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			pe := &ProbeEvent{}
			got, err := pe.deviceInUseByMayastor(tt.bd, tt.bdAPIList)
			if (err != nil) != tt.wantErr {
				t.Errorf("deviceInUseByMayastor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeviceInUseByZFSLocalPV(t *testing.T) {
	fakePartTableID := "fake-part-table-uuid"
	fakeBD := blockdevice.BlockDevice{
		PartitionInfo: blockdevice.PartitionInformation{
			PartitionTableUUID: fakePartTableID,
		},
	}
	fakeUUID, _ := generateUUIDFromPartitionTable(fakeBD)

	tests := map[string]struct {
		bd                     blockdevice.BlockDevice
		bdAPIList              *apis.BlockDeviceList
		bdCache                blockdevice.Hierarchy
		createdOrUpdatedBDName string
		want                   bool
		wantErr                bool
	}{
		"device not in use": {
			bd: blockdevice.BlockDevice{
				DevUse: blockdevice.DeviceUsage{
					InUse: false,
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
			},
			bdAPIList:              &apis.BlockDeviceList{},
			bdCache:                nil,
			createdOrUpdatedBDName: "",
			want:                   true,
			wantErr:                false,
		},
		"device in use, not by zfs localPV": {
			bd: blockdevice.BlockDevice{
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.CStor,
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
			},
			bdAPIList:              &apis.BlockDeviceList{},
			bdCache:                nil,
			createdOrUpdatedBDName: "",
			want:                   true,
			wantErr:                false,
		},
		"deviceType partition, parent device used by zfs localPV": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.ZFSLocalPV,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent: "/dev/sda",
				},
			},
			bdAPIList: &apis.BlockDeviceList{},
			bdCache: blockdevice.Hierarchy{
				"/dev/sda": {
					DevUse: blockdevice.DeviceUsage{
						InUse:  true,
						UsedBy: blockdevice.ZFSLocalPV,
					},
				},
			},
			createdOrUpdatedBDName: "",
			want:                   false,
			wantErr:                false,
		},
		"deviceType partition, parent device used by cstor": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent: "/dev/sda",
				},
			},
			bdAPIList: &apis.BlockDeviceList{},
			bdCache: blockdevice.Hierarchy{
				"/dev/sda": {
					DevUse: blockdevice.DeviceUsage{
						InUse:  true,
						UsedBy: blockdevice.CStor,
					},
				},
			},
			createdOrUpdatedBDName: "",
			want:                   true,
			wantErr:                false,
		},
		// if multiple partitions are there, this test may need to be revisited
		"deviceType partition, parent device not in use": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent: "/dev/sda",
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.ZFSLocalPV,
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: fakePartTableID,
				},
			},
			bdAPIList: &apis.BlockDeviceList{},
			bdCache: blockdevice.Hierarchy{
				"/dev/sda": {
					DevUse: blockdevice.DeviceUsage{
						InUse: false,
					},
				},
			},
			createdOrUpdatedBDName: fakeUUID,
			want:                   false,
			wantErr:                false,
		},
		"deviceType disk, used by zfs PV and is connected to the cluster for the first time": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.ZFSLocalPV,
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: fakePartTableID,
				},
			},
			bdAPIList:              &apis.BlockDeviceList{},
			bdCache:                nil,
			createdOrUpdatedBDName: fakeUUID,
			want:                   false,
			wantErr:                false,
		},
		"deviceType disk, used by zfs PV and is moved from disconnected and reconnected to the node at a different path": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.ZFSLocalPV,
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: fakePartTableID,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: fakeUUID,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdb",
						},
					},
				},
			},
			bdCache:                nil,
			createdOrUpdatedBDName: fakeUUID,
			want:                   false,
			wantErr:                false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := scheme.Scheme
			s.AddKnownTypes(apis.GroupVersion, &apis.BlockDevice{})
			s.AddKnownTypes(apis.GroupVersion, &apis.BlockDeviceList{})
			cl := fake.NewFakeClientWithScheme(s)

			// initialize client with all the bd resources
			for _, bdAPI := range tt.bdAPIList.Items {
				cl.Create(context.TODO(), &bdAPI)
			}

			ctrl := &controller.Controller{
				Clientset:   cl,
				BDHierarchy: tt.bdCache,
			}
			pe := &ProbeEvent{
				Controller: ctrl,
			}
			got, err := pe.deviceInUseByZFSLocalPV(tt.bd, tt.bdAPIList)
			if (err != nil) != tt.wantErr {
				t.Errorf("deviceInUseByZFSLocalPV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)

			// check if a BD has been created or updated
			if len(tt.createdOrUpdatedBDName) != 0 {
				gotBDAPI := &apis.BlockDevice{}
				err := cl.Get(context.TODO(), client.ObjectKey{Name: tt.createdOrUpdatedBDName}, gotBDAPI)
				if err != nil {
					t.Errorf("error in getting blockdevice %s", tt.createdOrUpdatedBDName)
				}
				// verify the block-device-tag on the resource, also verify the path and node name
				assert.Equal(t, string(blockdevice.ZFSLocalPV), gotBDAPI.GetLabels()[kubernetes.BlockDeviceTagLabel])
				assert.Equal(t, tt.bd.DevPath, gotBDAPI.Spec.Path)
				assert.Equal(t, tt.bd.NodeAttributes[blockdevice.NodeName], gotBDAPI.Spec.NodeAttributes.NodeName)
			}
		})
	}
}

func TestIsParentDeviceInUse(t *testing.T) {
	cache := map[string]blockdevice.BlockDevice{
		"/dev/sda": {
			Identifier: blockdevice.Identifier{
				DevPath: "/dev/sda",
			},
			DependentDevices: blockdevice.DependentBlockDevices{
				Parent:     "",
				Partitions: []string{"/dev/sda1", "/dev/sda2"},
			},
			DeviceAttributes: blockdevice.DeviceAttribute{
				DeviceType: blockdevice.BlockDeviceTypeDisk,
			},
			DevUse: blockdevice.DeviceUsage{
				InUse: false,
			},
		},
		"/dev/sda1": {
			Identifier: blockdevice.Identifier{
				DevPath: "/dev/sda1",
			},
			DependentDevices: blockdevice.DependentBlockDevices{
				Parent: "/dev/sda",
			},
			DeviceAttributes: blockdevice.DeviceAttribute{
				DeviceType: blockdevice.BlockDeviceTypePartition,
			},
			DevUse: blockdevice.DeviceUsage{
				InUse: true,
			},
		},
		"/dev/sda2": {
			Identifier: blockdevice.Identifier{
				DevPath: "/dev/sda2",
			},
			DependentDevices: blockdevice.DependentBlockDevices{
				Parent: "/dev/sda",
			},
			DeviceAttributes: blockdevice.DeviceAttribute{
				DeviceType: blockdevice.BlockDeviceTypePartition,
			},
			DevUse: blockdevice.DeviceUsage{
				InUse: false,
			},
		},
		"/dev/sdb": {
			Identifier: blockdevice.Identifier{
				DevPath: "/dev/sdb",
			},
			DependentDevices: blockdevice.DependentBlockDevices{
				Parent: "",
			},
			DeviceAttributes: blockdevice.DeviceAttribute{
				DeviceType: blockdevice.BlockDeviceTypeDisk,
			},
			DevUse: blockdevice.DeviceUsage{
				InUse: true,
			},
		},
	}
	pe := &ProbeEvent{
		Controller: &controller.Controller{
			BDHierarchy: cache,
		},
	}
	tests := map[string]struct {
		bd      blockdevice.BlockDevice
		want    bool
		wantErr bool
	}{
		"check for existing parent device": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
			},
			want:    false,
			wantErr: false,
		},
		"check for partition that is in use": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent: "/dev/sda",
				},
			},
			want:    false,
			wantErr: false,
		},
		"check for parent device in use": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sdb1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent: "/dev/sdb",
				},
			},
			want:    true,
			wantErr: false,
		},
		"non existent parent device": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sdc1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent: "/dev/sdc",
				},
			},
			want:    false,
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotErr := pe.isParentDeviceInUse(tt.bd)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, gotErr != nil)
		})
	}
}

func TestGetExistingBDWithFsUuid(t *testing.T) {

	fakeFSUUID := "fake-fs-uuid"

	tests := map[string]struct {
		bd        blockdevice.BlockDevice
		bdAPIList *apis.BlockDeviceList
		want      *apis.BlockDevice
	}{
		"bd does not have a filesystem": {
			bd: blockdevice.BlockDevice{},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "blockdevice-123",
						},
					},
				},
			},
			want: nil,
		},
		"bd with fs uuid exists": {
			bd: blockdevice.BlockDevice{
				FSInfo: blockdevice.FileSystemInformation{
					FileSystemUUID: fakeFSUUID,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "blockdevice-123",
							Annotations: map[string]string{
								internalUUIDSchemeAnnotation: legacyUUIDScheme,
								internalFSUUIDAnnotation:     fakeFSUUID,
							},
						},
					},
				},
			},
			want: &apis.BlockDevice{
				ObjectMeta: metav1.ObjectMeta{
					Name: "blockdevice-123",
					Annotations: map[string]string{
						internalUUIDSchemeAnnotation: legacyUUIDScheme,
						internalFSUUIDAnnotation:     fakeFSUUID,
					},
				},
			},
		},
		"bd with fs uuid does not exists": {
			bd: blockdevice.BlockDevice{
				FSInfo: blockdevice.FileSystemInformation{
					FileSystemUUID: fakeFSUUID,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "blockdevice-123",
							Annotations: map[string]string{
								internalUUIDSchemeAnnotation: legacyUUIDScheme,
								internalFSUUIDAnnotation:     "12345",
							},
						},
					},
				},
			},
			want: nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := getExistingBDWithFsUuid(tt.bd, tt.bdAPIList)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetExistingBDWithPartitionUUID(t *testing.T) {
	fakePartTableUUID := "fake-part-table-uuid"
	tests := map[string]struct {
		bd        blockdevice.BlockDevice
		bdAPIList *apis.BlockDeviceList
		want      *apis.BlockDevice
	}{
		"bd does not have a partition table": {
			bd: blockdevice.BlockDevice{},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "blockdevice-123",
						},
					},
				},
			},
			want: nil,
		},
		"bd with partition table uuid exists": {
			bd: blockdevice.BlockDevice{
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: fakePartTableUUID,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "blockdevice-123",
							Annotations: map[string]string{
								internalUUIDSchemeAnnotation:    legacyUUIDScheme,
								internalPartitionUUIDAnnotation: fakePartTableUUID,
							},
						},
					},
				},
			},
			want: &apis.BlockDevice{
				ObjectMeta: metav1.ObjectMeta{
					Name: "blockdevice-123",
					Annotations: map[string]string{
						internalUUIDSchemeAnnotation:    legacyUUIDScheme,
						internalPartitionUUIDAnnotation: fakePartTableUUID,
					},
				},
			},
		},
		"bd with fs uuid does not exists": {
			bd: blockdevice.BlockDevice{
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: fakePartTableUUID,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "blockdevice-123",
							Annotations: map[string]string{
								internalUUIDSchemeAnnotation:    legacyUUIDScheme,
								internalPartitionUUIDAnnotation: "12345",
							},
						},
					},
				},
			},
			want: nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := getExistingBDWithPartitionUUID(tt.bd, tt.bdAPIList)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestHandleUnmanagedDevices(t *testing.T) {

	fakePartTableID := "fake-part-table-uuid"
	fakeBD := blockdevice.BlockDevice{
		PartitionInfo: blockdevice.PartitionInformation{
			PartitionTableUUID: fakePartTableID,
		},
	}

	fakeUUID, _ := generateUUIDFromPartitionTable(fakeBD)
	tests := map[string]struct {
		bd                     blockdevice.BlockDevice
		bdAPIList              *apis.BlockDeviceList
		bdCache                blockdevice.Hierarchy
		createdOrUpdatedBDName string
		want                   bool
		wantErr                bool
	}{
		"device not in use": {
			bd: blockdevice.BlockDevice{
				DevUse: blockdevice.DeviceUsage{
					InUse: false,
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
			},
			bdAPIList:              &apis.BlockDeviceList{},
			bdCache:                nil,
			createdOrUpdatedBDName: "",
			want:                   true,
			wantErr:                false,
		},
		"device in use, but not by mayastor or zfs localPV": {
			bd: blockdevice.BlockDevice{
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.LocalPV,
				},
			},
			bdAPIList:              &apis.BlockDeviceList{},
			bdCache:                nil,
			createdOrUpdatedBDName: "",
			want:                   true,
			wantErr:                false,
		},
		"device in use by mayastor": {
			bd: blockdevice.BlockDevice{
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.Mayastor,
				},
			},
			bdAPIList:              &apis.BlockDeviceList{},
			bdCache:                nil,
			createdOrUpdatedBDName: "",
			want:                   false,
			wantErr:                false,
		},
		"device in use, not by zfs localPV": {
			bd: blockdevice.BlockDevice{
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.CStor,
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
			},
			bdAPIList:              &apis.BlockDeviceList{},
			bdCache:                nil,
			createdOrUpdatedBDName: "",
			want:                   true,
			wantErr:                false,
		},
		"deviceType partition, parent device used by zfs localPV": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.ZFSLocalPV,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent: "/dev/sda",
				},
			},
			bdAPIList: &apis.BlockDeviceList{},
			bdCache: blockdevice.Hierarchy{
				"/dev/sda": {
					DevUse: blockdevice.DeviceUsage{
						InUse:  true,
						UsedBy: blockdevice.ZFSLocalPV,
					},
				},
			},
			createdOrUpdatedBDName: "",
			want:                   false,
			wantErr:                false,
		},
		"deviceType partition, parent device used by cstor": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent: "/dev/sda",
				},
			},
			bdAPIList: &apis.BlockDeviceList{},
			bdCache: blockdevice.Hierarchy{
				"/dev/sda": {
					DevUse: blockdevice.DeviceUsage{
						InUse:  true,
						UsedBy: blockdevice.CStor,
					},
				},
			},
			createdOrUpdatedBDName: "",
			want:                   true,
			wantErr:                false,
		},
		// if multiple partitions are there, this test may need to be revisited
		"deviceType partition, parent device not in use": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent: "/dev/sda",
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.ZFSLocalPV,
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: fakePartTableID,
				},
			},
			bdAPIList: &apis.BlockDeviceList{},
			bdCache: blockdevice.Hierarchy{
				"/dev/sda": {
					DevUse: blockdevice.DeviceUsage{
						InUse: false,
					},
				},
			},
			createdOrUpdatedBDName: fakeUUID,
			want:                   false,
			wantErr:                false,
		},
		"deviceType disk, used by zfs PV and is connected to the cluster for the first time": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.ZFSLocalPV,
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: fakePartTableID,
				},
			},
			bdAPIList:              &apis.BlockDeviceList{},
			bdCache:                nil,
			createdOrUpdatedBDName: fakeUUID,
			want:                   false,
			wantErr:                false,
		},
		"deviceType disk, used by zfs PV and is moved from disconnected and reconnected to the node at a different path": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.ZFSLocalPV,
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: fakePartTableID,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: fakeUUID,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdb",
						},
					},
				},
			},
			bdCache:                nil,
			createdOrUpdatedBDName: fakeUUID,
			want:                   false,
			wantErr:                false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := scheme.Scheme
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDevice{})
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDeviceList{})
			cl := fake.NewFakeClientWithScheme(s)

			// initialize client with all the bd resources
			for _, bdAPI := range tt.bdAPIList.Items {
				cl.Create(context.TODO(), &bdAPI)
			}

			ctrl := &controller.Controller{
				Clientset:   cl,
				BDHierarchy: tt.bdCache,
			}
			pe := &ProbeEvent{
				Controller: ctrl,
			}
			got, err := pe.handleUnmanagedDevices(tt.bd, tt.bdAPIList)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleUnmanagedDevices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)

			// check if a BD has been created or updated
			if len(tt.createdOrUpdatedBDName) != 0 {
				gotBDAPI := &apis.BlockDevice{}
				err := cl.Get(context.TODO(), client.ObjectKey{Name: tt.createdOrUpdatedBDName}, gotBDAPI)
				if err != nil {
					t.Errorf("error in getting blockdevice %s", tt.createdOrUpdatedBDName)
				}
				// verify the block-device-tag on the resource, also verify the path and node name
				assert.Equal(t, string(tt.bd.DevUse.UsedBy), gotBDAPI.GetLabels()[kubernetes.BlockDeviceTagLabel])
				assert.Equal(t, tt.bd.DevPath, gotBDAPI.Spec.Path)
				assert.Equal(t, tt.bd.NodeAttributes[blockdevice.NodeName], gotBDAPI.Spec.NodeAttributes.NodeName)
			}
		})
	}
}

func TestCreateBlockDeviceResourceIfNoHolders(t *testing.T) {
	tests := map[string]struct {
		bd                     blockdevice.BlockDevice
		bdAPIList              *apis.BlockDeviceList
		createdOrUpdatedBDName string
		wantErr                bool
	}{
		"bd does not have holder": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
					UUID:    "blockdevice-123",
				},
				DependentDevices: blockdevice.DependentBlockDevices{},
			},
			bdAPIList:              &apis.BlockDeviceList{},
			createdOrUpdatedBDName: "blockdevice-123",
			wantErr:                false,
		},
		"bd has holder devices": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
					UUID:    "blockdevice-123",
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Holders: []string{
						"/dev/dm-0", "/dev/dm-1",
					},
				},
			},
			bdAPIList:              &apis.BlockDeviceList{},
			createdOrUpdatedBDName: "",
			wantErr:                false,
		},
		"bd without holder has been disconnected and reconnected at different path": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
					UUID:    "blockdevice-123",
				},
				DependentDevices: blockdevice.DependentBlockDevices{},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "blockdevice-123",
							Labels: make(map[string]string),
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sda",
						},
					},
				},
			},
			createdOrUpdatedBDName: "blockdevice-123",
			wantErr:                false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := scheme.Scheme
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDevice{})
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDeviceList{})
			cl := fake.NewFakeClientWithScheme(s)

			// initialize client with all the bd resources
			for _, bdAPI := range tt.bdAPIList.Items {
				cl.Create(context.TODO(), &bdAPI)
			}

			ctrl := &controller.Controller{
				Clientset: cl,
			}
			pe := &ProbeEvent{
				Controller: ctrl,
			}
			if err := pe.createBlockDeviceResourceIfNoHolders(tt.bd, tt.bdAPIList); (err != nil) != tt.wantErr {
				t.Errorf("createBlockDeviceResourceIfNoHolders() error = %v, wantErr %v", err, tt.wantErr)
			}

			// check if a BD has been created or updated
			if len(tt.createdOrUpdatedBDName) != 0 {
				gotBDAPI := &apis.BlockDevice{}
				err := cl.Get(context.TODO(), client.ObjectKey{Name: tt.createdOrUpdatedBDName}, gotBDAPI)
				if err != nil {
					t.Errorf("error in getting blockdevice %s", tt.createdOrUpdatedBDName)
				}
				// verify the uuid scheme on the resource, also verify the path and node name
				assert.Equal(t, gptUUIDScheme, gotBDAPI.GetAnnotations()[internalUUIDSchemeAnnotation])
				assert.Equal(t, tt.bd.DevPath, gotBDAPI.Spec.Path)
				assert.Equal(t, tt.bd.NodeAttributes[blockdevice.NodeName], gotBDAPI.Spec.NodeAttributes.NodeName)
			}
		})
	}
}

func TestUpgradeDeviceInUseByCStor(t *testing.T) {

	physicalBlockDevice := blockdevice.BlockDevice{
		Identifier: blockdevice.Identifier{
			DevPath: "/dev/sda",
		},
		DeviceAttributes: blockdevice.DeviceAttribute{
			WWN:        fakeWWN,
			Serial:     fakeSerial,
			Model:      "SanDiskSSD",
			DeviceType: blockdevice.BlockDeviceTypeDisk,
			IDType:     blockdevice.BlockDeviceTypeDisk,
		},
	}

	virtualBlockDevice := blockdevice.BlockDevice{
		Identifier: blockdevice.Identifier{
			DevPath: "/dev/sda",
		},
		DeviceAttributes: blockdevice.DeviceAttribute{
			Model:      "Virtual_disk",
			DeviceType: blockdevice.BlockDeviceTypeDisk,
		},
	}

	fakePartitionEntry := "fake-part-entry-1"
	fakePartTable := "fake-part-table"

	gptUuidForPhysicalDevice, _ := generateUUID(physicalBlockDevice)
	gptUuidForPhysicalDevicePartition := blockdevice.BlockDevicePrefix + util.Hash(fakePartitionEntry)
	legacyUuidForPhysicalDevice, _ := generateLegacyUUID(physicalBlockDevice)
	legacyUuidForVirtualDevice, _ := generateLegacyUUID(virtualBlockDevice)

	tests := map[string]struct {
		bd                     blockdevice.BlockDevice
		bdAPIList              *apis.BlockDeviceList
		bdCache                blockdevice.Hierarchy
		createdOrUpdatedBDName string
		want                   bool
		wantErr                bool
	}{
		"deviceType: disk, using gpt based algorithm": {
			bd: physicalBlockDevice,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: gptUuidForPhysicalDevice,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "",
			want:                   true,
			wantErr:                false,
		},
		"deviceType: partition, using gpt based algorithm": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionEntryUUID: fakePartitionEntry,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: gptUuidForPhysicalDevicePartition,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "",
			want:                   true,
			wantErr:                false,
		},
		"deviceType: disk, using gpt algorithm, but resource is in unclaimed state": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: gptUuidForPhysicalDevice,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "",
			want:                   false,
			wantErr:                true,
		},
		"deviceType: disk, resource with legacy UUID is present in not unclaimed state": {
			bd: physicalBlockDevice,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        legacyUuidForPhysicalDevice,
							Annotations: make(map[string]string),
							Labels:      make(map[string]string),
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: legacyUuidForPhysicalDevice,
			want:                   false,
			wantErr:                false,
		},
		"deviceType: disk, resource with matching partition uuid annotation is present in not unclaimed state": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					Model:      "SanDiskSSD",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: fakePartTable,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "blockdevice-123",
							Labels: make(map[string]string),
							Annotations: map[string]string{
								internalPartitionUUIDAnnotation: fakePartTable,
								internalUUIDSchemeAnnotation:    legacyUUIDScheme,
							},
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "blockdevice-123",
			want:                   false,
			wantErr:                false,
		},
		"deviceType: disk, no resource with legacy UUID or matching partition UUID": {
			bd:                     physicalBlockDevice,
			bdAPIList:              &apis.BlockDeviceList{},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: legacyUuidForPhysicalDevice,
			want:                   false,
			wantErr:                false,
		},
		"deviceType: disk, resource with both legacy uuid and matching partition uuid is present": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					Model:      "SanDiskSSD",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: fakePartTable,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "blockdevice-123",
							Labels: make(map[string]string),
							Annotations: map[string]string{
								internalPartitionUUIDAnnotation: fakePartTable,
								internalUUIDSchemeAnnotation:    legacyUUIDScheme,
							},
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        legacyUuidForPhysicalDevice,
							Labels:      make(map[string]string),
							Annotations: make(map[string]string),
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "blockdevice-123",
			want:                   false,
			wantErr:                false,
		},
		"deviceType: disk, resource with legacy UUID is present in unclaimed state and device is virtual": {
			bd: virtualBlockDevice,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        legacyUuidForPhysicalDevice,
							Annotations: make(map[string]string),
							Labels:      make(map[string]string),
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: legacyUuidForVirtualDevice,
			want:                   false,
			wantErr:                false,
		},
		"deviceType: disk, resource with matching partition uuid annotation is present in unclaimed state and device is virtual": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					Model:      "Virtual_disk",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: fakePartTable,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "blockdevice-123",
							Labels: make(map[string]string),
							Annotations: map[string]string{
								internalPartitionUUIDAnnotation: fakePartTable,
								internalUUIDSchemeAnnotation:    legacyUUIDScheme,
							},
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "blockdevice-123",
			want:                   false,
			wantErr:                false,
		},
		"deviceType: disk, resource with legacy UUID is present in unclaimed state and device is not virtual": {
			bd: physicalBlockDevice,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        legacyUuidForPhysicalDevice,
							Annotations: make(map[string]string),
							Labels:      make(map[string]string),
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: legacyUuidForPhysicalDevice,
			want:                   false,
			wantErr:                true,
		},
		"deviceType: disk, resource with matching partition uuid annotation is present in unclaimed state is not virtual": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					Model:      "SanDiskSSD",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: fakePartTable,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "blockdevice-123",
							Annotations: map[string]string{
								internalPartitionUUIDAnnotation: fakePartTable,
								internalUUIDSchemeAnnotation:    legacyUUIDScheme,
							},
							Labels: make(map[string]string),
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "blockdevice-123",
			want:                   false,
			wantErr:                true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := scheme.Scheme
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDevice{})
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDeviceList{})
			cl := fake.NewFakeClientWithScheme(s)

			// initialize client with all the bd resources
			for _, bdAPI := range tt.bdAPIList.Items {
				cl.Create(context.TODO(), &bdAPI)
			}

			ctrl := &controller.Controller{
				Clientset:   cl,
				BDHierarchy: tt.bdCache,
			}
			pe := &ProbeEvent{
				Controller: ctrl,
			}
			got, err := pe.upgradeDeviceInUseByCStor(tt.bd, tt.bdAPIList)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("upgradeDeviceInUseByCStor() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			assert.Equal(t, tt.want, got)

			// check if a BD has been created or updated
			if len(tt.createdOrUpdatedBDName) != 0 {
				gotBDAPI := &apis.BlockDevice{}
				err := cl.Get(context.TODO(), client.ObjectKey{Name: tt.createdOrUpdatedBDName}, gotBDAPI)
				if err != nil {
					t.Errorf("error in getting blockdevice %s: %v", tt.createdOrUpdatedBDName, err)
				}
				// verify the annotation on the resource, also verify the path and node name
				assert.Equal(t, legacyUUIDScheme, gotBDAPI.GetAnnotations()[internalUUIDSchemeAnnotation])
				assert.Equal(t, tt.bd.PartitionInfo.PartitionTableUUID, gotBDAPI.GetAnnotations()[internalPartitionUUIDAnnotation])
				assert.Equal(t, tt.bd.DevPath, gotBDAPI.Spec.Path)
				assert.Equal(t, tt.bd.NodeAttributes[blockdevice.NodeName], gotBDAPI.Spec.NodeAttributes.NodeName)
			}
		})
	}
}

func TestUpgradeDeviceInUseByLocalPV(t *testing.T) {
	physicalBlockDevice := blockdevice.BlockDevice{
		Identifier: blockdevice.Identifier{
			DevPath: "/dev/sda",
		},
		DeviceAttributes: blockdevice.DeviceAttribute{
			WWN:        fakeWWN,
			Serial:     fakeSerial,
			Model:      "SanDiskSSD",
			DeviceType: blockdevice.BlockDeviceTypeDisk,
			IDType:     blockdevice.BlockDeviceTypeDisk,
		},
	}

	virtualBlockDevice := blockdevice.BlockDevice{
		Identifier: blockdevice.Identifier{
			DevPath: "/dev/sda",
		},
		DeviceAttributes: blockdevice.DeviceAttribute{
			Model:      "Virtual_disk",
			DeviceType: blockdevice.BlockDeviceTypeDisk,
		},
	}

	fakePartitionEntry := "fake-part-entry-1"
	fakefsUuid := "fake-fs-uuid"

	gptUuidForPhysicalDevice, _ := generateUUID(physicalBlockDevice)
	gptUuidForPhysicalDevicePartition := blockdevice.BlockDevicePrefix + util.Hash(fakePartitionEntry)
	legacyUuidForPhysicalDevice, _ := generateLegacyUUID(physicalBlockDevice)
	legacyUuidForVirtualDevice, _ := generateLegacyUUID(virtualBlockDevice)

	tests := map[string]struct {
		bd                     blockdevice.BlockDevice
		bdAPIList              *apis.BlockDeviceList
		bdCache                blockdevice.Hierarchy
		createdOrUpdatedBDName string
		want                   bool
		wantErr                bool
	}{
		"deviceType: disk, using gpt based algorithm": {
			bd: physicalBlockDevice,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: gptUuidForPhysicalDevice,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "",
			want:                   true,
			wantErr:                false,
		},
		"deviceType: partition, using gpt based algorithm": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionEntryUUID: fakePartitionEntry,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: gptUuidForPhysicalDevicePartition,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "",
			want:                   true,
			wantErr:                false,
		},
		"deviceType: disk, using gpt algorithm, but resource is in unclaimed state": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: gptUuidForPhysicalDevice,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "",
			want:                   false,
			wantErr:                true,
		},
		"deviceType: disk, resource with legacy UUID is present in not unclaimed state": {
			bd: physicalBlockDevice,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        legacyUuidForPhysicalDevice,
							Annotations: make(map[string]string),
							Labels:      make(map[string]string),
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: legacyUuidForPhysicalDevice,
			want:                   false,
			wantErr:                false,
		},
		"deviceType: disk, resource with matching fs uuid annotation is present in not unclaimed state": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					Model:      "SanDiskSSD",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
				FSInfo: blockdevice.FileSystemInformation{
					FileSystemUUID: fakefsUuid,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "blockdevice-123",
							Labels: make(map[string]string),
							Annotations: map[string]string{
								internalFSUUIDAnnotation:     fakefsUuid,
								internalUUIDSchemeAnnotation: legacyUUIDScheme,
							},
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "blockdevice-123",
			want:                   false,
			wantErr:                false,
		},
		"deviceType: disk, no resource with legacy UUID or matching fs UUID": {
			bd:                     physicalBlockDevice,
			bdAPIList:              &apis.BlockDeviceList{},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: legacyUuidForPhysicalDevice,
			want:                   false,
			wantErr:                false,
		},
		"deviceType: disk, resource with both legacy uuid and matching fs uuid is present": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					Model:      "SanDiskSSD",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
				FSInfo: blockdevice.FileSystemInformation{
					FileSystemUUID: fakefsUuid,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "blockdevice-123",
							Labels: make(map[string]string),
							Annotations: map[string]string{
								internalFSUUIDAnnotation:     fakefsUuid,
								internalUUIDSchemeAnnotation: legacyUUIDScheme,
							},
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        legacyUuidForPhysicalDevice,
							Labels:      make(map[string]string),
							Annotations: make(map[string]string),
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "blockdevice-123",
			want:                   false,
			wantErr:                false,
		},
		"deviceType: disk, resource with legacy UUID is present in unclaimed state and device is virtual": {
			bd: virtualBlockDevice,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        legacyUuidForPhysicalDevice,
							Annotations: make(map[string]string),
							Labels:      make(map[string]string),
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: legacyUuidForVirtualDevice,
			want:                   false,
			wantErr:                false,
		},
		"deviceType: disk, resource with matching fs uuid annotation is present in unclaimed state and device is virtual": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					Model:      "Virtual_disk",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
				FSInfo: blockdevice.FileSystemInformation{
					FileSystemUUID: fakefsUuid,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "blockdevice-123",
							Labels: make(map[string]string),
							Annotations: map[string]string{
								internalFSUUIDAnnotation:     fakefsUuid,
								internalUUIDSchemeAnnotation: legacyUUIDScheme,
							},
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "blockdevice-123",
			want:                   false,
			wantErr:                false,
		},
		"deviceType: disk, resource with legacy UUID is present in unclaimed state and device is not virtual": {
			bd: physicalBlockDevice,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        legacyUuidForPhysicalDevice,
							Annotations: make(map[string]string),
							Labels:      make(map[string]string),
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: legacyUuidForPhysicalDevice,
			want:                   false,
			wantErr:                true,
		},
		"deviceType: disk, resource with matching fs uuid annotation is present in unclaimed state is not virtual": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					Model:      "SanDiskSSD",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
				FSInfo: blockdevice.FileSystemInformation{
					FileSystemUUID: fakefsUuid,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "blockdevice-123",
							Annotations: map[string]string{
								internalFSUUIDAnnotation:     fakefsUuid,
								internalUUIDSchemeAnnotation: legacyUUIDScheme,
							},
							Labels: make(map[string]string),
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "blockdevice-123",
			want:                   false,
			wantErr:                true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := scheme.Scheme
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDevice{})
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDeviceList{})
			cl := fake.NewFakeClientWithScheme(s)

			// initialize client with all the bd resources
			for _, bdAPI := range tt.bdAPIList.Items {
				cl.Create(context.TODO(), &bdAPI)
			}

			ctrl := &controller.Controller{
				Clientset:   cl,
				BDHierarchy: tt.bdCache,
			}
			pe := &ProbeEvent{
				Controller: ctrl,
			}
			got, err := pe.upgradeDeviceInUseByLocalPV(tt.bd, tt.bdAPIList)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("upgradeDeviceInUseByLocalPV() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			assert.Equal(t, tt.want, got)

			// check if a BD has been created or updated
			if len(tt.createdOrUpdatedBDName) != 0 {
				gotBDAPI := &apis.BlockDevice{}
				err := cl.Get(context.TODO(), client.ObjectKey{Name: tt.createdOrUpdatedBDName}, gotBDAPI)
				if err != nil {
					t.Errorf("error in getting blockdevice %s: %v", tt.createdOrUpdatedBDName, err)
				}
				// verify the annotation on the resource, also verify the path and node name
				assert.Equal(t, legacyUUIDScheme, gotBDAPI.GetAnnotations()[internalUUIDSchemeAnnotation])
				assert.Equal(t, tt.bd.FSInfo.FileSystemUUID, gotBDAPI.GetAnnotations()[internalFSUUIDAnnotation])
				assert.Equal(t, tt.bd.DevPath, gotBDAPI.Spec.Path)
				assert.Equal(t, tt.bd.NodeAttributes[blockdevice.NodeName], gotBDAPI.Spec.NodeAttributes.NodeName)
			}
		})
	}
}

func TestUpgradeBD(t *testing.T) {
	physicalBlockDevice := blockdevice.BlockDevice{
		Identifier: blockdevice.Identifier{
			DevPath: "/dev/sda",
		},
		DeviceAttributes: blockdevice.DeviceAttribute{
			WWN:        fakeWWN,
			Serial:     fakeSerial,
			Model:      "SanDiskSSD",
			DeviceType: blockdevice.BlockDeviceTypeDisk,
			IDType:     blockdevice.BlockDeviceTypeDisk,
		},
	}

	legacyUuidForPhysicalDevice, _ := generateLegacyUUID(physicalBlockDevice)

	tests := map[string]struct {
		bd                     blockdevice.BlockDevice
		bdAPIList              *apis.BlockDeviceList
		bdCache                blockdevice.Hierarchy
		createdOrUpdatedBDName string
		want                   bool
		wantErr                bool
	}{
		"device not in use": {
			bd: blockdevice.BlockDevice{
				DevUse: blockdevice.DeviceUsage{
					InUse: false,
				},
			},
			bdAPIList:              &apis.BlockDeviceList{},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "",
			want:                   true,
			wantErr:                false,
		},
		"device in use, but not used by cstor or localPV": {
			bd: blockdevice.BlockDevice{
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.Jiva,
				},
			},
			bdAPIList:              &apis.BlockDeviceList{},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "",
			want:                   true,
			wantErr:                false,
		},
		"device in use by cstor": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					Model:      "SanDiskSSD",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
					IDType:     blockdevice.BlockDeviceTypeDisk,
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.CStor,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        legacyUuidForPhysicalDevice,
							Annotations: make(map[string]string),
							Labels:      make(map[string]string),
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: legacyUuidForPhysicalDevice,
			want:                   false,
			wantErr:                false,
		},
		"device in use by localpv": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					Model:      "SanDiskSSD",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
					IDType:     blockdevice.BlockDeviceTypeDisk,
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.LocalPV,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        legacyUuidForPhysicalDevice,
							Annotations: make(map[string]string),
							Labels:      make(map[string]string),
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: legacyUuidForPhysicalDevice,
			want:                   false,
			wantErr:                false,
		},
		"device in use by cstor with invalid state": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					Model:      "SanDiskSSD",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
					IDType:     blockdevice.BlockDeviceTypeDisk,
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.CStor,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        legacyUuidForPhysicalDevice,
							Annotations: make(map[string]string),
							Labels:      make(map[string]string),
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: legacyUuidForPhysicalDevice,
			want:                   false,
			wantErr:                true,
		},
		"device in use by localPV with invalid state": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					Model:      "SanDiskSSD",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
					IDType:     blockdevice.BlockDeviceTypeDisk,
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.LocalPV,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        legacyUuidForPhysicalDevice,
							Annotations: make(map[string]string),
							Labels:      make(map[string]string),
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: legacyUuidForPhysicalDevice,
			want:                   false,
			wantErr:                true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := scheme.Scheme
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDevice{})
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDeviceList{})
			cl := fake.NewFakeClientWithScheme(s)

			// initialize client with all the bd resources
			for _, bdAPI := range tt.bdAPIList.Items {
				cl.Create(context.TODO(), &bdAPI)
			}

			ctrl := &controller.Controller{
				Clientset:   cl,
				BDHierarchy: tt.bdCache,
			}
			pe := &ProbeEvent{
				Controller: ctrl,
			}
			got, err := pe.upgradeBD(tt.bd, tt.bdAPIList)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("upgradeDeviceInUseByLocalPV() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			assert.Equal(t, tt.want, got)

			// check if a BD has been created or updated
			if len(tt.createdOrUpdatedBDName) != 0 {
				gotBDAPI := &apis.BlockDevice{}
				err := cl.Get(context.TODO(), client.ObjectKey{Name: tt.createdOrUpdatedBDName}, gotBDAPI)
				if err != nil {
					t.Errorf("error in getting blockdevice %s: %v", tt.createdOrUpdatedBDName, err)
				}
				// verify the annotation on the resource, also verify the path and node name
				assert.Equal(t, legacyUUIDScheme, gotBDAPI.GetAnnotations()[internalUUIDSchemeAnnotation])
				if tt.bd.DevUse.UsedBy == blockdevice.CStor {
					assert.Equal(t, tt.bd.PartitionInfo.PartitionTableUUID, gotBDAPI.GetAnnotations()[internalPartitionUUIDAnnotation])
				} else {
					assert.Equal(t, tt.bd.FSInfo.FileSystemUUID, gotBDAPI.GetAnnotations()[internalFSUUIDAnnotation])
				}
				assert.Equal(t, tt.bd.DevPath, gotBDAPI.Spec.Path)
				assert.Equal(t, tt.bd.NodeAttributes[blockdevice.NodeName], gotBDAPI.Spec.NodeAttributes.NodeName)
			}
		})
	}
}

func TestAddBlockDevice(t *testing.T) {
	fakePartTableID := "fake-part-table-uuid"
	fakePartEntryID := "fake-part-entry-1"
	fakeBD := blockdevice.BlockDevice{
		PartitionInfo: blockdevice.PartitionInformation{
			PartitionTableUUID: fakePartTableID,
		},
	}
	physicalBlockDevice := blockdevice.BlockDevice{
		Identifier: blockdevice.Identifier{
			DevPath: "/dev/sda",
		},
		DeviceAttributes: blockdevice.DeviceAttribute{
			WWN:        fakeWWN,
			Serial:     fakeSerial,
			Model:      "SanDiskSSD",
			DeviceType: blockdevice.BlockDeviceTypeDisk,
			IDType:     blockdevice.BlockDeviceTypeDisk,
		},
	}
	fakeBDForPartition := blockdevice.BlockDevice{
		DeviceAttributes: blockdevice.DeviceAttribute{
			DeviceType: blockdevice.BlockDeviceTypePartition,
		},
		PartitionInfo: blockdevice.PartitionInformation{
			PartitionEntryUUID: fakePartEntryID,
		},
	}

	fakeUUID, _ := generateUUIDFromPartitionTable(fakeBD)
	gptUuidForPhysicalDevice, _ := generateUUID(physicalBlockDevice)
	gptUuidForPartition, _ := generateUUID(fakeBDForPartition)
	legacyUuidForPhysicalDevice, _ := generateLegacyUUID(physicalBlockDevice)

	tests := map[string]struct {
		bd                     blockdevice.BlockDevice
		bdAPIList              *apis.BlockDeviceList
		bdCache                blockdevice.Hierarchy
		createdOrUpdatedBDName string
		wantErr                bool
	}{
		"device used by mayastor": {
			bd: blockdevice.BlockDevice{
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.Mayastor,
				},
			},
			bdAPIList:              &apis.BlockDeviceList{},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "",
			wantErr:                false,
		},
		"device used by zfs-localpv": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.ZFSLocalPV,
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: fakePartTableID,
				},
			},
			bdAPIList:              &apis.BlockDeviceList{},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: fakeUUID,
			wantErr:                false,
		},
		"deviceType partition, but parent device is in use": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.CStor,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent: "/dev/sda",
				},
			},
			bdAPIList: &apis.BlockDeviceList{},
			bdCache: map[string]blockdevice.BlockDevice{
				"/dev/sda": {
					DevUse: blockdevice.DeviceUsage{
						InUse: true,
					},
				},
			},
			createdOrUpdatedBDName: "",
			wantErr:                false,
		},
		"device used by cstor with legacy UUID": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					Model:      "SanDiskSSD",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
					IDType:     blockdevice.BlockDeviceTypeDisk,
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.CStor,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        legacyUuidForPhysicalDevice,
							Annotations: make(map[string]string),
							Labels:      make(map[string]string),
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: legacyUuidForPhysicalDevice,
			wantErr:                false,
		},
		"device used by localPV with legacy UUID": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					Model:      "SanDiskSSD",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
					IDType:     blockdevice.BlockDeviceTypeDisk,
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.LocalPV,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        legacyUuidForPhysicalDevice,
							Annotations: make(map[string]string),
							Labels:      make(map[string]string),
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdX",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: legacyUuidForPhysicalDevice,
			wantErr:                false,
		},
		"unused virtual disk with partitions/holders": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					Model:      "Virtual_disk",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
				DevUse: blockdevice.DeviceUsage{
					InUse: false,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Holders: []string{"/dev/dm-0"},
				},
			},
			bdAPIList:              &apis.BlockDeviceList{},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: "",
			wantErr:                false,
		},
		// test case for virtual disk without partition is not added, since it needs a write operation
		// on the disk
		"unused physical disk moved from a different node": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				NodeAttributes: map[string]string{
					blockdevice.NodeName: "node1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					Model:      "SanDiskSSD",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
					IDType:     blockdevice.BlockDeviceTypeDisk,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        gptUuidForPhysicalDevice,
							Labels:      make(map[string]string),
							Annotations: make(map[string]string),
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdx",
							NodeAttributes: apis.NodeAttribute{
								NodeName: "node0",
							},
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: gptUuidForPhysicalDevice,
			wantErr:                false,
		},
		"used physical disk moved from a different node": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				NodeAttributes: map[string]string{
					blockdevice.NodeName: "node1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					Model:      "SanDiskSSD",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
					IDType:     blockdevice.BlockDeviceTypeDisk,
				},
				DevUse: blockdevice.DeviceUsage{
					InUse:  true,
					UsedBy: blockdevice.CStor,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: gptUuidForPhysicalDevice,
							Labels: map[string]string{
								kubernetes.KubernetesHostNameLabel: "node0",
							},
							Annotations: make(map[string]string),
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdx",
							NodeAttributes: apis.NodeAttribute{
								NodeName: "node0",
							},
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: gptUuidForPhysicalDevice,
			wantErr:                false,
		},
		"deviceType: partition, with parent device resource not present": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent: "/dev/sda",
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: fakePartTableID,
					PartitionEntryUUID: fakePartEntryID,
				},
			},
			bdAPIList: &apis.BlockDeviceList{},
			bdCache: map[string]blockdevice.BlockDevice{
				"/dev/sda": {
					Identifier: blockdevice.Identifier{
						DevPath: "/dev/sda",
					},
					DeviceAttributes: blockdevice.DeviceAttribute{
						DeviceType: blockdevice.BlockDeviceTypePartition,
					},
					DependentDevices: blockdevice.DependentBlockDevices{
						Partitions: []string{"/dev/sda1"},
					},
					PartitionInfo: blockdevice.PartitionInformation{
						PartitionTableUUID: fakePartTableID,
					},
				},
			},
			createdOrUpdatedBDName: gptUuidForPartition,
			wantErr:                false,
		},
		"deviceType: partition, with parent device in use": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent: "/dev/sda",
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: fakePartTableID,
					PartitionEntryUUID: fakePartEntryID,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: gptUuidForPhysicalDevice,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sda",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
						},
					},
				},
			},
			bdCache: map[string]blockdevice.BlockDevice{
				"/dev/sda": {
					Identifier: blockdevice.Identifier{
						DevPath: "/dev/sda",
					},
					DeviceAttributes: blockdevice.DeviceAttribute{
						WWN:        fakeWWN,
						Serial:     fakeSerial,
						DeviceType: blockdevice.BlockDeviceTypePartition,
					},
					DependentDevices: blockdevice.DependentBlockDevices{
						Partitions: []string{"/dev/sda1"},
					},
					PartitionInfo: blockdevice.PartitionInformation{
						PartitionTableUUID: fakePartTableID,
					},
					DevUse: blockdevice.DeviceUsage{
						InUse: true,
					},
				},
			},
			createdOrUpdatedBDName: "",
			wantErr:                false,
		},
		"deviceType: partition, with parent device not in use": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent: "/dev/sda",
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: fakePartTableID,
					PartitionEntryUUID: fakePartEntryID,
				},
			},
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: gptUuidForPhysicalDevice,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sda",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
						},
					},
				},
			},
			bdCache: map[string]blockdevice.BlockDevice{
				"/dev/sda": {
					Identifier: blockdevice.Identifier{
						DevPath: "/dev/sda",
					},
					DeviceAttributes: blockdevice.DeviceAttribute{
						WWN:        fakeWWN,
						Serial:     fakeSerial,
						DeviceType: blockdevice.BlockDeviceTypePartition,
					},
					DependentDevices: blockdevice.DependentBlockDevices{
						Partitions: []string{"/dev/sda1"},
					},
					PartitionInfo: blockdevice.PartitionInformation{
						PartitionTableUUID: fakePartTableID,
					},
				},
			},
			createdOrUpdatedBDName: gptUuidForPartition,
			wantErr:                false,
		},
		"new disk connected first time to cluster": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				NodeAttributes: map[string]string{
					blockdevice.NodeName: "node1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:        fakeWWN,
					Serial:     fakeSerial,
					Model:      "SanDiskSSD",
					DeviceType: blockdevice.BlockDeviceTypeDisk,
					IDType:     blockdevice.BlockDeviceTypeDisk,
				},
			},
			bdAPIList:              &apis.BlockDeviceList{},
			bdCache:                make(blockdevice.Hierarchy),
			createdOrUpdatedBDName: gptUuidForPhysicalDevice,
			wantErr:                false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := scheme.Scheme
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDevice{})
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDeviceList{})
			cl := fake.NewFakeClientWithScheme(s)

			// initialize client with all the bd resources
			for _, bdAPI := range tt.bdAPIList.Items {
				cl.Create(context.TODO(), &bdAPI)
			}

			ctrl := &controller.Controller{
				Clientset:   cl,
				BDHierarchy: tt.bdCache,
			}
			pe := &ProbeEvent{
				Controller: ctrl,
			}
			err := pe.addBlockDevice(tt.bd, tt.bdAPIList)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("addBlockDevice() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			// check if a BD has been created or updated
			if len(tt.createdOrUpdatedBDName) != 0 {
				gotBDAPI := &apis.BlockDevice{}
				err := cl.Get(context.TODO(), client.ObjectKey{Name: tt.createdOrUpdatedBDName}, gotBDAPI)
				if err != nil {
					t.Errorf("error in getting blockdevice %s: %v", tt.createdOrUpdatedBDName, err)
				}
				// verify the resource
				assert.Equal(t, tt.bd.DevPath, gotBDAPI.Spec.Path)
				assert.Equal(t, tt.bd.NodeAttributes[blockdevice.NodeName], gotBDAPI.Spec.NodeAttributes.NodeName)
			}
		})
	}
}

func TestProbeEvent_createOrUpdateWithFSUUID(t *testing.T) {
	tests := map[string]struct {
		bd                     blockdevice.BlockDevice
		existingBD             *apis.BlockDevice
		createdOrUpdatedBDName string
		wantErr                bool
	}{
		"existing resource has no annotation": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					UUID: "blockdevice-123",
				},
				FSInfo: blockdevice.FileSystemInformation{
					FileSystemUUID: "123",
				},
			},
			existingBD: &apis.BlockDevice{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "blockdevice-123",
					Annotations: make(map[string]string),
					Labels:      make(map[string]string),
				},
			},
			createdOrUpdatedBDName: "blockdevice-123",
			wantErr:                false,
		},
		"existing resource has annotation": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					UUID: "blockdevice-123",
				},
				FSInfo: blockdevice.FileSystemInformation{
					FileSystemUUID: "123",
				},
			},
			existingBD: &apis.BlockDevice{
				ObjectMeta: metav1.ObjectMeta{
					Name: "blockdevice-123",
					Annotations: map[string]string{
						"keyX": "valX",
					},
					Labels: make(map[string]string),
				},
			},
			createdOrUpdatedBDName: "blockdevice-123",
			wantErr:                false,
		},
		"resource does not exist": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					UUID: "blockdevice-123",
				},
				FSInfo: blockdevice.FileSystemInformation{
					FileSystemUUID: "123",
				},
			},
			existingBD:             nil,
			createdOrUpdatedBDName: "blockdevice-123",
			wantErr:                false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := scheme.Scheme
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDevice{})
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDeviceList{})
			cl := fake.NewFakeClientWithScheme(s)

			// initialize client with the bd resource
			if tt.existingBD != nil {
				cl.Create(context.TODO(), tt.existingBD)
			}

			ctrl := &controller.Controller{
				Clientset: cl,
			}
			pe := &ProbeEvent{
				Controller: ctrl,
			}
			err := pe.createOrUpdateWithFSUUID(tt.bd, tt.existingBD)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("createOrUpdateWithAnnotation() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			// check if a BD has been created or updated
			if len(tt.createdOrUpdatedBDName) != 0 {
				gotBDAPI := &apis.BlockDevice{}
				err := cl.Get(context.TODO(), client.ObjectKey{Name: tt.createdOrUpdatedBDName}, gotBDAPI)
				if err != nil {
					t.Errorf("error in getting blockdevice %s: %v", tt.createdOrUpdatedBDName, err)
					return
				}
				assert.Equal(t, tt.bd.FSInfo.FileSystemUUID, gotBDAPI.GetAnnotations()[internalFSUUIDAnnotation])
				assert.Equal(t, legacyUUIDScheme, gotBDAPI.GetAnnotations()[internalUUIDSchemeAnnotation])
			}
		})
	}
}

func TestProbeEvent_createOrUpdateWithPartitionUUID(t *testing.T) {

	tests := map[string]struct {
		bd                     blockdevice.BlockDevice
		existingBD             *apis.BlockDevice
		createdOrUpdatedBDName string
		wantErr                bool
	}{
		"existing resource has no annotation": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					UUID: "blockdevice-123",
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: "123",
				},
			},
			existingBD: &apis.BlockDevice{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "blockdevice-123",
					Annotations: make(map[string]string),
					Labels:      make(map[string]string),
				},
			},
			createdOrUpdatedBDName: "blockdevice-123",
			wantErr:                false,
		},
		"existing resource has annotation": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					UUID: "blockdevice-123",
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: "123",
				},
			},
			existingBD: &apis.BlockDevice{
				ObjectMeta: metav1.ObjectMeta{
					Name: "blockdevice-123",
					Annotations: map[string]string{
						"keyX": "valX",
					},
					Labels: make(map[string]string),
				},
			},
			createdOrUpdatedBDName: "blockdevice-123",
			wantErr:                false,
		},
		"resource does not exist": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					UUID: "blockdevice-123",
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableUUID: "123",
				},
			},
			existingBD:             nil,
			createdOrUpdatedBDName: "blockdevice-123",
			wantErr:                false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := scheme.Scheme
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDevice{})
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDeviceList{})
			cl := fake.NewFakeClientWithScheme(s)

			// initialize client with the bd resource
			if tt.existingBD != nil {
				cl.Create(context.TODO(), tt.existingBD)
			}

			ctrl := &controller.Controller{
				Clientset: cl,
			}
			pe := &ProbeEvent{
				Controller: ctrl,
			}
			err := pe.createOrUpdateWithPartitionUUID(tt.bd, tt.existingBD)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("createOrUpdateWithAnnotation() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			// check if a BD has been created or updated
			if len(tt.createdOrUpdatedBDName) != 0 {
				gotBDAPI := &apis.BlockDevice{}
				err := cl.Get(context.TODO(), client.ObjectKey{Name: tt.createdOrUpdatedBDName}, gotBDAPI)
				if err != nil {
					t.Errorf("error in getting blockdevice %s: %v", tt.createdOrUpdatedBDName, err)
					return
				}
				assert.Equal(t, tt.bd.PartitionInfo.PartitionTableUUID, gotBDAPI.GetAnnotations()[internalPartitionUUIDAnnotation])
				assert.Equal(t, legacyUUIDScheme, gotBDAPI.GetAnnotations()[internalUUIDSchemeAnnotation])
			}
		})
	}
}

func TestCreateOrUpdateWithAnnotation(t *testing.T) {

	tests := map[string]struct {
		bd                     blockdevice.BlockDevice
		annotation             map[string]string
		existingBD             *apis.BlockDevice
		createdOrUpdatedBDName string
		wantErr                bool
	}{
		"existing resource has no annotation": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					UUID: "blockdevice-123",
				},
			},
			annotation: map[string]string{
				"key1": "val1",
				"key2": "val2",
			},
			existingBD: &apis.BlockDevice{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "blockdevice-123",
					Annotations: make(map[string]string),
					Labels:      make(map[string]string),
				},
			},
			createdOrUpdatedBDName: "blockdevice-123",
			wantErr:                false,
		},
		"existing resource has annotation": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					UUID: "blockdevice-123",
				},
			},
			annotation: map[string]string{
				"key1": "val1",
				"key2": "val2",
			},
			existingBD: &apis.BlockDevice{
				ObjectMeta: metav1.ObjectMeta{
					Name: "blockdevice-123",
					Annotations: map[string]string{
						"keyX": "valX",
					},
					Labels: make(map[string]string),
				},
			},
			createdOrUpdatedBDName: "blockdevice-123",
			wantErr:                false,
		},
		"resource does not exist": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					UUID: "blockdevice-123",
				},
			},
			annotation: map[string]string{
				"key1": "val1",
				"key2": "val2",
			},
			existingBD:             nil,
			createdOrUpdatedBDName: "blockdevice-123",
			wantErr:                false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := scheme.Scheme
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDevice{})
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDeviceList{})
			cl := fake.NewFakeClientWithScheme(s)

			// initialize client with the bd resource
			if tt.existingBD != nil {
				cl.Create(context.TODO(), tt.existingBD)
			}

			ctrl := &controller.Controller{
				Clientset: cl,
			}
			pe := &ProbeEvent{
				Controller: ctrl,
			}
			err := pe.createOrUpdateWithAnnotation(tt.annotation, tt.bd, tt.existingBD)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("createOrUpdateWithAnnotation() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			// check if a BD has been created or updated
			if len(tt.createdOrUpdatedBDName) != 0 {
				gotBDAPI := &apis.BlockDevice{}
				err := cl.Get(context.TODO(), client.ObjectKey{Name: tt.createdOrUpdatedBDName}, gotBDAPI)
				if err != nil {
					t.Errorf("error in getting blockdevice %s: %v", tt.createdOrUpdatedBDName, err)
					return
				}
				// verify the annotation on the resource
				for k, v := range tt.annotation {
					assert.Equal(t, v, gotBDAPI.GetAnnotations()[k])
				}
			}
		})
	}
}
