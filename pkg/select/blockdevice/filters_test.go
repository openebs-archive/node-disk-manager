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

package blockdevice

import (
	"fmt"

	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
	"github.com/openebs/node-disk-manager/db/kubernetes"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

const (
	TestNoOfBDs = 6
)

type BDLabel map[string]string
type BDLabelList []BDLabel

func TestFilterBlockDeviceTag(t *testing.T) {

	// label list with no additional labels
	bdLabelList1 := []BDLabel{
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host1",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host2",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host3",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host4",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host5",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host6",
		},
	}

	// label list with all devices having same tag
	bdLabelList2 := []BDLabel{
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host1",
			kubernetes.BlockDeviceTagLabel:     "X",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host2",
			kubernetes.BlockDeviceTagLabel:     "X",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host3",
			kubernetes.BlockDeviceTagLabel:     "X",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host4",
			kubernetes.BlockDeviceTagLabel:     "X",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host5",
			kubernetes.BlockDeviceTagLabel:     "X",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host6",
			kubernetes.BlockDeviceTagLabel:     "X",
		},
	}

	// label list with some devices having default label and some devices
	// with device tag
	bdLabelList3 := []BDLabel{
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host1",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host2",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host3",
			kubernetes.BlockDeviceTagLabel:     "X",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host4",
			kubernetes.BlockDeviceTagLabel:     "X",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host5",
			kubernetes.BlockDeviceTagLabel:     "Y",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host6",
			kubernetes.BlockDeviceTagLabel:     "Y",
		},
	}

	// label list with some devices having default label and some devices
	// with device tag, and some devices with empty tag
	bdLabelList4 := []BDLabel{
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host1",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host2",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host3",
			kubernetes.BlockDeviceTagLabel:     "X",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host4",
			kubernetes.BlockDeviceTagLabel:     "X",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host5",
			kubernetes.BlockDeviceTagLabel:     "",
		},
		map[string]string{
			kubernetes.KubernetesHostNameLabel: "host6",
			kubernetes.BlockDeviceTagLabel:     "",
		},
	}

	type args struct {
		bdLabelList BDLabelList
		spec        *apis.DeviceClaimSpec
	}
	tests := map[string]struct {
		args          args
		wantedNoofBDs int
	}{
		"no labels on any BD and no selector on BDC": {
			args: args{
				bdLabelList: bdLabelList1,
				spec:        &apis.DeviceClaimSpec{},
			},
			wantedNoofBDs: 6,
		},
		"all BDs have same device tag label and no selector": {
			args: args{
				bdLabelList: bdLabelList2,
				spec:        &apis.DeviceClaimSpec{},
			},
			wantedNoofBDs: 0,
		},
		"all BDs have same device tag label and selector for tag": {
			args: args{
				bdLabelList: bdLabelList2,
				spec: &apis.DeviceClaimSpec{
					Selector: &v1.LabelSelector{
						MatchLabels: map[string]string{kubernetes.BlockDeviceTagLabel: "X"},
					},
				},
			},
			wantedNoofBDs: 6,
		},
		"all BDs have same device tag label and custom label used in selector": {
			args: args{
				bdLabelList: bdLabelList2,
				spec: &apis.DeviceClaimSpec{
					Selector: &v1.LabelSelector{
						MatchLabels: map[string]string{"ndm.io/test": "test"},
					},
				},
			},
			wantedNoofBDs: 0,
		},
		"some BDs with tag and some without tag, combined with no selector": {
			args: args{
				bdLabelList: bdLabelList3,
				spec: &apis.DeviceClaimSpec{
					Selector: &v1.LabelSelector{},
				},
			},
			wantedNoofBDs: 2,
		},
		"some BDs with tag key, but with empty selector": {
			args: args{
				bdLabelList: bdLabelList4,
				spec: &apis.DeviceClaimSpec{
					Selector: &v1.LabelSelector{},
				},
			},
			wantedNoofBDs: 2,
		},
		"some BDs with tag key, with selector matching empty tag": {
			args: args{
				bdLabelList: bdLabelList4,
				spec: &apis.DeviceClaimSpec{
					Selector: &v1.LabelSelector{
						MatchLabels: map[string]string{kubernetes.BlockDeviceTagLabel: ""},
					},
				},
			},
			wantedNoofBDs: 4,
		},
		"some BDs with tag key, with selector matching tag exists operation": {
			args: args{
				bdLabelList: bdLabelList4,
				spec: &apis.DeviceClaimSpec{
					Selector: &v1.LabelSelector{
						MatchExpressions: []v1.LabelSelectorRequirement{
							{
								Key:      kubernetes.BlockDeviceTagLabel,
								Operator: v1.LabelSelectorOpExists,
							},
						},
					},
				},
			},
			wantedNoofBDs: 4,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bdLabelList := test.args.bdLabelList
			spec := test.args.spec
			wantedNoOfBDs := test.wantedNoofBDs
			originalBDList := createFakeBlockDeviceList(bdLabelList, TestNoOfBDs)
			got := filterBlockDeviceTag(originalBDList, spec)
			assert.Equal(t, wantedNoOfBDs, len(got.Items))
		})
	}
}

func createFakeBlockDeviceList(labelList BDLabelList, noOfBDs int) *apis.BlockDeviceList {
	bdListAPI := &apis.BlockDeviceList{
		TypeMeta: v1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
		Items: []apis.BlockDevice{},
	}
	for i := 0; i < noOfBDs; i++ {
		bdName := fmt.Sprint("bd", i)
		bdListAPI.Items = append(bdListAPI.Items, createFakeBlockDevice(bdName, labelList[i]))
	}
	return bdListAPI
}

func createFakeBlockDevice(name string, label map[string]string) apis.BlockDevice {
	bdAPI := apis.BlockDevice{
		TypeMeta: v1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}
	bdAPI.Name = name
	bdAPI.Labels = label
	return bdAPI
}
