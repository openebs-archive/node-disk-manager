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

package services

import (
	"context"

	"github.com/openebs/node-disk-manager/api-service/node"
	"github.com/openebs/node-disk-manager/pkg/version"
	protos "github.com/openebs/node-disk-manager/spec/ndm"

	"k8s.io/klog"
)

// Info helps in using types defined in package Node
type Info struct {
	node.Info
}

// NewInfo returns an instance of type Node
func NewInfo() *Info {
	return &Info{node.Info{}}
}

// FindVersion detects the version and gitCommit of NDM
func (i *Info) FindVersion(ctx context.Context, null *protos.Null) (*protos.VersionInfo, error) {

	klog.V(4).Infof(" Version : %v , commit hash : %v", version.GetVersion(), version.GetGitCommit())

	return &protos.VersionInfo{Version: version.GetVersion(), GitCommit: version.GetGitCommit()}, nil

}
