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

package sysfs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddDevPrefix(t *testing.T) {
	tests := map[string]struct {
		devPaths []string
		want     []string
	}{
		"single device name given": {
			devPaths: []string{"sda"},
			want:     []string{"/dev/sda"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := addDevPrefix(tt.devPaths)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReadSysFSFileAsString(t *testing.T) {
	tests := map[string]struct {
		path        string
		fileName    string
		fileContent string
		want        string
		wantErr     bool
	}{
		"valid sysfs path for dm uuid": {
			path:        "/tmp/dm-0/dm/",
			fileName:    "uuid",
			fileContent: "LVM-OSlVs5gIXuqSKVPukc2aGPh0AeJw31TJqYIRuRHoodYg9Jwkmyvvk0QNYK4YulHt",
			want:        "LVM-OSlVs5gIXuqSKVPukc2aGPh0AeJw31TJqYIRuRHoodYg9Jwkmyvvk0QNYK4YulHt",
			wantErr:     false,
		},
		"valid sysfs path with tailing new line": {
			path:        "/tmp/dm-0/dm/",
			fileName:    "uuid",
			fileContent: "LVM-OSlVs5gIXuqSKVPukc2aGPh0AeJw31TJqYIRuRHoodYg9Jwkmyvvk0QNYK4YulHt\n",
			want:        "LVM-OSlVs5gIXuqSKVPukc2aGPh0AeJw31TJqYIRuRHoodYg9Jwkmyvvk0QNYK4YulHt",
			wantErr:     false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			filePath := tt.path + tt.fileName
			os.MkdirAll(tt.path, 0700)
			file, err := os.Create(filePath)
			if err != nil {
				t.Fatalf("unable to write to file %s, %v", filePath, err)
				return
			}
			file.Write([]byte(tt.fileContent))
			file.Close()
			got, err := readSysFSFileAsString(filePath)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("readSysFSFileAsString() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReadSysFSFileAsInt64(t *testing.T) {
	tests := map[string]struct {
		path        string
		fileName    string
		fileContent string
		want        int64
		wantErr     bool
	}{
		"valid no of block sizes for device size": {
			path:        "/tmp/sda/queue",
			fileName:    "hw_sector_size",
			fileContent: "512",
			want:        512,
			wantErr:     false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			filePath := tt.path + tt.fileName
			os.MkdirAll(tt.path, 0700)
			file, err := os.Create(filePath)
			if err != nil {
				t.Fatalf("unable to write to file %s, %v", filePath, err)
				return
			}
			file.Write([]byte(tt.fileContent))
			file.Close()
			got, err := readSysFSFileAsInt64(filePath)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("readSysFSFileAsInt64() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
