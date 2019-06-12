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

// runJob creates a new cleanup job in the namespace
func (c *Cleaner) runJob(bd *v1alpha1.BlockDevice, volumeMode VolumeMode) error {
	job, err := NewCleanupJob(bd, volumeMode, c.Namespace)
	if err != nil {
		return err
	}
	return c.Client.Create(context.TODO(), job)
}

// getVolumeMode returns the volume mode of the BD that is being released
func getVolumeMode(spec v1alpha1.DeviceSpec) VolumeMode {
	if spec.FileSystem.Mountpoint != "" {
		return VolumeModeFileSystem
	} else {
		return VolumeModeBlock
	}
}
