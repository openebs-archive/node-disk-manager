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
/*
	The contents of this package has its origins from the feature gate
	implementation in kubernetes.
	Refer :
		https://github.com/kubernetes/component-base/tree/master/featuregate
		https://github.com/kubernetes/kubernetes/tree/master/pkg/features

*/

package features

import (
	"fmt"
	"strings"

	"github.com/openebs/node-disk-manager/pkg/util"

	"k8s.io/klog"
)

// Feature is a typed string for a given feature
type Feature string

const (
	// GPTBasedUUID feature flag is used to enable the
	// blockdevice UUID algorithm mentioned in
	// https://github.com/openebs/openebs/pull/2666
	GPTBasedUUID Feature = "GPTBasedUUID"

	// APIService feature flag starts the GRPC server which provides functionality to manage block devices
	APIService Feature = "APIService"

	UseOSDisk Feature = "UseOSDisk"

	// ChangeDetection is used to enable detecting changes to
	// blockdevice size, filesystem, and mount-points.
	ChangeDetection Feature = "ChangeDetection"
)

// supportedFeatures is the list of supported features. This is used while parsing the
// feature flag given via command line
var supportedFeatures = []Feature{
	GPTBasedUUID,
	APIService,
	UseOSDisk,
	ChangeDetection,
}

// defaultFeatureGates is the default features that will be applied to the application
var defaultFeatureGates = map[Feature]bool{
	GPTBasedUUID:    true,
	APIService:      false,
	UseOSDisk:       false,
	ChangeDetection: false,
}

var featureDependencies = map[Feature][]Feature{}

// featureFlag is a map representing the flag and its state
type featureFlag map[Feature]bool

// FeatureGates is the global feature gate that can be used to check if a feature flag is enabled
// or disabled
var FeatureGates = NewFeatureGate()

// NewFeatureGate gets a new map with the default feature gates for the application
func NewFeatureGate() featureFlag {
	fg := make(featureFlag)

	// set the default feature gates
	for k, v := range defaultFeatureGates {
		fg[k] = v
	}

	return fg
}

// IsEnabled returns true if the feature is enabled
func (fg featureFlag) IsEnabled(f Feature) bool {
	return fg[f]
}

// SetFeatureFlag parses a slice of string and sets the feature flag.
func (fg featureFlag) SetFeatureFlag(features []string) error {
	if len(features) == 0 {
		klog.V(4).Info("No feature flags are set, default values will be used")
	}
	// iterate through each feature and set its state onto the featureFlag map
	for _, feature := range features {
		var f Feature
		// by default if a feature gate is provided, it is enabled
		isEnabled := true
		// if the feature is specified in the format
		// MyFeature=false, the string need to be parsed and
		// corresponding state to be set on the feature
		s := strings.Split(feature, "=")
		f = Feature(s[0])
		// only if length after splitting =2, we need to check whether the
		// feature is enabled or disabled
		if len(s) == 2 {
			isEnabled = util.CheckTruthy(s[1])
		} else if len(s) > 2 {
			// if length > 2 , there is some error in the format specified
			return fmt.Errorf("incorrect format. cannot parse feature %s", feature)
		}
		// check if the feature flag provided is available in the list of
		// supported features
		if !containsFeature(supportedFeatures, f) {
			return fmt.Errorf("unknown feature flag %s", f)
		}
		fg[f] = isEnabled
	}

	// We make sure features are only turned on if their dependencies are met
	// We memoize values to avoid computing the same dependency twice
	memoizedValues := make(featureFlag)
	for feature := range fg {
		_ = ValidateDependencies(feature, fg, memoizedValues)
	}

	for k, v := range fg {
		klog.Infof("Feature gate: %s, state: %s", k, util.StateStatus(v))
	}

	return nil
}

// Ensures features are disabled if their dependencies are unmet
// Returns true if a feature is enabled after validation
func ValidateDependencies(feature Feature, flags featureFlag, memoizedValues featureFlag) bool {
	if value, isMemoized := memoizedValues[feature]; isMemoized {
		return value
	}
	if disabled := !flags[feature]; disabled {
		return false
	}
	dependencies := featureDependencies[feature]
	for _, dependency := range dependencies {
		missingDependency := !ValidateDependencies(dependency, flags, memoizedValues)
		if missingDependency {
			flags[feature] = false
			klog.Infof("Feature %v was set to false due to missing dependency %v", feature, dependency)
		}
	}
	memoizedValues[feature] = flags[feature]
	return flags[feature]
}
