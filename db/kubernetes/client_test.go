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
