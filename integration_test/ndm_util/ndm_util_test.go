package ndmutil

import (
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/openebs/CITF/utils/log"
	sysutil "github.com/openebs/CITF/utils/system"
)

func TestGetNDMDir(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Test GetNDMDir",
			want: path.Join(os.Getenv("GOPATH"), "src/github.com/openebs/node-disk-manager/"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNDMDir(); got != tt.want {
				t.Errorf("GetNDMDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNDMBinDir(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Test GetNDMBinDir",
			want: path.Join(os.Getenv("GOPATH"), "src/github.com/openebs/node-disk-manager/bin/"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNDMBinDir(); got != tt.want {
				t.Errorf("GetNDMBinDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNDMTestConfigurationDir(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Test GetNDMTestConfiguration",
			want: "/tmp/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNDMTestConfigurationDir(); got != tt.want {
				t.Errorf("GetNDMTestConfigurationDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNDMTestConfigurationFileName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Test GetNDMTestConfigurationFileName",
			want: "NDM_Test_node-disk-manager.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNDMTestConfigurationFileName(); got != tt.want {
				t.Errorf("GetNDMTestConfigurationFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNDMConfigurationFileName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Test GetNDMConfigurationFileName",
			want: "node-disk-manager.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNDMConfigurationFileName(); got != tt.want {
				t.Errorf("GetNDMConfigurationFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNDMConfigurationFilePath(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Test GetNDMConfigurationFilePath",
			want: path.Join(os.Getenv("GOPATH"), "src/github.com/openebs/node-disk-manager/", "samples", "node-disk-manager.yaml"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNDMConfigurationFilePath(); got != tt.want {
				t.Errorf("GetNDMConfigurationFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNDMOperatorFileName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Test GetNDMOperatorFileName",
			want: "ndm-operator.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNDMOperatorFileName(); got != tt.want {
				t.Errorf("GetNDMOperatorFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNDMOperatorFilePath(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Test GetNDMConfigurationFilePath",
			want: path.Join(os.Getenv("GOPATH"), "src/github.com/openebs/node-disk-manager/", "ndm-operator.yaml"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNDMOperatorFilePath(); got != tt.want {
				t.Errorf("GetNDMOperatorFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNDMTestConfigurationFilePath(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Test GetNDMTestConfigurationFilePath",
			want: "/tmp/NDM_Test_node-disk-manager.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNDMTestConfigurationFilePath(); got != tt.want {
				t.Errorf("GetNDMTestConfigurationFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDockerImageName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Test GetDockerImageName",
			want: "openebs/node-disk-manager-" + sysutil.GetenvFallback("XC_ARCH", runtime.GOARCH),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDockerImageName(); got != tt.want {
				t.Errorf("GetDockerImageName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func SetTagEnv() {
	tag, err := sysutil.ExecCommand("git describe --tags --always")
	if err != nil {
		tag = ""
	}
	os.Setenv("TAG", tag)
}

func TestGetDockerImageTag(t *testing.T) {
	var tag string
	var tagPresent bool
	var Path string
	wantTag, _ := sysutil.ExecCommand("git describe --tags --always")
	tests := []struct {
		name       string
		want       string
		beforefunc func()
		afterfunc  func()
	}{
		{
			name: "After setting TAG environment variable",
			want: "abcdef",
			beforefunc: func() {
				tag, tagPresent = os.LookupEnv("TAG")
				os.Setenv("TAG", "abcdef")
			},
			afterfunc: func() {
				if !tagPresent {
					os.Unsetenv("TAG")
				} else {
					os.Setenv("TAG", tag)
				}
			},
		},
		{
			name:       "When TAG environment variable is not present and ExecCommand works fine",
			want:       strings.TrimSpace(wantTag),
			beforefunc: func() {},
			afterfunc:  func() {},
		},
		{
			name: "When TAG environment variable is not present and ExecCommand gives error",
			want: "",
			beforefunc: func() {
				tag, tagPresent = os.LookupEnv("TAG")
				Path, _ = os.Getwd()
				os.Chdir(os.Getenv("HOME"))
				if tagPresent {
					os.Unsetenv("TAG")
				}
			},
			afterfunc: func() {
				if tagPresent {
					os.Setenv("TAG", tag)
				}
				os.Chdir(Path)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beforefunc()
			if got := GetDockerImageTag(); got != tt.want {
				t.Errorf("GetDockerImageTag() = %v, want %v", got, tt.want)
			}
			tt.afterfunc()
		})
	}
}

func TestGetNDMNamespace(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Test GetNDMNamespace",
			want: NdmNamespace,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNDMNamespace(); got != tt.want {
				t.Errorf("GetNDMNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateNdmLog(t *testing.T) {
	type args struct {
		log string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "when container has `started the controller` in log",
			args: args{
				log: "started the controller",
			},
			want: true,
		},
		{
			name: "when container doesn't have `started the controller` in log",
			args: args{
				log: "controller not started",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateNdmLog(tt.args.log); got != tt.want {
				t.Errorf("ValidateNdmLog() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYAMLPrepare(t *testing.T) {
	type requiredFields struct {
		ImageName       string
		ImagePullPolicy string
	}
	tests := []struct {
		name    string
		want    requiredFields
		wantErr bool
	}{
		{
			name: "Test YAMLPrepare",
			want: requiredFields{
				ImageName:       GetDockerImageName() + ":" + GetDockerImageTag(),
				ImagePullPolicy: "IfNotPresent",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := YAMLPrepare()
			if (err != nil) != tt.wantErr {
				t.Errorf("YAMLPrepare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotRequiredFields := requiredFields{
				ImageName:       got.Spec.Template.Spec.Containers[0].Image,
				ImagePullPolicy: string(got.Spec.Template.Spec.Containers[0].ImagePullPolicy),
			}
			if !reflect.DeepEqual(gotRequiredFields, tt.want) {
				t.Errorf("YAMLPrepare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYAMLPrepareAndWriteInTempPath(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Test YAMLPrepareAndWriteInTempPath",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := YAMLPrepareAndWriteInTempPath(); (err != nil) != tt.wantErr {
				t.Errorf("YAMLPrepareAndWriteInTempPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReplaceImageInYAMLAndApply(t *testing.T) {
	logDebugEnabled := log.DebugEnabled
	log.DebugEnabled = true
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Test ReplaceImageInYAMLAndApply",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ReplaceImageInYAMLAndApply(); (err != nil) != tt.wantErr {
				t.Errorf("ReplaceImageInYAMLAndApply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	log.DebugEnabled = logDebugEnabled
}
