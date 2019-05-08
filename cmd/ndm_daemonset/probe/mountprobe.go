package probe

import (
	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/mount"
	"github.com/openebs/node-disk-manager/pkg/util"
)

// mountProbe contains required variables for populating diskInfo
type mountProbe struct {
	// Every new probe needs a controller object to register itself.
	// Here Controller consists of Clientset, kubeClientset, probes, etc which is used to
	// create, update, delete, deactivate the disk resources or list the probes already registered.
	Controller      *controller.Controller
	MountIdentifier *mount.Identifier
}

const (
	mountProbePriority = 5
	mountConfigKey     = "mount-probe"
)

var (
	mountProbeName  = "mount probe"
	mountProbeState = defaultEnabled
)

// init is used to get a controller object and then register itself
var mountProbeRegister = func() {
	// Get a controller object
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		glog.Error("unable to configure", mountProbeName)
		return
	}
	if ctrl.NDMConfig != nil {
		for _, probeConfig := range ctrl.NDMConfig.ProbeConfigs {
			if probeConfig.Key == mountConfigKey {
				mountProbeName = probeConfig.Name
				mountProbeState = util.CheckTruthy(probeConfig.State)
				break
			}
		}
	}
	newRegisterProbe := &registerProbe{
		priority:   mountProbePriority,
		name:       mountProbeName,
		state:      mountProbeState,
		pi:         &mountProbe{Controller: ctrl},
		controller: ctrl,
	}
	// Here we register the probe (smart probe in this case)
	newRegisterProbe.register()
}

// newMountProbe returns mountProbe struct which helps populate diskInfo struct
// with the mount related details like mountpoint
func newMountProbe(devPath string) *mountProbe {
	mountIdentifier := &mount.Identifier{
		DevPath: devPath,
	}
	mountProbe := &mountProbe{
		MountIdentifier: mountIdentifier,
	}
	return mountProbe
}

// Start is mainly used for one time activities such as monitoring.
// It is a part of probe interface but here we does not require to perform
// such activities, hence empty implementation
func (mp *mountProbe) Start() {}

// FillDiskDetails fills details in diskInfo struct using information it gets from probe
func (mp *mountProbe) FillDiskDetails(d *controller.DiskInfo) {
	if d.ProbeIdentifiers.MountIdentifier == "" {
		glog.Error("mountIdentifier is found empty, mount probe will not fetch mount information.")
		return
	}
	mountProbe := newMountProbe(d.ProbeIdentifiers.MountIdentifier)
	basicMountInfo, err := mountProbe.MountIdentifier.DeviceBasicMountInfo()
	if err != nil {
		glog.Error(err)
		return
	}
	d.FileSystemInformation.MountPoint = basicMountInfo.MountPoint
	if d.FileSystemInformation.FileSystem == "" {
		d.FileSystemInformation.FileSystem = basicMountInfo.FileSystem
	}
}
