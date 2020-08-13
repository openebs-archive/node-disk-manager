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
	"testing"

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/db/kubernetes"

	"github.com/stretchr/testify/assert"
)

func TestCustomTagProbeFillBlockDeviceDetails(t *testing.T) {
	tests := map[string]struct {
		bd             *blockdevice.BlockDevice
		customTags     []tag
		wantTagLabel   string
		wantTagLabelOk bool
	}{
		"no custom tags are provided": {
			bd: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
			},
			customTags:     nil,
			wantTagLabel:   "",
			wantTagLabelOk: false,
		},
		"single custom tag using path is present with matching device": {
			bd: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
			},
			customTags: []tag{
				{
					tagType: tagTypePath,
					regex:   "/dev/sda",
					label:   "label1",
				},
			},
			wantTagLabel:   "label1",
			wantTagLabelOk: true,
		},
		"single custom tag using path is present without matching device": {
			bd: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sdb",
				},
			},
			customTags: []tag{
				{
					tagType: tagTypePath,
					regex:   "/dev/sda",
					label:   "label1",
				},
			},
			wantTagLabel:   "",
			wantTagLabelOk: false,
		},
		"single custom tag with regex using path is present with matching device": {
			bd: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sdb",
				},
			},
			customTags: []tag{
				{
					tagType: tagTypePath,
					regex:   "/dev/sd[a|b]",
					label:   "label1",
				},
			},
			wantTagLabel:   "label1",
			wantTagLabelOk: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			bd := tt.bd
			ctp := &customTagProbe{
				tags: tt.customTags,
			}
			ctp.FillBlockDeviceDetails(bd)

			tagValue, ok := bd.Labels[kubernetes.BlockDeviceTagLabel]
			assert.Equal(t, tt.wantTagLabelOk, ok)

			if tt.wantTagLabelOk {
				assert.Equal(t, tt.wantTagLabel, tagValue)
			}

		})
	}
}
