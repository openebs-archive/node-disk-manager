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

package blockdevice

import (
	"github.com/openebs/node-disk-manager/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Config stores the configuration for selecting a block device from a
// block device claim. It contains the claim spec, selection type and
// client to interface with etcd
type Config struct {
	Client          client.Client
	ClaimSpec       *v1alpha1.DeviceClaimSpec
	ManualSelection bool
}

// NewConfig creates a new Config struct for the block device claim
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
