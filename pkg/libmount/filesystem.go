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

package libmount

import "strings"

type Filesystem struct {
	tab      *MountTab
	id       int
	parent   int
	devno    uint32
	bindSrc  string
	source   string
	tagName  string
	tagValue string
	root     string
	target   string
	fsType   string
	options  string
	vfsOpts  string
	fsOpts   string
	/* skipped fields
	char		*opt_fields;
	char		*user_optstr;
	*/
	attrs  string
	freq   int
	passNo int

	swapType string
	size     int
	usedSize int
	priority int
	flags    int
	tid      int
	comment  string
	// void *userdata

}

type FsFilter func(*Filesystem) bool

func NewFilesystem() *Filesystem {
	return &Filesystem{}
}

func (fs *Filesystem) GetID() int {
	return fs.id
}

func (fs *Filesystem) GetSource() string {
	return fs.source
}

func (fs *Filesystem) GetTarget() string {
	return fs.target
}

func (fs *Filesystem) GetVFSOptions() string {
	return fs.vfsOpts
}

func (fs *Filesystem) GetFSOptions() string {
	return fs.fsOpts
}

func (fs *Filesystem) SetTag(name string, value string) {
	fs.tagName = name
	fs.tagValue = value
}

func (fs *Filesystem) SetSource(src string) {
	fs.source = src
}

func (fs *Filesystem) SetTarget(target string) {
	fs.target = target
}

func (fs *Filesystem) SetFsType(fsType string) {
	fs.fsType = fsType
}

func (fs *Filesystem) GetMountTable() *MountTab {
	return fs.tab
}

func (fs *Filesystem) SetMountTable(tab *MountTab) {
	fs.tab = tab
}

func SourceFilter(source string) FsFilter {
	return func(f *Filesystem) bool {
		return f != nil && f.source == source
	}
}

func TargetFilter(target string) FsFilter {
	return func(f *Filesystem) bool {
		return f != nil && f.target == target
	}
}

func IDFilter(id int) FsFilter {
	return func(f *Filesystem) bool {
		return f != nil && f.id == id
	}
}

func TargetContainsFilter(substr string) FsFilter {
	return func(f *Filesystem) bool {
		return f != nil && strings.Contains(f.target, substr)
	}
}
