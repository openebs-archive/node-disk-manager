package probe

import (
	"github.com/openebs/node-disk-manager/blockdevice"
	"testing"
)

func Test_customTagProbe_FillBlockDeviceDetails(t *testing.T) {
	type fields struct {
		tags []tag
	}
	type args struct {
		bd *blockdevice.BlockDevice
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctp := &customTagProbe{
				tags: tt.fields.tags,
			}
		})
	}
}
