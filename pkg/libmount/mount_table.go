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

import (
	"bufio"
	"os"
)

const (
	MNT_FMT_GUESS MountTabFormat = iota
	MNT_FMT_FSTAB
	MNT_FMT_MOUNTINFO
	MNT_FMT_UTAB
	MNT_FMT_SWAPS
	MNT_FMT_MTAB = MNT_FMT_FSTAB
)

type MountTabFormat int
type MountTabOpt func(*MountTab) error

type MountTab struct {
	format MountTabFormat
	/* skipped entries
	int		comms;
	char		*comm_intro;
	char		*comm_tail;
	struct libmnt_cache *cache;
	void		*fltrcb_data;
	*/
	entries      []*Filesystem
	allowFilters []FsFilter
	denyFilters  []FsFilter
	// void		*userdata;
}

func NewMountTab(opts ...MountTabOpt) (*MountTab, error) {
	mt := MountTab{}
	for _, opt := range opts {
		opt(&mt)
	}
	return &mt, nil
}

func FromFile(fileName string, format MountTabFormat) MountTabOpt {
	return func(mt *MountTab) error {
		_, err := os.Stat(fileName)
		if err != nil {
			return err
		}
		mt.format = format
		return mt.parseFile(fileName)
	}
}

func WithAllowFilter(filter FsFilter) MountTabOpt {
	return func(mt *MountTab) error {
		mt.allowFilters = append(mt.allowFilters, filter)
		return nil
	}
}

func WithDenyFilter(filter FsFilter) MountTabOpt {
	return func(mt *MountTab) error {
		mt.denyFilters = append(mt.denyFilters, filter)
		return nil
	}
}

func (mt *MountTab) applyFilters(fs *Filesystem) bool {
	isAllowed := false
	isDenied := false

	if len(mt.allowFilters) == 0 {
		isAllowed = true
	}

	for _, filter := range mt.denyFilters {
		isDenied = isDenied || filter(fs)
	}

	for _, filter := range mt.allowFilters {
		isAllowed = isAllowed || filter(fs)
	}

	return !isDenied && isAllowed
}

func (mt *MountTab) AddFilesystem(fs *Filesystem) error {
	if fs == nil {
		// TODO: return EINVAL
		return nil
	}

	if fs.GetMountTable() != nil {
		// TODO: return EFSBUSY
		return nil
	}

	if !mt.applyFilters(fs) {
		// TODO: return denied by filter error
		return nil
	}

	mt.entries = append(mt.entries, fs)
	fs.SetMountTable(mt)
	return nil
}

func (mt *MountTab) Size() int {
	return len(mt.entries)
}

func (mt *MountTab) Find(filters ...FsFilter) *Filesystem {
	if len(filters) == 0 {
		return nil
	}
	for _, entry := range mt.entries {
		res := true
		for _, filter := range filters {
			res = res && filter(entry)
		}
		if res {
			return entry
		}
	}
	return nil
}

func (mt *MountTab) Entries() []*Filesystem {
	return mt.entries
}

func (mt *MountTab) parseFile(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	stream := bufio.NewScanner(file)
	parser := NewParser(mt.format)

	for stream.Scan() {
		line := stream.Text()
		fs, err := parser.Parse(line)
		if err != nil {
			// TODO: deal with error. Two possibilities - recoverable error, irrecoverable error
			return err
		}
		err = mt.AddFilesystem(fs)
		if err != nil {
			// TODO: treat this error as irrecoverable?
			return err
		}
	}
	return nil
}
