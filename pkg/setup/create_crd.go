package setup

import (
	"fmt"
	"github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// createDiskCRD creates a Disk CRD
func (sc Config) createDiskCRD() error {
	return sc.createCRD(v1alpha1.DiskCRD())
}

// createBlockDeviceCRD creates a BlockDevice CRD
func (sc Config) createBlockDeviceCRD() error {
	return sc.createCRD(v1alpha1.BlockDeviceCRD())
}

// createBlockDeviceClaimCRD creates a BlockDeviceClaim CRD
func (sc Config) createBlockDeviceClaimCRD() error {
	return sc.createCRD(v1alpha1.BlockDeviceClaimCRD())
}

// createCRD creates a CRD in the cluster and waits for it to get into active state
// It will return error, if the CRD creation failed, or the Name conficts with other CRD already
// in the group
func (sc Config) createCRD(crd *apiext.CustomResourceDefinition) error {
	if _, err := sc.apiExtClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd); err != nil {
		if errors.IsAlreadyExists(err) {
			// CRD is already present, no need to do anything
			// In future can implement upgrade of CRDs here.
			// For upgrade, a patch can be created which can then be
			// used to upgrade the CRD
		} else {
			return err
		}
	}

	return wait.Poll(CRDRetryInterval, CRDTimeout, func() (done bool, err error) {
		c, err := sc.apiExtClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crd.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cond := range c.Status.Conditions {
			switch cond.Type {
			case apiext.Established:
				if cond.Status == apiext.ConditionTrue {
					return true, err
				}
			case apiext.NamesAccepted:
				if cond.Status == apiext.ConditionFalse {
					return false, fmt.Errorf("name conflict for %s : %v", crd.Name, cond.Reason)
				}
			}
		}

		return false, err
	})
}
