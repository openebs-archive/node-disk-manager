package blockdeviceselect

import (
	"fmt"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/selector/verify"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

	for _, bd := range bdList.Items {
		// if manual selection is enabled, we select the device and breaks the iteration
		if c.ManualSelection && bd.Name == c.ClaimSpec.BlockDeviceName {
			candidateBD.Items = append(candidateBD.Items, bd)
			break
		}
		// check for active and unclaimed devices
		if bd.Status.State != controller.NDMActive ||
			bd.Status.ClaimState != apis.BlockDeviceUnclaimed {
			continue
		}
		if verifyDeviceType && bd.Spec.Details.DeviceType != c.ClaimSpec.DeviceType {
			continue
		}
		candidateBD.Items = append(candidateBD.Items, bd)
	}

	if len(candidateBD.Items) == 0 {
		return nil, fmt.Errorf("no devices found matching the criteria")
	}

	return candidateBD, nil
}

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
