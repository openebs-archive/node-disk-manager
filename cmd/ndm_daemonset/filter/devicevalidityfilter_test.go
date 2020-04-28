package filter

import (
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeviceValidityFilterExclude(t *testing.T) {
	tests := map[string]struct {
		blockDevice *blockdevice.BlockDevice
		want        bool
	}{
		"valid BlockDevice": {
			blockDevice: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				Capacity: blockdevice.CapacityInformation{
					Storage: 1024,
				},
			},
			want: true,
		},
		"invalid Path in BlockDevice": {
			blockDevice: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "",
				},
				Capacity: blockdevice.CapacityInformation{
					Storage: 1024,
				},
			},
			want: false,
		},
		"invalid Capacity in BlockDevice": {
			blockDevice: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				Capacity: blockdevice.CapacityInformation{
					Storage: 0,
				},
			},
			want: false,
		},
		"invalid Capacity and DevPath": {
			blockDevice: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "",
				},
				Capacity: blockdevice.CapacityInformation{
					Storage: 0,
				},
			},
			want: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			dvf := deviceValidityFilter{}
			dvf.Start()
			assert.Equal(t, test.want, dvf.Exclude(test.blockDevice))
		})
	}
}

func TestIsValidDevPath(t *testing.T) {
	tests := map[string]struct {
		bd   *blockdevice.BlockDevice
		want bool
	}{
		"valid dev path": {
			bd: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
			},
			want: true,
		},
		"invalid dev path": {
			bd: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "",
				},
			},
			want: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, isValidDevPath(test.bd))
		})
	}
}

func Test_isValidCapacity(t *testing.T) {
	tests := map[string]struct {
		bd   *blockdevice.BlockDevice
		want bool
	}{
		"valid capacity": {
			bd: &blockdevice.BlockDevice{
				Capacity: blockdevice.CapacityInformation{
					Storage: 102400,
				},
			},
			want: true,
		},
		"invalid capacity": {
			bd: &blockdevice.BlockDevice{
				Capacity: blockdevice.CapacityInformation{
					Storage: 0,
				},
			},
			want: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, isValidCapacity(test.bd))
		})
	}
}
