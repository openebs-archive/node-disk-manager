/*
Copyright 2019 The OpenEBS Authors

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

package kubernetes

import (
	"errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		want interface{}
		err  error
	}{
		"when key is empty": {
			args: args{
				key:   "",
				value: "machine1",
			},
			want: nil,
			err:  errors.New("key/value is empty for label filter"),
		},
		"when value is empty": {
			args: args{
				key:   "hostname",
				value: "",
			},
			want: nil,
			err:  errors.New("key/value is empty for label filter"),
		},
		"when both key and value are empty": {
			args: args{
				key:   "",
				value: "",
			},
			want: nil,
			err:  errors.New("key/value is empty for label filter"),
		},
		"when valid key and value is given": {
			args: args{
				key:   "ndm.io/managed",
				value: "false",
			},
			want: client.MatchingLabels{"ndm.io/managed": "false"},
			err:  nil,
		},
		"when a valid hostname key is present": {
			args: args{
				key:   "hostname",
				value: "machine1",
			},
			want: client.MatchingLabels{"kubernetes.io/hostname": "machine1"},
			err:  nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := GenerateLabelFilter(test.args.key, test.args.value)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.err, err)
		})
	}
}
