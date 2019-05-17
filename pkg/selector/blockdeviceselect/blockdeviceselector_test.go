package blockdeviceselect

/*func TestSelectBlockDevice(t *testing.T) {
	bdList := openebsv1alpha1.BlockDeviceList{}
	bd1 := GetFakeDeviceObject("blockdevice-example1", 102400)
	bd2 := GetFakeDeviceObject("blockdevice-example1", 1024000)
	bdList.Items = append(bdList.Items, *bd1, *bd2)

	resourceList1 := v1.ResourceList{openebsv1alpha1.ResourceCapacity: resource.MustParse("102400")}
	resourceList2 := v1.ResourceList{openebsv1alpha1.ResourceCapacity: resource.MustParse("2048000")}

	tests := map[string]struct {
		deviceList     openebsv1alpha1.BlockDeviceList
		rList          v1.ResourceList
		expectedDevice openebsv1alpha1.BlockDevice
		expectedOk     bool
	}{
		"can find a block device with matching requirements":    {bdList, resourceList1, *bd1, true},
		"cannot find a block device with matching requirements": {bdList, resourceList2, openebsv1alpha1.BlockDevice{}, false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bd, ok := selectBlockDevice(test.deviceList, test.rList)
			assert.Equal(t, test.expectedDevice, bd)
			assert.Equal(t, test.expectedOk, ok)
		})
	}
}*/

/*func TestMatchResourceRequirements(t *testing.T) {
	blockDevice := GetFakeDeviceObject(deviceName, capacity)
	tests := map[string]struct {
		blockDevice *openebsv1alpha1.BlockDevice
		rList       v1.ResourceList
		expected    bool
	}{
		"block device capacity greater than requested capacity": {blockDevice,
			v1.ResourceList{openebsv1alpha1.ResourceCapacity: resource.MustParse("1024000")},
			true},
		"block device capacity is less than requested capacity": {blockDevice,
			v1.ResourceList{openebsv1alpha1.ResourceCapacity: resource.MustParse("404800000")},
			false},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, matchResourceRequirements(*test.blockDevice, test.rList))
		})
	}
}*/
