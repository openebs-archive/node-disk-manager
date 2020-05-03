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

package features

import (
	"reflect"
	"testing"
)

func TestFeatureGateIsEnabled(t *testing.T) {
	testFG := make(FeatureGate)
	testFG["feature1"] = false
	testFG["feature2"] = true
	tests := map[string]struct {
		fg      FeatureGate
		feature Feature
		want    bool
	}{
		"when feature gate is empty": {
			fg:      nil,
			feature: "test",
			want:    false,
		},
		"when feature gate does not have the feature": {
			fg:      testFG,
			feature: "feature3",
			want:    false,
		},
		"when feature gate has the feature and feature is disabled": {
			fg:      testFG,
			feature: "feature1",
			want:    false,
		},
		"when feature gate has the feature and feature is enabled": {
			fg:      testFG,
			feature: "feature2",
			want:    true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.fg.IsEnabled(tt.feature); got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFeatureGate(t *testing.T) {
	F1 := Feature("FeatureGate1")
	F2 := Feature("FeatureGate2")
	F3 := Feature("FeatureGate3")
	type args struct {
		features   []string
		defaultFGs []Feature
	}
	tests := map[string]struct {
		args    args
		want    FeatureGate
		wantErr bool
	}{
		"empty feature gate slice": {
			args: args{
				features:   nil,
				defaultFGs: DefaultFeatureGates,
			},
			want:    make(map[Feature]bool),
			wantErr: false,
		},
		"a single feature is added": {
			args: args{
				features:   []string{"GPTBasedUUID"},
				defaultFGs: DefaultFeatureGates,
			},
			want: map[Feature]bool{
				GPTBasedUUID: true,
			},
		},
		"a single feature is set in disabled mode": {
			args: args{
				features:   []string{"GPTBasedUUID=false"},
				defaultFGs: DefaultFeatureGates,
			},
			want: map[Feature]bool{
				GPTBasedUUID: false,
			},
			wantErr: false,
		},
		"feature that is not present in the default feature": {
			args: args{
				features:   []string{"WrongFeatureGate"},
				defaultFGs: DefaultFeatureGates,
			},
			want:    make(map[Feature]bool),
			wantErr: true,
		},
		"multiple features in enabled and disabled state": {
			args: args{
				features:   []string{"FeatureGate1", "FeatureGate2=false", "FeatureGate3=true"},
				defaultFGs: []Feature{F1, F2, F3},
			},
			want: FeatureGate{
				F1: true,
				F2: false,
				F3: true,
			},
			wantErr: false,
		},
		"wrong format in one feature gate": {
			args: args{
				features:   []string{"FeatureGate1", "FeatureGate2=true=true"},
				defaultFGs: []Feature{F1, F2, F3},
			},
			want: FeatureGate{
				F1: true,
			},
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseFeatureGate(tt.args.features, tt.args.defaultFGs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFeatureGate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseFeatureGate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
