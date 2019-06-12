package cleaner

import (
	"context"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	JobContainerName    = "cleaner"
	JobNamePrefix       = "cleanup-"
	JobImageName        = "busybox"
	BDLabel             = "blockdevice"
	BlockCleanerCommand = "dd if=/dev/zero of="
	FSCleanerCommand    = "rm -rf /tmp"
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
		jobContainer.Command = getCommand(BlockCleanerCommand + bd.Spec.Path)
	} else if volMode == VolumeModeFileSystem {
		jobContainer.Command = getCommand(FSCleanerCommand)
		mountName := "fsMount"
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

	podTemplate.ObjectMeta = v12.ObjectMeta{
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
		return CleanupStateUnknown, err
	}
	if job.Status.Succeeded == 0 {
		return CleanupStateRunning, nil
	}

	if err = c.client.Delete(context.TODO(), job); err != nil {
		return CleanupStateUnknown, err
	}

	return CleanupStateSucceeded, nil
}

func generateCleaningJobName(bdName string) string {
	return JobNamePrefix + bdName
}

func getCommand(cmd string) []string {
	return strings.Split(cmd, " ")
}
