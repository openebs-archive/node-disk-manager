package blockdevice

import (
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestMatchResourceRequirements(t *testing.T) {
	blockDevice := &apis.BlockDevice{
		TypeMeta: metav1.TypeMeta{
			Kind:       controller.NDMBlockDeviceKind,
			APIVersion: controller.NDMVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-block-device",
			Namespace: "",
		},
		Spec: apis.DeviceSpec{
			Capacity: apis.DeviceCapacity{
				Storage: uint64(1024000),
			},
		},
	}
	tests := map[string]struct {
		blockDevice *apis.BlockDevice
		rList       v1.ResourceList
		expected    bool
	}{
		"block device capacity greater than requested capacity": {blockDevice,
			v1.ResourceList{apis.ResourceCapacity: resource.MustParse("1024000")},
			true},
		"block device capacity is less than requested capacity": {blockDevice,
			v1.ResourceList{apis.ResourceCapacity: resource.MustParse("404800000")},
			false},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, matchResourceRequirements(*test.blockDevice, test.rList))
		})
	}
}
