package features

import (
	"fmt"
	"strings"

	"github.com/openebs/node-disk-manager/pkg/util"

	"k8s.io/klog"
)

// FeatureGate type represents the map of features and the state
type FeatureGate map[Feature]bool

// Feature is a typed string for a given feature
type Feature string

const (
	// GPTBasedUUID feature flag is used to enable the
	// blockdevice UUID algorithm mentioned in
	// https://github.com/openebs/openebs/pull/2666
	GPTBasedUUID Feature = "GPTBasedUUID"
)

// DefaultFeatureGates is the list of feature gates that are added by default
var DefaultFeatureGates = []Feature{
	GPTBasedUUID,
}

// IsEnabled returns true if the feature is enabled
func (fg FeatureGate) IsEnabled(f Feature) bool {
	return fg[f]
}

// ParseFeatureGate parses a slice of string and create the feature-gate map
func ParseFeatureGate(features []string, defaultFGs []Feature) (FeatureGate, error) {
	fg := make(FeatureGate)
	if len(features) == 0 {
		klog.V(4).Info("No feature gates are set")
		return fg, nil
	}
	// iterate through each feature and set its state onto the FeatureGate map
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
			return fg, fmt.Errorf("incorrect format. cannot parse feature %s", feature)
		}
		// check if the feature flag provided was available in the list of
		// supported features
		if !containsFeature(defaultFGs, f) {
			return fg, fmt.Errorf("unknown feature flag %s", f)
		}
		fg[f] = isEnabled
		klog.Infof("Feature gate: %s, state: %s", f, util.StateStatus(isEnabled))
	}
	return fg, nil
}
