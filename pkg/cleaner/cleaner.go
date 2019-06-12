package cleaner

import (
	"context"
	"github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CleanupState int

const (
	CleanupStateUnknown CleanupState = iota + 1
	CleanupStateRunning
	CleanupStateSucceeded
)

type VolumeMode string

const (
	VolumeModeBlock      = "BlockVolumeMode"
	VolumeModeFileSystem = "FileSystemVolumeMode"
)

// Cleaner handles BD cleanup
// For filesystem/mount based block devices, it deletes the contents of the directory
// For raw block devices, a `dd` command will be issued.
type Cleaner struct {
	Client        client.Client
	Namespace     string
	CleanupStatus *CleanupStatusTracker
}

type CleanupStatusTracker struct {
	JobController JobController
}

func NewCleaner(client client.Client, namespace string, cleanupTracker *CleanupStatusTracker) *Cleaner {
	return &Cleaner{
		Client:        client,
		Namespace:     namespace,
		CleanupStatus: cleanupTracker,
	}
}

func (c *Cleaner) Clean(blockDevice *v1alpha1.BlockDevice) (bool, error) {
	bdName := blockDevice.Name
	if c.CleanupStatus.InProgress(bdName) {
		return false, nil
	}
	// Check if cleaning was just completed.
	state, err := c.CleanupStatus.RemoveStatus(bdName)
	if err != nil {
		return false, nil
	}

	switch state {
	case CleanupStateSucceeded:
		return true, nil
	}

	volMode := getVolumeMode(blockDevice.Spec)

	err = c.runJob(blockDevice, volMode)

	return false, err
}

func (c *CleanupStatusTracker) InProgress(bdName string) bool {
	return c.JobController.IsCleaningJobRunning(bdName)
}

func (c *CleanupStatusTracker) RemoveStatus(bdName string) (CleanupState, error) {
	return c.JobController.RemoveJob(bdName)
}

func (c *Cleaner) runJob(bd *v1alpha1.BlockDevice, volumeMode VolumeMode) error {
	job, err := NewCleanupJob(bd, volumeMode, c.Namespace)
	if err != nil {
		return err
	}
	return c.Client.Create(context.TODO(), job)
}

func getVolumeMode(spec v1alpha1.DeviceSpec) VolumeMode {
	if spec.FileSystem.Mountpoint != "" {
		return VolumeModeFileSystem
	} else {
		return VolumeModeBlock
	}
}
