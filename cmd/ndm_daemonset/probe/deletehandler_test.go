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
	"github.com/openebs/node-disk-manager/pkg/util"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestRemoveBlockDeviceFromHierarchyCache(t *testing.T) {
	tests := map[string]struct {
		cache     blockdevice.Hierarchy
		bd        blockdevice.BlockDevice
		wantCache blockdevice.Hierarchy
		wantOk    bool
	}{
		"device present in cache": {
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
				},
			},
			wantCache: make(blockdevice.Hierarchy),
			wantOk:    true,
		},
		"device not present in cache": {
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
			gotOk := pe.removeBlockDeviceFromHierarchyCache(tt.bd)
			assert.Equal(t, tt.wantCache, pe.Controller.BDHierarchy)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}

func TestDeleteBlockDevice(t *testing.T) {

	fakeWWN := "fake-wwn"
	fakeSerial := "fake-serial"
	fakePartEntry := "fake-part1"
	fakeVendor := "fake-vendor"
	fakePartTable := "fake-part-table"
	fakeFSUUID := "fake-fs-uuid"

	physicalDisk := blockdevice.BlockDevice{
		Identifier: blockdevice.Identifier{
			DevPath: "/dev/sda",
		},
		DeviceAttributes: blockdevice.DeviceAttribute{
			WWN:    fakeWWN,
			Serial: fakeSerial,
			Vendor: fakeVendor,
		},
	}
	physicalDiskPart1 := blockdevice.BlockDevice{
		Identifier: blockdevice.Identifier{
			DevPath: "/dev/sda1",
		},
		DeviceAttributes: blockdevice.DeviceAttribute{
			WWN:        fakeWWN,
			Serial:     fakeSerial,
			DeviceType: blockdevice.BlockDeviceTypePartition,
		},
		PartitionInfo: blockdevice.PartitionInformation{
			PartitionEntryUUID: fakePartEntry,
		},
	}
	physicalDiskUsedByZFSPV := blockdevice.BlockDevice{
		Identifier: blockdevice.Identifier{
			DevPath: "/dev/sda",
		},
		DeviceAttributes: blockdevice.DeviceAttribute{
			WWN:    fakeWWN,
			Serial: fakeSerial,
		},
		PartitionInfo: blockdevice.PartitionInformation{
			PartitionTableUUID: fakePartTable,
		},
	}
	virtualDiskUsedByCstor1 := blockdevice.BlockDevice{
		Identifier: blockdevice.Identifier{
			DevPath: "/dev/sda",
		},
		NodeAttributes: blockdevice.NodeAttribute{
			blockdevice.NodeName: "node1",
		},
		DeviceAttributes: blockdevice.DeviceAttribute{
			Model: "Virtual_disk",
		},
		PartitionInfo: blockdevice.PartitionInformation{
			PartitionTableUUID: fakePartTable,
		},
	}
	virtualDiskUsedByCstor2 := blockdevice.BlockDevice{
		Identifier: blockdevice.Identifier{
			DevPath: "/dev/sdb",
		},
		NodeAttributes: blockdevice.NodeAttribute{
			blockdevice.NodeName: "node1",
		},
		DeviceAttributes: blockdevice.DeviceAttribute{
			Model: "Virtual_disk",
		},
		PartitionInfo: blockdevice.PartitionInformation{
			PartitionTableUUID: fakePartTable,
		},
	}
	virtualDiskUsedByLocalPV1 := blockdevice.BlockDevice{
		Identifier: blockdevice.Identifier{
			DevPath: "/dev/sda",
		},
		NodeAttributes: blockdevice.NodeAttribute{
			blockdevice.NodeName: "node1",
		},
		FSInfo: blockdevice.FileSystemInformation{
			FileSystemUUID: fakeFSUUID,
		},
		DeviceAttributes: blockdevice.DeviceAttribute{
			Model: "Virtual_disk",
		},
	}
	virtualDiskUsedByLocalPV2 := blockdevice.BlockDevice{
		Identifier: blockdevice.Identifier{
			DevPath: "/dev/sdb",
		},
		NodeAttributes: blockdevice.NodeAttribute{
			blockdevice.NodeName: "node1",
		},
		DeviceAttributes: blockdevice.DeviceAttribute{
			Model: "Virtual_disk",
		},
		FSInfo: blockdevice.FileSystemInformation{
			FileSystemUUID: fakeFSUUID,
		},
	}

	fakePhysicalDiskGPTBasedUUID, _ := generateUUID(physicalDisk)
	fakePhysicalDiskGPTBasedUUIDPart1, _ := generateUUID(physicalDiskPart1)
	fakePhysicalDiskLegacyUUID, _ := generateLegacyUUID(physicalDisk)
	fakecstorVirtualDiskLegacyUUID, _ := generateLegacyUUID(virtualDiskUsedByCstor1)
	fakelocalpvVirtualDiskLegacyUUID, _ := generateLegacyUUID(virtualDiskUsedByLocalPV1)
	fakezfspvPhysicalDiskUUID, _ := generateUUIDFromPartitionTable(physicalDiskUsedByZFSPV)

	tests := map[string]struct {
		bd        blockdevice.BlockDevice
		bdAPIList *apis.BlockDeviceList
		// name of the deactivated BDs
		deactivatedBDs []string
		wantErr        bool
	}{
		"Type: disk, physical disk, has one partition": {
			bd: physicalDisk,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: fakePhysicalDiskGPTBasedUUID,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sda",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
							State:      apis.BlockDeviceActive,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: fakePhysicalDiskGPTBasedUUIDPart1,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sda1",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
							State:      apis.BlockDeviceActive,
						},
					},
				},
			},
			deactivatedBDs: []string{fakePhysicalDiskGPTBasedUUID},
			wantErr:        false,
		},
		"Type: partition, physical disk, parent BD resource also present": {
			bd: physicalDiskPart1,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: fakePhysicalDiskGPTBasedUUID,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sda",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
							State:      apis.BlockDeviceActive,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: fakePhysicalDiskGPTBasedUUIDPart1,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sda1",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
							State:      apis.BlockDeviceActive,
						},
					},
				},
			},
			deactivatedBDs: []string{fakePhysicalDiskGPTBasedUUIDPart1},
		},
		"Type: disk, physical disk, no partitions": {
			bd: physicalDisk,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: fakePhysicalDiskGPTBasedUUID,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sda",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
							State:      apis.BlockDeviceActive,
						},
					},
				},
			},
			deactivatedBDs: []string{fakePhysicalDiskGPTBasedUUID},
			wantErr:        false,
		},
		"Type: disk, physical disk, upgraded a claimed BD": {
			bd: physicalDisk,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: fakePhysicalDiskLegacyUUID,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sda",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
							State:      apis.BlockDeviceActive,
						},
					},
				},
			},
			deactivatedBDs: []string{fakePhysicalDiskLegacyUUID},
			wantErr:        false,
		},
		"Type: disk, virtual disk, upgraded a claimed BD used by cstor with no path change": {
			bd: virtualDiskUsedByCstor1,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: fakecstorVirtualDiskLegacyUUID,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sda",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
							State:      apis.BlockDeviceActive,
						},
					},
				},
			},
			deactivatedBDs: []string{fakecstorVirtualDiskLegacyUUID},
			wantErr:        false,
		},
		"Type: disk, virtual disk, upgraded a claimed BD used by cstor with path change": {
			bd: virtualDiskUsedByCstor2,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: fakecstorVirtualDiskLegacyUUID,
							Annotations: map[string]string{
								internalPartitionUUIDAnnotation: fakePartTable,
							},
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdb",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
							State:      apis.BlockDeviceActive,
						},
					},
				},
			},
			deactivatedBDs: []string{fakecstorVirtualDiskLegacyUUID},
			wantErr:        false,
		},
		"Type: disk, virtual disk, upgraded a claimed BD used by localPV with no path change": {
			bd: virtualDiskUsedByLocalPV1,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: fakelocalpvVirtualDiskLegacyUUID,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sda",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
							State:      apis.BlockDeviceActive,
						},
					},
				},
			},
			deactivatedBDs: []string{fakelocalpvVirtualDiskLegacyUUID},
			wantErr:        false,
		},
		"Type: disk, virtual disk, upgraded a claimed BD used by localPV with path change": {
			bd: virtualDiskUsedByLocalPV2,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: fakecstorVirtualDiskLegacyUUID,
							Annotations: map[string]string{
								internalFSUUIDAnnotation: fakeFSUUID,
							},
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sdb",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceClaimed,
							State:      apis.BlockDeviceActive,
						},
					},
				},
			},
			deactivatedBDs: []string{fakelocalpvVirtualDiskLegacyUUID},
			wantErr:        false,
		},
		"Type: disk, physical disk, used by zfs localPV": {
			bd: physicalDiskUsedByZFSPV,
			bdAPIList: &apis.BlockDeviceList{
				Items: []apis.BlockDevice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: fakezfspvPhysicalDiskUUID,
						},
						Spec: apis.DeviceSpec{
							Path: "/dev/sda",
						},
						Status: apis.DeviceStatus{
							ClaimState: apis.BlockDeviceUnclaimed,
							State:      apis.BlockDeviceActive,
						},
					},
				},
			},
			deactivatedBDs: []string{fakezfspvPhysicalDiskUUID},
			wantErr:        false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// pinning variables
			bd := tt.bd
			bdAPIList := tt.bdAPIList
			s := scheme.Scheme
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDevice{})
			s.AddKnownTypes(apis.SchemeGroupVersion, &apis.BlockDeviceList{})
			cl := fake.NewFakeClientWithScheme(s)
			ctrl := &controller.Controller{
				Clientset:   cl,
				BDHierarchy: make(blockdevice.Hierarchy),
			}

			// add the bd to cache so that removing from cache does not error out.
			ctrl.BDHierarchy[bd.DevPath] = bd

			// initialize client with all the bd resources
			for _, bdAPI := range bdAPIList.Items {
				cl.Create(context.TODO(), &bdAPI)
			}

			pe := &ProbeEvent{
				Controller: ctrl,
			}

			if err := pe.deleteBlockDevice(bd, bdAPIList); (err != nil) != tt.wantErr {
				t.Errorf("deleteBlockDevice() error = %v, wantErr %v", err, tt.wantErr)
			}

			gotBDList := &apis.BlockDeviceList{}
			if err := cl.List(context.TODO(), gotBDList); err != nil {
				t.Errorf("List call failed error = %v", err)
			}

			noOfDeactivatedBDs := 0
			for _, gotBDAPI := range gotBDList.Items {
				if util.Contains(tt.deactivatedBDs, gotBDAPI.Name) {
					assert.Equal(t, apis.BlockDeviceInactive, gotBDAPI.Status.State)
					noOfDeactivatedBDs++
				} else {
					assert.Equal(t, apis.BlockDeviceActive, gotBDAPI.Status.State)
				}
			}

			assert.Equal(t, noOfDeactivatedBDs, len(tt.deactivatedBDs))

		})
	}
}
