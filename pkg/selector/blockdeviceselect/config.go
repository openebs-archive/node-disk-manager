package blockdeviceselect

import (
	"github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	Client          client.Client
	ClaimSpec       *v1alpha1.DeviceClaimSpec
	ManualSelection bool
}

func NewConfig(claimSpec *v1alpha1.DeviceClaimSpec, client client.Client) *Config {
	isManualSelection := false
	if claimSpec.BlockDeviceName != "" {
		isManualSelection = true
	}
	c := &Config{
		Client:          client,
		ClaimSpec:       claimSpec,
		ManualSelection: isManualSelection,
	}
	return c
}
