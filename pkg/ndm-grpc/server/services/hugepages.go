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

package services

import (
	"context"
	"io/ioutil"
	"strconv"
	"strings"

	protos "github.com/openebs/node-disk-manager/pkg/ndm-grpc/protos/ndm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

// SetHugepages service can set 2MB hugepages on a node
func (n *Node) SetHugepages(ctx context.Context, h *protos.Hugepages) (*protos.HugepagesResult, error) {

	//Note: Calling this method doesn't gurantee that the said number of pages will be set.
	// This is because OS might not have the demanded memory. It would be best to check if this is satisfied with GetHugePages()

	klog.Info("Setting Hugepages")

	hugepages := protos.Hugepages{
		Pages: h.Pages,
	}

	msg := []byte(strconv.Itoa(int(hugepages.Pages)))
	err := ioutil.WriteFile("/sys/kernel/mm/hugepages/hugepages-2048kB/nr_hugepages", msg, 0644)
	if err != nil {
		klog.Errorf("Error setting huge pages: %v", err)
		return nil, status.Errorf(codes.Internal, "Error setting hugepages")
	}

	return &protos.HugepagesResult{Result: true}, nil
}

// GetHugepages services gets the number of hugepages on a node
func (n *Node) GetHugepages(ctx context.Context, null *protos.Null) (*protos.Hugepages, error) {

	klog.Info("Getting the number of hugepages")

	hugepages, err := ioutil.ReadFile("/sys/kernel/mm/hugepages/hugepages-2048kB/nr_hugepages")
	if err != nil {
		klog.Errorf("Error fetching number of hugepages %v", err)
		return nil, status.Errorf(codes.Internal, "Error fetching the number of hugepages set on the node")
	}

	pages, err := strconv.Atoi(strings.TrimRight(string(hugepages), "\n"))
	if err != nil {
		klog.Errorf("Error converting number of hugepages %v", err)
		return nil, status.Errorf(codes.Internal, "Error converting the number of hugepages set on the node")
	}

	return &protos.Hugepages{
		Pages: int32(pages),
	}, nil
}
