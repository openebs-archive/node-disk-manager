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
	"os"
	"reflect"
	"testing"

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestClientListBlockDevice(t *testing.T) {
	type fields struct {
		cfg       *rest.Config
		client    client.Client
		namespace string
	}
	type args struct {
		filters []string
	}
	tests := map[string]struct {
		fields  fields
		args    args
		want    []blockdevice.BlockDevice
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cl := &Client{
				cfg:       test.fields.cfg,
				client:    test.fields.client,
				namespace: test.fields.namespace,
			}
			got, err := cl.ListBlockDevice(test.args.filters...)
			if (err != nil) != test.wantErr {
				t.Errorf("ListBlockDevice() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("ListBlockDevice() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestClientsetNamespace(t *testing.T) {

	fakeNamespace := "openebs"
	cl := &Client{}

	// setting namespace in client when env is not available
	err1 := cl.setNamespace()
	ns1 := cl.namespace

	//setting namespace environment variable
	_ = os.Setenv(NamespaceENV, fakeNamespace)

	// setting namespace in client when env is available
	err2 := cl.setNamespace()
	ns2 := cl.namespace

	tests := map[string]struct {
		wantNamespace string
		gotNamespace  string
		wantErr       bool
		gotErr        error
	}{
		"when NAMESPACE env is not available": {
			ns1,
			"",
			true,
			err1,
		},
		"when NAMESPACE env is available": {
			ns2,
			fakeNamespace,
			false,
			err2,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.wantNamespace, test.gotNamespace)
			assert.Equal(t, test.wantErr, test.gotErr != nil)
		})
	}
}
