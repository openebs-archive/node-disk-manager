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
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

const (
	// JobContainerName is the name of the cleanup job container
	JobContainerName = "cleaner"
	// JobNamePrefix is the prefix for the cleanup job name
	JobNamePrefix = "cleanup-"
	// JobImageName is the image to be used for the cleanup job container
	JobImageName = "busybox"
	// BDLabel is the label set on the job for identification of the BD
	BDLabel = "blockdevice"
	// BlockCleanerCommand is the command used to clean raw block
	BlockCleanerCommand = "dd"
	// FSCleanerCommand is the command to clear data on the filesystem
	FSCleanerCommand = "rm"
)

// JobController defines the interface for the job controller.
type JobController interface {
	IsCleaningJobRunning(bdName string) bool
	RemoveJob(bdName string) (CleanupState, error)
}

var _ JobController = &jobController{}

type jobController struct {
	client    client.Client
	namespace string
}

// NewCleanupJob creates a new cleanup job in the  namespace. It returns a Job object which can be used to
// start the job
func NewCleanupJob(bd *v1alpha1.BlockDevice, volMode VolumeMode, namespace string) (*batchv1.Job, error) {
	nodeName := bd.Labels[controller.NDMHostKey]

	priv := true
	jobContainer := v1.Container{
		Name:  JobContainerName,
		Image: JobImageName,
		SecurityContext: &v1.SecurityContext{
			Privileged: &priv,
		},
	}

	podSpec := v1.PodSpec{}

	if volMode == VolumeModeBlock {
		input := "if=/dev/zero"
		output := "of=" + bd.Spec.Path
		blockSize := "bs=1M"
		// get no of blocks required for a block size of 1M
		blockCount := bd.Spec.Capacity.Storage / 1024 / 1024
		count := "count=" + strconv.FormatUint(blockCount, 10)
		jobContainer.Command = getCommand(BlockCleanerCommand, input, output, blockSize, count)
	} else if volMode == VolumeModeFileSystem {
		deleteOptions := "-rf"
		directory := "/tmp/*"
		jobContainer.Command = getCommand(FSCleanerCommand, deleteOptions, directory)
		mountName := "vol-mount"
		volumes := []v1.Volume{
			{
				Name: mountName,
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: bd.Spec.FileSystem.Mountpoint,
					},
				},
			},
		}
		jobContainer.VolumeMounts = []v1.VolumeMount{{
			Name:      mountName,
			MountPath: "/tmp",
		}}
		podSpec.Volumes = volumes
	}

	podSpec.Containers = []v1.Container{jobContainer}
	podSpec.NodeSelector = map[string]string{controller.NDMHostKey: nodeName}

	podTemplate := v1.Pod{}
	podTemplate.Spec = podSpec

	labels := map[string]string{
		controller.NDMHostKey: nodeName,
		BDLabel:               bd.Name,
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

	err = c.client.Delete(context.TODO(), job, client.PropagationPolicy(metav1.DeletePropagationForeground))
	if err != nil {
		return CleanupStateUnknown, err
	}

	return CleanupStateSucceeded, nil
}

func generateCleaningJobName(bdName string) string {
	return JobNamePrefix + bdName
}

func getCommand(cmd string, args ...string) []string {
	var command []string
	command = append(command, cmd)
	for _, arg := range args {
		command = append(command, arg)
	}
	return command
}
