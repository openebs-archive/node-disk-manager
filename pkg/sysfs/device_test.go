package sysfs

import (
	"reflect"
	"testing"
)

func TestNewSysFsDeviceFromDevPath(t *testing.T) {
	type args struct {
		devPath string
	}
	tests := []struct {
		name    string
		args    args
		want    *Device
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSysFsDeviceFromDevPath(tt.args.devPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSysFsDeviceFromDevPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSysFsDeviceFromDevPath() got = %v, want %v", got, tt.want)
			}
		})
	}
}
