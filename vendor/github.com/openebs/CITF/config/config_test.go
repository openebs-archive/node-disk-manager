/*
Copyright 2018 The OpenEBS Authors.
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

package config

import (
	"os"
	"testing"

	"github.com/openebs/CITF/utils/log"
)

var logger log.Logger

// CreateFile creates yaml file for test purpose
func CreateFile() {
	fileData1 := `
environment: minikube
`
	f, err := os.Create("./test-config.yaml")

	logger.LogError(err, "unable to create config file")

	f.WriteString(fileData1)

	// Create yaml file with bad indentation
	fileData2 := `
	environment: minikube
	`
	f, err = os.Create("./test-bad-config.yaml")
	logger.LogError(err, "unable to create bad config file")

	f.WriteString(fileData2)
}

// DeleteFile deletes yaml file
func DeleteFile() {
	err := os.Remove("./test-config.yaml")
	logger.LogError(err, "unable to delete config file")

	err = os.Remove("./test-bad-config.yaml")
	logger.LogError(err, "unable to delete bad config file")
}

func TestLoadConf(t *testing.T) {
	CreateFile()
	type args struct {
		confFilePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "LoadConfFileNotPresent",
			args: args{
				confFilePath: "./file.yaml",
			},
			wantErr: true,
		},
		{
			name: "LoadConfSuccess",
			args: args{
				confFilePath: "./test-config.yaml",
			},
			wantErr: false,
		},
		{
			name: "LoadConfEmptyFileName",
			args: args{
				confFilePath: "",
			},
			wantErr: false,
		},
		{
			name: "LoadConfBadYamlFile",
			args: args{
				confFilePath: "./test-bad-config.yaml",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoadConf(tt.args.confFilePath); (err != nil) != tt.wantErr {
				t.Errorf("LoadConf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	DeleteFile()
}

func Test_getConfValueByStringField(t *testing.T) {
	type args struct {
		conf  Configuration
		field string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "KeyPresentInYaml",
			args: args{
				conf: Configuration{
					Environment: "minikube",
				},
				field: "Environment",
			},
			want: "minikube",
		},
		{
			name: "KeyNotPresentInYaml",
			args: args{
				conf: Configuration{
					Environment: "minikube",
				},
				field: "environment",
			},
			want: "<invalid Value>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getConfValueByStringField(tt.args.conf, tt.args.field); got != tt.want {
				t.Errorf("getConfValueByStringField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDefaultValueByStringField(t *testing.T) {
	type args struct {
		field string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "FieldPresentInYaml",
			args: args{
				field: "Environment",
			},
			want: "minikube",
		},
		{
			name: "FieldNotPresentInYaml",
			args: args{
				field: "environment",
			},
			want: "<invalid Value>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDefaultValueByStringField(tt.args.field); got != tt.want {
				t.Errorf("GetDefaultValueByStringField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUserConfValueByStringField(t *testing.T) {
	type args struct {
		field string
	}
	// Set Value of Conf
	Conf = Configuration{
		Environment: "minikube",
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "FieldPresentInConf",
			args: args{
				field: "Environment",
			},
			want: "minikube",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUserConfValueByStringField(tt.args.field); got != tt.want {
				t.Errorf("GetUserConfValueByStringField() = %v, want %v", got, tt.want)
			}
		})
	}
	Conf = Configuration{} // Reset value of Conf to default (being a global guy)
}

func TestGetConf(t *testing.T) {
	os.Setenv("CITF_CONF_ENVIRONMENT", "minikube")

	Conf = Configuration{
		Environment: "Dear minikube",
	}

	type args struct {
		field string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		cleanup func()
	}{
		{
			name: "GetConfWithEnvSuccess",
			args: args{
				field: "Environment",
			},
			want: "minikube",
			cleanup: func() {
				os.Unsetenv("CITF_CONF_ENVIRONMENT")
			},
		},
		{
			name: "GetConfWithConfValueSuccess",
			args: args{
				field: "Environment",
			},
			want: "Dear minikube",
			cleanup: func() {
				Conf = Configuration{
					Environment: "",
				}
			},
		},
		{
			name: "GetConfWithDefaultConfValueSuccess",
			args: args{
				field: "Environment",
			},
			want:    "minikube",
			cleanup: func() {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetConf(tt.args.field); got != tt.want {
				t.Errorf("GetConf() = %v, want %v", got, tt.want)
			}
		})
		tt.cleanup()
	}
}

func TestEnvironment(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Environment",
			want: "minikube",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Environment(); got != tt.want {
				t.Errorf("Environment() = %v, want %v", got, tt.want)
			}
		})
	}
}
