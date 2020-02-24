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
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/pkg/apis"
	"github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	// NamespaceENV is the name of environment variable to get the namespace
	NamespaceENV = "NAMESPACE"
)

// Client is the wrapper over the k8s client that will be used by
// NDM to interface with etcd
type Client struct {
	// cfg is configuration used to generate the client
	cfg *rest.Config

	// client is the controller-runtime client used to interface with etcd
	client client.Client

	// namespace in which this client is operating
	namespace string
}

// New creates a new client object using the default config
func New() (Client, error) {

	c := Client{}

	// get the kube cfg
	cfg, err := config.GetConfig()
	if err != nil {
		klog.Errorf("error getting cfg. %v", err)
		return c, err
	}

	c.cfg = cfg

	klog.V(2).Info("Client config created.")

	err = c.setNamespace()
	if err != nil {
		klog.Errorf("error setting namespace for client. %v", err)
		return c, err
	}

	klog.V(2).Infof("Namespace \"%s\" set for the client", c.namespace)

	err = c.InitClient()
	if err != nil {
		return c, err
	}
	return c, nil
}

// InitClient sets the client using the config
func (cl *Client) InitClient() error {
	c, err := client.New(cl.cfg, client.Options{})
	if err != nil {
		return err
	}
	cl.client = c
	return nil
}

// SetClient sets the given client
func (cl *Client) SetClient(client2 client.Client) {
	cl.client = client2
}

// RegisterAPI registers the API scheme in the client using the manager.
// This function needs to be called only once a client object
func (cl *Client) RegisterAPI() error {
	mgr, err := manager.New(cl.cfg, manager.Options{Namespace: cl.namespace})
	if err != nil {
		return err
	}

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}
	return nil
}

// setNamespace sets the namespace in which NDM is running
func (cl *Client) setNamespace() error {
	ns, ok := os.LookupEnv(NamespaceENV)
	if !ok {
		return fmt.Errorf("error getting namespace from ENV variable")
	}

	cl.namespace = ns

	return nil
}

// ListBlockDevice lists the block device from etcd based on
// the filters used
func (cl *Client) ListBlockDevice(filters ...string) ([]blockdevice.BlockDevice, error) {
	bdList := &v1alpha1.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	listOptions := &client.ListOptions{Namespace: cl.namespace}

	// apply the filter only if any filters are provided
	if len(filters) != 0 {
		_ = listOptions.SetLabelSelector(strings.Join(filters, ","))
	}

	err := cl.client.List(context.TODO(), listOptions, bdList)
	if err != nil {
		klog.Error("error in listing BDs. ", err)
		return nil, err
	}

	blockDeviceList := make([]blockdevice.BlockDevice, 0)
	err = convertBlockDeviceAPIListToBlockDeviceList(bdList, &blockDeviceList)
	if err != nil {
		return blockDeviceList, err
	}

	klog.V(4).Infof("no of blockdevices listed : %d", len(blockDeviceList))

	return blockDeviceList, nil
}
