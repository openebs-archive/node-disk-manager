package preupgrade

import (
	"context"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	oldBDCFinalizer = "blockdeviceclaim.finalizer"
	newBDCFinalizer = "openebs.io/bdc-protection"
)

type PreUpgrade struct {
	from   string
	to     string
	client client.Client
	err    error
}

func NewPreUpgradeTask(from, to string, c client.Client) *PreUpgrade {
	return &PreUpgrade{from: from, to: to, client: c}
}

func (p *PreUpgrade) FromVersion() string {
	return p.from
}

func (p *PreUpgrade) ToVersion() string {
	return p.to
}

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

func (p *PreUpgrade) Upgrade() bool {
	if p.err != nil {
		return false
	}
	return true
}

func (p *PreUpgrade) PostUpgrade() bool {
	if p.err != nil {
		return false
	}
	return true
}

func (p *PreUpgrade) IsSuccess() error {
	return p.err
}

func (p *PreUpgrade) renameFinalizer(claim *apis.BlockDeviceClaim) error {
	if util.Contains(claim.Finalizers, oldBDCFinalizer) {
		claim.Finalizers = util.RemoveString(claim.Finalizers, oldBDCFinalizer)
		claim.Finalizers = append(claim.Finalizers, newBDCFinalizer)
		return p.client.Update(context.TODO(), claim)
	}
	return nil
}
