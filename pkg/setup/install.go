package setup

import "fmt"

// Install installs the components based on configuration provided
func (sc Config) Install() error {

	var err error
	// create CRDs
	if err = sc.createDiskCRD(); err != nil {
		return fmt.Errorf("disk CRD creation failed : %v", err)
	}
	if err = sc.createBlockDeviceCRD(); err != nil {
		return fmt.Errorf("block device CRD creation failed : %v", err)
	}
	if err = sc.createBlockDeviceClaimCRD(); err != nil {
		return fmt.Errorf("block device claim CRD creation failed : %v", err)
	}

	return nil
}
