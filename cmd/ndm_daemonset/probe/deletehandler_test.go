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
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/stretchr/testify/assert"
	"testing"
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
