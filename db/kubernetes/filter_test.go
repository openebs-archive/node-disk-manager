package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateLabelFilter(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := map[string]struct {
		args args
		want string
	}{
		"when key is empty": {
			args: args{
				key:   "",
				value: "machine1",
			},
			want: "",
		},
		"when value is empty": {
			args: args{
				key:   "hostname",
				value: "",
			},
			want: "",
		},
		"when both key and value are empty": {
			args: args{
				key:   "",
				value: "",
			},
			want: "",
		},
		"when valid key and value is given": {
			args: args{
				key:   "ndm.io/managed",
				value: "false",
			},
			want: "ndm.io/managed=false",
		},
		"when a valid hostname key is present": {
			args: args{
				key:   "hostname",
				value: "machine1",
			},
			want: "kubernetes.io/hostname=machine1",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := GenerateLabelFilter(test.args.key, test.args.value)
			assert.Equal(t, test.want, got)
		})
	}
}
