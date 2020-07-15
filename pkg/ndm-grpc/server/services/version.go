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
	protos "github.com/openebs/node-disk-manager/pkg/ndm-grpc/protos/ndm"
	"github.com/openebs/node-disk-manager/pkg/ndm-grpc/server"
	"github.com/openebs/node-disk-manager/pkg/version"
	"k8s.io/klog"

	"context"
)

// Info lets types defined in Server used
type Info struct {
	server.Info
}

// NewInfo is a constructor
func NewInfo() *Info {
	return &Info{server.Info{}}
}

// FindVersion detects the version and gitCommit of NDM
func (i *Info) FindVersion(ctx context.Context, null *protos.Null) (*protos.VersionInfo, error) {

	klog.Infof("Print Version : %v , commit hash : %v", version.GetVersion(), version.GetGitCommit())

	return &protos.VersionInfo{Version: version.GetVersion(), GitCommit: version.GetGitCommit()}, nil

}
