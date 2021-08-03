/*
Copyright 2019 OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

This pkg is inspired from the deleter pkg in local-static-provisioner
in kubernetes-sigs
	https://github.com/kubernetes-sigs/sig-storage-local-static-provisioner/tree/master/pkg/deleter
*/

package cleaner

import (
	"context"
	"fmt"
	"github.com/openebs/node-disk-manager/api/v1alpha1"
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/env"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// JobContainerName is the name of the cleanup job container
	JobContainerName = "cleaner"
	// JobNamePrefix is the prefix for the cleanup job name
	JobNamePrefix = "cleanup-"
	// BDLabel is the label set on the job for identification of the BD
	BDLabel = "blockdevice"
)

// JobController defines the interface for the job controller.
type JobController interface {
	IsCleaningJobRunning(bdName string) bool
	CancelJob(bdName string) error
	RemoveJob(bdName string) (CleanupState, error)
}

var _ JobController = &jobController{}

type jobController struct {
	client    client.Client
	namespace string
}

// NewCleanupJob creates a new cleanup job in the  namespace. It returns a Job object which can be used to
// start the job
func NewCleanupJob(bd *v1alpha1.BlockDevice, volMode VolumeMode, tolerations []v1.Toleration, namespace string) (*batchv1.Job, error) {
	nodeName := bd.Labels[controller.KubernetesHostNameLabel]

	priv := true
	jobContainer := v1.Container{
		Name:  JobContainerName,
		Image: getCleanUpImage(),
		SecurityContext: &v1.SecurityContext{
			Privileged: &priv,
		},
	}

	podSpec := v1.PodSpec{}
	mountName := "vol-mount"

	if volMode == VolumeModeBlock {
		jobContainer.Command = []string{"/bin/sh", "-c"}

		// fdisk is used to get all the partitions of the device.
		// Example
		// $ fdisk -o Device -l /dev/sda
		// 	Disk /dev/sda: 465.8 GiB, 500107862016 bytes, 976773168 sectors
		// 	Units: sectors of 1 * 512 = 512 bytes
		// 	Sector size (logical/physical): 512 bytes / 4096 bytes
		// 	I/O size (minimum/optimal): 4096 bytes / 4096 bytes
		// 	Disklabel type: dos
		// 	Disk identifier: 0x065e2357
		//
		// 	Device
		// 	/dev/sda1
		// 	/dev/sda2
		// 	/dev/sda5
		// 	/dev/sda6
		// 	/dev/sda7
		//
		// From the above output the partitions are filtered using grep,
		//
		// first all the partitions are cleared off any filesystem signatures, then the actual partition table
		// header is removed. partprobe is called so as to re-read partition table, and update system with
		// the changes.  Partprobe will be called only if the device is a block file; else if its sparse file, wipefs
		// will be done.
		// wipefs erases the filesystem signature from the block
		// -a    wipe all magic strings
		// -f    force erasure
		args := fmt.Sprintf("(fdisk -o Device -l %[1]s "+
			"| grep \"^%[1]s\" "+
			"| xargs -I '{}' wipefs -fa '{}') "+
			"&& wipefs -fa %[1]s ",
			bd.Spec.Path)

		// partprobe need to be executed only if the device is of type disk.
		if bd.Spec.Details.DeviceType == blockdevice.BlockDeviceTypeDisk {
			args += fmt.Sprintf("&& partprobe %s ", bd.Spec.Path)
		}

		jobContainer.Args = []string{args}

		// in case of sparse disk, need to mount the sparse file directory
		// and clear the sparse file
		if bd.Spec.Details.DeviceType == blockdevice.SparseBlockDeviceType {
			volume, volumeMount := getVolumeMounts(bd.Spec.Path, bd.Spec.Path, mountName)
			jobContainer.VolumeMounts = []v1.VolumeMount{volumeMount}
			podSpec.Volumes = []v1.Volume{volume}
		}

	} else if volMode == VolumeModeFileSystem {
		jobContainer.Command = []string{"/bin/sh", "-c"}
		jobContainer.Args = []string{"find /tmp -mindepth 1 -maxdepth 1 -print0 | xargs -0 rm -rf"}
		volume, volumeMount := getVolumeMounts(bd.Spec.FileSystem.Mountpoint, "/tmp", mountName)

		jobContainer.VolumeMounts = []v1.VolumeMount{volumeMount}
		podSpec.Volumes = []v1.Volume{volume}
	}

	podSpec.Tolerations = tolerations
	podSpec.ServiceAccountName = getServiceAccount()
	podSpec.Containers = []v1.Container{jobContainer}
	podSpec.NodeSelector = map[string]string{controller.KubernetesHostNameLabel: nodeName}
	podSpec.ImagePullSecrets = env.GetOpenEBSImagePullSecrets()
	podTemplate := v1.Pod{}
	podTemplate.Spec = podSpec

	labels := map[string]string{
		controller.KubernetesHostNameLabel: nodeName,
		BDLabel:                            bd.Name,
	}

	podTemplate.ObjectMeta = metav1.ObjectMeta{
		Name:      generateCleaningJobName(bd.Name),
		Namespace: namespace,
		Labels:    labels,
	}

	job := &batchv1.Job{}
	job.ObjectMeta = podTemplate.ObjectMeta
	job.Spec.Template.Spec = podTemplate.Spec
	job.Spec.Template.Spec.RestartPolicy = v1.RestartPolicyOnFailure

	return job, nil
}

