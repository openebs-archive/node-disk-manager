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

package preupgrade

import (
	"context"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// oldBDCFinalizer is the old string from which BDC should be updated
	oldBDCFinalizer = "blockdeviceclaim.finalizer"
	// newBDCFinalizer is the new string to which BDC to be updated
	newBDCFinalizer = "openebs.io/bdc-protection"
)

// PreUpgrade is the struct which implements the UpgradeTask interface
// which can be used to perform the upgrade
type PreUpgrade struct {
	from   string
	to     string
	client client.Client
	err    error
}

// NewPreUpgradeTask creates a new preupgrade with given client
// and specified `from` and `to` version
func NewPreUpgradeTask(from, to string, c client.Client) *PreUpgrade {
	return &PreUpgrade{from: from, to: to, client: c}
}

// FromVersion returns the version from which the components need to be updated
func (p *PreUpgrade) FromVersion() string {
	return p.from
}

// ToVersion returns the version to which components will be updated. This should be
// the current version
func (p *PreUpgrade) ToVersion() string {
	return p.to
}

// PreUpgrade runs the preupgrade tasks and returns whether it succeeded or not
func (p *PreUpgrade) PreUpgrade() bool {
	var err error
	bdcList := &apis.BlockDeviceClaimList{}
	opts := &client.ListOptions{}
	err = p.client.List(context.TODO(), opts, bdcList)
	if err != nil {
		p.err = err
		return false
	}

	for _, bdc := range bdcList.Items {
		err = p.renameFinalizer(&bdc)
		if err != nil {
			p.err = err
			return false
		}
	}
	return true
}

// Upgrade runs the main upgrade tasks and returns whether it succeeded or not
func (p *PreUpgrade) Upgrade() bool {
	if p.err != nil {
		return false
	}
	return true
}

// PostUpgrade runs the tasks that need to be performed after upgrade and returns
// whether the tasks where success or not
func (p *PreUpgrade) PostUpgrade() bool {
	if p.err != nil {
		return false
	}
	return true
}

// IsSuccess returns error if the upgrade failed, at any step. Else nil will
// be returned
func (p *PreUpgrade) IsSuccess() error {
	return p.err
}

// renameFinalizer renames the finalizer from old to new in BDC
func (p *PreUpgrade) renameFinalizer(claim *apis.BlockDeviceClaim) error {
	if util.Contains(claim.Finalizers, oldBDCFinalizer) {
		claim.Finalizers = util.RemoveString(claim.Finalizers, oldBDCFinalizer)
		claim.Finalizers = append(claim.Finalizers, newBDCFinalizer)
		return p.client.Update(context.TODO(), claim)
	}
	return nil
}
