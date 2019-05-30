package verify

import (
	"fmt"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

// GetRequestedCapacity gets the requested capacity from the BlockDeviceClaim
// It returns an error if the Quantity cannot be parsed
func GetRequestedCapacity(list v1.ResourceList) (int64, error) {

	resourceCapacity := list[apis.ResourceCapacity]
	// Check if deviceClaim has valid capacity request
	capacity, err := (&resourceCapacity).AsInt64()
	if !err || capacity <= 0 {
		return 0, fmt.Errorf("invalid capacity requested, %v", err)
	}
	return capacity, nil
}