// NewJobController returns a job controller struct which can be used to get the status
// of the running job
func NewJobController(client client.Client, namespace string) *jobController {
	return &jobController{
		client:    client,
		namespace: namespace,
	}
}

func (c *jobController) IsCleaningJobRunning(bdName string) bool {
	jobName := generateCleaningJobName(bdName)
	objKey := client.ObjectKey{
		Namespace: c.namespace,
		Name:      jobName,
	}
	job := &batchv1.Job{}

	err := c.client.Get(context.TODO(), objKey, job)
	if errors.IsNotFound(err) {
		return false
	}

	if err != nil {
		// failed to check whether it is running, assuming job is still running
		return true
	}

	return job.Status.Succeeded <= 0
}

func (c *jobController) RemoveJob(bdName string) (CleanupState, error) {
	jobName := generateCleaningJobName(bdName)
	objKey := client.ObjectKey{
		Namespace: c.namespace,
		Name:      jobName,
	}
	job := &batchv1.Job{}

	err := c.client.Get(context.TODO(), objKey, job)
	if err != nil {
		if errors.IsNotFound(err) {
			return CleanupStateNotFound, nil
		}
		return CleanupStateUnknown, err
	}
	if job.Status.Succeeded == 0 {
		return CleanupStateRunning, nil
	}

	// cancel the job
	err = c.CancelJob(bdName)
	if err != nil {
		return CleanupStateUnknown, err
	}

	return CleanupStateSucceeded, nil
}

// CancelJob deletes a job, if it is present. if the job is not present, it will return an error.
func (c *jobController) CancelJob(bdName string) error {
	jobName := generateCleaningJobName(bdName)
	objKey := client.ObjectKey{
		Namespace: c.namespace,
		Name:      jobName,
	}
	job := &batchv1.Job{}
	err := c.client.Get(context.TODO(), objKey, job)

	err = c.client.Delete(context.TODO(), job, client.PropagationPolicy(metav1.DeletePropagationForeground))
	return err
}

func generateCleaningJobName(bdName string) string {
	return JobNamePrefix + bdName
}

// GetNodeName gets the Node name from BlockDevice
func GetNodeName(bd *v1alpha1.BlockDevice) string {
	return bd.Spec.NodeAttributes.NodeName
}

// getVolumeMounts returns the volume and volume mount for the given hostpath and
// mountpath
func getVolumeMounts(hostPath, mountPath, mountName string) (v1.Volume, v1.VolumeMount) {
	volumes := v1.Volume{
		Name: mountName,
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: hostPath,
			},
		},
	}

	volumeMount := v1.VolumeMount{
		Name:      mountName,
		MountPath: mountPath,
	}

	return volumes, volumeMount
}
