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

package v041_042

import (
	"context"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// UpgradeTask is the struct which implements the Task interface
// which can be used to perform the upgrade
type UpgradeTask struct {
	from   string
	to     string
	client client.Client
	err    error
}

// NewUpgradeTask creates a new upgrade task with given client
// and specified `from` and `to` version
func NewUpgradeTask(from, to string, c client.Client) *UpgradeTask {
	return &UpgradeTask{from: from, to: to, client: c}
}

// FromVersion returns the version from which the components need to be updated
func (p *UpgradeTask) FromVersion() string {
	return p.from
}

// ToVersion returns the version to which components will be updated. This should be
// the current version
func (p *UpgradeTask) ToVersion() string {
	return p.to
}

// PreUpgrade runs the preupgrade tasks and returns whether it succeeded or not
func (p *UpgradeTask) PreUpgrade() bool {
	var err error
	bdcList := &apis.BlockDeviceClaimList{}
	opts := &client.ListOptions{}
	err = p.client.List(context.TODO(), opts, bdcList)
	if err != nil {
		p.err = err
		return false
	}

	for _, bdc := range bdcList.Items {
		err = p.copyHostName(&bdc)
		if err != nil {
			p.err = err
			return false
		}
	}
	return true
}

// Upgrade runs the main upgrade tasks and returns whether it succeeded or not
func (p *UpgradeTask) Upgrade() bool {
	if p.err != nil {
		return false
	}
	return true
}

// PostUpgrade runs the tasks that need to be performed after upgrade and returns
// whether the tasks where success or not
func (p *UpgradeTask) PostUpgrade() bool {
	if p.err != nil {
		return false
	}
	return true
}

// IsSuccess returns error if the upgrade failed, at any step. Else nil will
// be returned
func (p *UpgradeTask) IsSuccess() error {
	return p.err
}

// copyHostName will copy the hostname string from .spec.hostName to
// .spec.nodeAttributes.hostName
func (p *UpgradeTask) copyHostName(claim *apis.BlockDeviceClaim) error {
	// copy the value only if .spec.hostName is non-empty and nodeAttributes.hostName
	// is empty. If hostname field is empty, it is not required to copy the value, as
	// it may be a new BDC.
	if len(claim.Spec.HostName) != 0 &&
		len(claim.Spec.BlockDeviceNodeAttributes.HostName) == 0 {
		claim.Spec.BlockDeviceNodeAttributes.HostName = claim.Spec.HostName
		return p.client.Update(context.TODO(), claim)
	}
	return nil
}
