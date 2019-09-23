package kubernetes

import (
	"context"
	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/pkg/apis"
	"github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Client is the wrapper over the k8s client that will be used by
// NDM to interface with etcd
type Client struct {
	cfg    *rest.Config
	client client.Client
}

// New creates a new client object using the given config
func New(config *rest.Config) (Client, error) {
	c := Client{
		cfg: config,
	}
	err := c.Set()
	if err != nil {
		return c, err
	}
	return c, nil
}

// Set sets the client using the config
func (cl *Client) Set() error {
	c, err := client.New(cl.cfg, client.Options{})
	if err != nil {
		return err
	}
	cl.client = c
	return nil
}

// RegisterAPI registers the API scheme in the client using the manager.
// This function needs to be called only once a client object
func (cl *Client) RegisterAPI() error {
	mgr, err := manager.New(cl.cfg, manager.Options{})
	if err != nil {
		return err
	}

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}
	return nil
}

// ListBlockDevice lists the block device from etcd based on
// the filters used
func (cl *Client) ListBlockDevice(filters ...string) ([]v1alpha1.BlockDevice, error) {
	bdList := &v1alpha1.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	listOptions := &client.ListOptions{}

	for _, filter := range filters {
		listOptions.SetLabelSelector(filter)
	}

	err := cl.client.List(context.TODO(), listOptions, bdList)
	if err != nil {
		glog.Error("error in listing BDs. ", err)
		return nil, err
	}
	return bdList.Items, nil
}
