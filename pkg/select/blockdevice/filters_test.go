package blockdevice

import (
	"fmt"
	"testing"

	controller "github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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
			controller.KubernetesHostNameLabel: "host1",
		},
		map[string]string{
			controller.KubernetesHostNameLabel: "host2",
		},
		map[string]string{
			controller.KubernetesHostNameLabel: "host3",
		},
		map[string]string{
			controller.KubernetesHostNameLabel: "host4",
		},
		map[string]string{
			controller.KubernetesHostNameLabel: "host5",
		},
		map[string]string{
			controller.KubernetesHostNameLabel: "host6",
		},
	}

	// label list with all devices having same tag
	bdLabelList2 := []BDLabel{
		map[string]string{
			controller.KubernetesHostNameLabel: "host1",
			controller.BlockDeviceTagLabel:     "X",
		},
		map[string]string{
			controller.KubernetesHostNameLabel: "host2",
			controller.BlockDeviceTagLabel:     "X",
		},
		map[string]string{
			controller.KubernetesHostNameLabel: "host3",
			controller.BlockDeviceTagLabel:     "X",
		},
		map[string]string{
			controller.KubernetesHostNameLabel: "host4",
			controller.BlockDeviceTagLabel:     "X",
		},
		map[string]string{
			controller.KubernetesHostNameLabel: "host5",
			controller.BlockDeviceTagLabel:     "X",
		},
		map[string]string{
			controller.KubernetesHostNameLabel: "host6",
			controller.BlockDeviceTagLabel:     "X",
		},
	}

	// label list with some devices having default label and some devices
	// with device tag
	bdLabelList3 := []BDLabel{
		map[string]string{
			controller.KubernetesHostNameLabel: "host1",
		},
		map[string]string{
			controller.KubernetesHostNameLabel: "host2",
		},
		map[string]string{
			controller.KubernetesHostNameLabel: "host3",
			controller.BlockDeviceTagLabel:     "X",
		},
		map[string]string{
			controller.KubernetesHostNameLabel: "host4",
			controller.BlockDeviceTagLabel:     "X",
		},
		map[string]string{
			controller.KubernetesHostNameLabel: "host5",
			controller.BlockDeviceTagLabel:     "Y",
		},
		map[string]string{
			controller.KubernetesHostNameLabel: "host6",
			controller.BlockDeviceTagLabel:     "Y",
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
						MatchLabels: map[string]string{controller.BlockDeviceTagLabel: "X"},
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
