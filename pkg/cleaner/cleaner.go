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
	"github.com/openebs/node-disk-manager/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
)

// CleanupState represents the current state of the cleanup job
type CleanupState int

const (
	// CleanupStateUnknown represents an unknown state of the cleanup job
	CleanupStateUnknown CleanupState = iota + 1
	// CleanupStateNotFound defines the state when a job does not exist
	CleanupStateNotFound
	// CleanupStateRunning represents a running cleanup job
	CleanupStateRunning
	// CleanupStateSucceeded represents that the cleanup job has been completed successfully
	CleanupStateSucceeded
)

// VolumeMode defines the volume mode of the BlockDevice. It can be either block mode or
// filesystem mode
type VolumeMode string

const (
	// VolumeModeBlock defines a raw block volume mode which means the block device should
	// be treated as raw block device
	VolumeModeBlock = "BlockVolumeMode"
	// VolumeModeFileSystem defines that the blockdevice should be treated as a block
	// formatted with filesystem and is mounted
	VolumeModeFileSystem = "FileSystemVolumeMode"
)

// Cleaner handles BD cleanup
// For filesystem/mount based block devices, it deletes the contents of the directory
// For raw block devices, a `wipefs` command will be issued.
type Cleaner struct {
	Client        client.Client
	Namespace     string
	CleanupStatus *CleanupStatusTracker
}

// CleanupStatusTracker is used to track the cleanup state using
// info provided by JobController
type CleanupStatusTracker struct {
	JobController JobController
}

// NewCleaner creates a new cleaner which can be used for cleaning BD, and checking status of the job
func NewCleaner(client client.Client, namespace string, cleanupTracker *CleanupStatusTracker) *Cleaner {
	return &Cleaner{
		Client:        client,
		Namespace:     namespace,
		CleanupStatus: cleanupTracker,
	}
}

// Clean will launch a job to delete data on the BD depending on the
// volume mode. Job will be launched only if another job is not running or a
// job is in unknown state
func (c *Cleaner) Clean(blockDevice *v1alpha1.BlockDevice) (bool, error) {
	bdName := blockDevice.Name
	// check if a cleanup job for the bd already exists and return
	if c.CleanupStatus.InProgress(bdName) {
		// check if the BD for which the cleanup is performed is still active,
		// if not, cancel the cleanup job
		if blockDevice.Status.State != v1alpha1.BlockDeviceActive {
			// cancel the job
			if err := c.CleanupStatus.CancelJob(bdName); err != nil {
				return false, err
			}
		}
		return false, nil
	}
	// Check if cleaning was just completed. if job was completed, it will be removed,
	// The state of the job will be returned.
	state, err := c.CleanupStatus.RemoveStatus(bdName)
	if err != nil {
		return false, nil
	}

	switch state {
	case CleanupStateSucceeded:
		return true, nil
	case CleanupStateNotFound:
		// if the BD is not active, do not start the job
		if blockDevice.Status.State != v1alpha1.BlockDeviceActive {
			return false, nil
		}
		// we are starting the job, since no job for the BD was found and
		// BD is in active state
	}

	volMode := getVolumeMode(blockDevice.Spec)

	// create a new job for the blockdevice
	err = c.runJob(blockDevice, volMode)

	return false, err
}

// InProgress returns whether a cleanup job is currently being done
// for the given BD
func (c *CleanupStatusTracker) InProgress(bdName string) bool {
	return c.JobController.IsCleaningJobRunning(bdName)
}

// RemoveStatus returns the Cleanupstate of a job. If job is succeeded, it will
// be deleted
func (c *CleanupStatusTracker) RemoveStatus(bdName string) (CleanupState, error) {
	return c.JobController.RemoveJob(bdName)
}

// CancelJob cancels a job without checking the status of the job.
func (c *CleanupStatusTracker) CancelJob(bdName string) error {
	return c.JobController.CancelJob(bdName)
}

// runJob creates a new cleanup job in the namespace
func (c *Cleaner) runJob(bd *v1alpha1.BlockDevice, volumeMode VolumeMode) error {

	// retrieve node Object to pass tolerations to the Job
	nodeName := GetNodeName(bd)
	selectedNode, err := c.getNodeObjectByNodeName(nodeName)
	if err != nil {
		return err
	}
	tolerations := getTolerationsForTaints(selectedNode.Spec.Taints...)

	job, err := NewCleanupJob(bd, volumeMode, tolerations, c.Namespace)
	if err != nil {
		return err
	}
	return c.Client.Create(context.TODO(), job)
}

// getNodeObjectByNodeName returns Node Object, using NodeName
func (c *Cleaner) getNodeObjectByNodeName(nodeName string) (*v1.Node, error) {
	node := &v1.Node{}
	err := c.Client.Get(context.TODO(), client.ObjectKey{Namespace: "", Name: nodeName}, node)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// getTolerationsForTaints returns tolerations, taking input as taints
func getTolerationsForTaints(taints ...v1.Taint) []v1.Toleration {
	tolerations := []v1.Toleration{}
	for i := range taints {
		var toleration v1.Toleration
		toleration.Key = taints[i].Key
		toleration.Effect = taints[i].Effect
		if len(taints[i].Value) == 0 {
			toleration.Operator = v1.TolerationOpExists
		} else {
			toleration.Value = taints[i].Value
			toleration.Operator = v1.TolerationOpEqual
		}
		tolerations = append(tolerations, toleration)
	}
	return tolerations
}

// getVolumeMode returns the volume mode of the BD that is being released
func getVolumeMode(spec v1alpha1.DeviceSpec) VolumeMode {
	if spec.FileSystem.Mountpoint != "" {
		return VolumeModeFileSystem
	}
	return VolumeModeBlock
}
