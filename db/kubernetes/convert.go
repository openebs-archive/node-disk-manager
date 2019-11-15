package kubernetes

import (
	"github.com/openebs/node-disk-manager/blockdevice"
	api "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
)

func convert_BlockDeviceAPIList_To_BlockDeviceList(in *api.BlockDeviceList, out *[]blockdevice.BlockDevice) error {
	var err error
	var bd blockdevice.BlockDevice

	for _, bdAPI := range in.Items {
		err = convert_BlockDeviceAPI_To_BlockDevice(&bdAPI, &bd)
		if err != nil {
			return err
		}
		*out = append(*out, bd)
	}
	return nil
}

func convert_BlockDeviceAPI_To_BlockDevice(in *api.BlockDevice, out *blockdevice.BlockDevice) error {
	out.UUID = in.Name

	//labels
	out.NodeAttributes = make(blockdevice.NodeAttribute)
	out.NodeAttributes[blockdevice.HostName] = in.Labels[KubernetesHostNameLabel]

	//spec
	out.Path = in.Spec.Path
	out.FSInfo.FileSystem = in.Spec.FileSystem.Type

	// currently only the first mount point is filled in. When API is changed, multiple mount points
	// will be added.
	out.FSInfo.MountPoint = make([]string, 0)
	out.FSInfo.MountPoint[0] = in.Spec.FileSystem.Mountpoint

	//status
	out.Status.State = string(in.Status.State)
	out.Status.ClaimPhase = string(in.Status.ClaimState)

	return nil
}
