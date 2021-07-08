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

// Filesystem represents a single filesystem entry in a mount table.
type Filesystem struct {
	tab      *MountTab
	id       int
	source   string
	tagName  string
	tagValue string
	target   string
	fsType   string
	vfsOpts  string
	fsOpts   string
}

type FsFilter func(*Filesystem) bool

// NewFilesystem returns an empty filesystem
func NewFilesystem() *Filesystem {
	return &Filesystem{}
}

// GetID returns the id of the filesystem
func (fs *Filesystem) GetID() int {
	return fs.id
}

// GetSource returns the source device of the filesystem
func (fs *Filesystem) GetSource() string {
	return fs.source
}

// GetTarget returns the target path of the filesystem
func (fs *Filesystem) GetTarget() string {
	return fs.target
}

// GetVFSOptions returns the vfs options of the filesystem
func (fs *Filesystem) GetVFSOptions() string {
	return fs.vfsOpts
}

// GetFsOptions returns the fs options of the filesystem
func (fs *Filesystem) GetFSOptions() string {
	return fs.fsOpts
}

// SetTag sets the tag for the filesystem
func (fs *Filesystem) SetTag(name string, value string) {
	fs.tagName = name
	fs.tagValue = value
}

// SetSource sets the source device of the filesystem
func (fs *Filesystem) SetSource(src string) {
	fs.source = src
}

// SetTarget sets the target path of the filesystem
func (fs *Filesystem) SetTarget(target string) {
	fs.target = target
}

// SetFsType sets the filsystem type of the filesystem
// Eg: ext4, vfat, etc.
func (fs *Filesystem) SetFsType(fsType string) {
	fs.fsType = fsType
}

// GetMountTable returns a pointer to the mount tab this filesystem
// is a part of. nil if the filesystem isn't a part of any mount tab.
func (fs *Filesystem) GetMountTable() *MountTab {
	return fs.tab
}

// SetMountTable sets the mount tab that this filesystem is a part of.
// A filesystem can only be a part of a single mount tab.
func (fs *Filesystem) SetMountTable(tab *MountTab) {
	fs.tab = tab
}

// SourceFilter returns a filter that can be used to filter filesystems
// by matching the source value.
func SourceFilter(source string) FsFilter {
	return func(f *Filesystem) bool {
		return f != nil && f.source == source
	}
}

// TargetFilter returns a filter that can be used to filter filesystems
// by matching target values.
func TargetFilter(target string) FsFilter {
	return func(f *Filesystem) bool {
		return f != nil && f.target == target
	}
}

// IDFilter returns a filter that can be used to filter filesystems
// by matching id values.
func IDFilter(id int) FsFilter {
	return func(f *Filesystem) bool {
		return f != nil && f.id == id
	}
}

// TargetContainsFilter returns a filter that can be used to filter filesystems
// by checking if the target contains the given substring.
func TargetContainsFilter(substr string) FsFilter {
	return func(f *Filesystem) bool {
		return f != nil && strings.Contains(f.target, substr)
	}
}

// SourceContainsFilter returns a filter that can be used to filter filesystems
// by checking if the source contains the given substring.
func SourceContainsFilter(substr string) FsFilter {
	return func(f *Filesystem) bool {
		return f != nil && strings.Contains(f.source, substr)
	}
}
