package kubernetes

import (
	"github.com/openebs/node-disk-manager/pkg/apis"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

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
