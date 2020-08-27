/*
Copyright 2020 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package probe

import (
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/db/kubernetes"
	"github.com/openebs/node-disk-manager/pkg/util"

	"k8s.io/klog"
)

const (
	customTagProbePriority = 7

	tagTypePath = "path"
)

var (
	customTagProbeState = defaultEnabled

	supportedTagTypes = []string{tagTypePath}
)

type customTagProbe struct {
	tags []tag
}

type tag struct {
	tagType string
	regex   string
	label   string
}

// The label validation regex should be the same as used in
// https://github.com/kubernetes/apimachinery/blob/master/pkg/util/validation/validation.go
const labelValidatorRegex = "(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?"

var customTagProbeRegister = func() {
	// Get a controller object
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		klog.Error("unable to configure custom tag probe")
		return
	}
	tagProbe := &customTagProbe{}

	if ctrl.NDMConfig != nil {
		for _, tagConfig := range ctrl.NDMConfig.TagConfigs {
			if !util.Contains(supportedTagTypes, tagConfig.Type) {
				klog.Errorf("unsupported tag type: %s", tagConfig.Type)
			}

			if !util.IsMatchRegex(labelValidatorRegex, tagConfig.TagName) {
				klog.Errorf("not a valid label \"%s\"", tagConfig.TagName)
			}

			tagProbe.tags = append(tagProbe.tags, tag{
				tagType: tagConfig.Type,
				regex:   tagConfig.Pattern,
				label:   tagConfig.TagName,
			})
		}
	}
	newRegisterProbe := &registerProbe{
		priority:   customTagProbePriority,
		name:       "Custom Tag Probe",
		state:      customTagProbeState,
		pi:         tagProbe,
		controller: ctrl,
	}
	newRegisterProbe.register()
}

func (ctp *customTagProbe) Start() {}

func (ctp *customTagProbe) FillBlockDeviceDetails(bd *blockdevice.BlockDevice) {
	for _, tag := range ctp.tags {
		var fieldToMatch string
		switch tag.tagType {
		case tagTypePath:
			fieldToMatch = bd.DevPath
		}
		if bd.Labels == nil {
			bd.Labels = make(map[string]string)
		}
		if util.IsMatchRegex(tag.regex, fieldToMatch) {
			bd.Labels[kubernetes.BlockDeviceTagLabel] = tag.label
			klog.Infof("Device: %s Label %s:%s added by custom tag probe", bd.DevPath, kubernetes.BlockDeviceTagLabel, tag.label)
		}
	}
}
