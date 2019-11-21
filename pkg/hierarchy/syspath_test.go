package hierarchy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getParent(t *testing.T) {
	type fields struct {
		DeviceName string
		SysPath    string
	}
	tests := map[string]struct {
		fields           fields
		wantedDeviceName string
		wantOk           bool
	}{
		"[block] given block device is a parent": {
			fields: fields{
				DeviceName: "sda",
				SysPath:    "/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda",
			},
			wantedDeviceName: "",
			wantOk:           false,
		},
		"[block] given blockdevice is a partition": {
			fields: fields{
				DeviceName: "sda4",
				SysPath:    "/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/sda4",
			},
			wantedDeviceName: "sda",
			wantOk:           true,
		},
		"[nvme] given blockdevice is a parent": {
			fields: fields{
				DeviceName: "nvme0n1",
				SysPath:    "/sys/devices/pci0000:00/0000:00:0e.0/nvme/nvme0/nvme0n1",
			},
			wantedDeviceName: "",
			wantOk:           false,
		},
		"[nvme] given blockdevice is a partition": {
			fields: fields{
				DeviceName: "nvme0n1p1",
				SysPath:    "/sys/devices/pci0000:00/0000:00:0e.0/nvme/nvme0/nvme0n1/nvme0n1p1",
			},
			wantedDeviceName: "nvme0n1",
			wantOk:           true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			s := deviceSysPath{
				DeviceName: test.fields.DeviceName,
				SysPath:    test.fields.SysPath,
			}
			gotDeviceName, gotOk := s.getParent()
			assert.Equal(t, test.wantedDeviceName, gotDeviceName)
			assert.Equal(t, test.wantOk, gotOk)
		})
	}
}
