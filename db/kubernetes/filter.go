/*
Copyright 2019 The OpenEBS Authors

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

package kubernetes

import (
	"errors"
	"github.com/openebs/node-disk-manager/blockdevice"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GenerateLabelFilter is used to generate a label filter that can be used
// while listing resources
func GenerateLabelFilter(key, value string) (interface{}, error) {

	// if key or value is empty, filter will be empty string
	if len(key) == 0 ||
		len(value) == 0 {
		return nil, errors.New("key/value is empty for label filter")
	}

	filterKey := getFilterKey(key)

	label := client.MatchingLabels{filterKey: value}

	return label, nil
}

// GenerateLabelFilterWithOp creates a label filter with the given operator
// between key and value
func GenerateLabelFilterWithOp(key, op, value string) (interface{}, error) {
	if len(key) == 0 ||
		len(value) == 0 ||
		len(op) == 0 {
		return nil, errors.New("key/operator/value is empty for filter")
	}
	operator := selection.Operator(op)
	sel := labels.NewSelector()
	req, err := labels.NewRequirement(key, operator, []string{value})
	if err != nil {
		return nil, errors.New("failed to create new requirement in label filter")
	}
	sel.Add(*req)

	return client.MatchingLabelsSelector{Selector: sel}, nil
}

func getFilterKey(key string) string {
	switch key {
	// if key is hostname, it should be changed to kubernetes hostname label
	case blockdevice.HostName:
		return KubernetesHostNameLabel
	}
	return key
}
