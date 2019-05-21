package blockdeviceselect

import (
	"fmt"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/resourceselector/verify"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BlockDeviceSelector selects a single block device from a list of block devices
func (c *Config) BlockDeviceSelector(bdList *apis.BlockDeviceList) (*apis.BlockDevice, error) {
	candidateDevices, err := c.getCandidateDevices(bdList)
	if err != nil {
		return nil, err
	}
	selectedDevice, err := c.getSelectedDevice(candidateDevices)
	if err != nil {
		return nil, err
	}
	return selectedDevice, nil
}

// getCandidateDevices selects a list of blockdevices from a given block device
// list based on criteria specified in the claim spec
func (c *Config) getCandidateDevices(bdList *apis.BlockDeviceList) (*apis.BlockDeviceList, error) {
	verifyDeviceType := false
	if c.ClaimSpec.DeviceType != "" {
		verifyDeviceType = true
	}
	candidateBD := &apis.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}
	// iterate through each BlockDevice and gets all the eligible devices.
	// in case of manual selection, all other checks are skipped, only the name is
	// matched
	for _, bd := range bdList.Items {
		if c.ManualSelection {
			if bd.Name == c.ClaimSpec.BlockDeviceName {
				candidateBD.Items = append(candidateBD.Items, bd)
				break
			}
		} else {
			if bd.Status.State != controller.NDMActive ||
				bd.Status.ClaimState != apis.BlockDeviceUnclaimed {
				continue
			}
			// sparse disk can be selected only by manual selection
			if bd.Spec.Details.DeviceType == controller.SparseBlockDeviceType {
				continue
			}
			if verifyDeviceType && bd.Spec.Details.DeviceType != c.ClaimSpec.DeviceType {
				continue
			}
			candidateBD.Items = append(candidateBD.Items, bd)
		}
	}

	if len(candidateBD.Items) == 0 {
		return nil, fmt.Errorf("no devices found matching the criteria")
	}

	return candidateBD, nil
}

// getSelectedDevice selects a single a block device based on the resource requirements
// requested by the claim
func (c *Config) getSelectedDevice(bdList *apis.BlockDeviceList) (*apis.BlockDevice, error) {
	if c.ManualSelection {
		return &bdList.Items[0], nil
	}
	for _, bd := range bdList.Items {
		if matchResourceRequirements(bd, c.ClaimSpec.Requirements.Requests) {
			return &bd, nil
		}
	}
	return nil, fmt.Errorf("could not find a device with matching requirements")
}

// matchResourceRequirements matches a block device with a ResourceList
func matchResourceRequirements(bd apis.BlockDevice, list v1.ResourceList) bool {
	capacity, _ := verify.GetRequestedCapacity(list)
	return bd.Spec.Capacity.Storage >= uint64(capacity)
}
