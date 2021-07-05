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

type MountAction uint8

const (
	MountActionMount MountAction = iota
	MountActionUmount
	MountActionMove
	MountActionRemount
)

type MountTabDiffEntry struct {
	oldFs  *Filesystem
	newFs  *Filesystem
	action MountAction
}

type MountTabDiff []*MountTabDiffEntry

func NewMountTabDiff() MountTabDiff {
	tabDiff := make(MountTabDiff, 0)
	return tabDiff
}

func (md MountTabDiff) AddDiffEntry(oldFs *Filesystem, newFs *Filesystem, action MountAction) MountTabDiff {
	return append(md, &MountTabDiffEntry{oldFs, newFs, action})
}

func (md MountTabDiff) getMountEntry(source string, id int) *MountTabDiffEntry {
	for _, diffEntry := range md {
		if diffEntry.action == MountActionMount &&
			diffEntry.newFs != nil &&
			diffEntry.newFs.GetID() == id &&
			diffEntry.newFs.GetSource() == source {
			return diffEntry
		}
	}
	return nil
}

func GenerateDiff(oldTab *MountTab, newTab *MountTab) MountTabDiff {
	diffTable := NewMountTabDiff()
	if oldTab == nil {
		oldTab = &MountTab{}
	}
	if newTab == nil {
		newTab = &MountTab{}
	}
	oldTabSize := oldTab.Size()
	newTabSize := newTab.Size()

	// Both tables empty
	if newTabSize == 0 && oldTabSize == 0 {
		return diffTable
	}
	// Old table empty => all entries in new table are new mounts
	if oldTabSize == 0 {
		for _, entry := range newTab.entries {
			diffTable = diffTable.AddDiffEntry(nil, entry, MountActionMount)
		}
		return diffTable
	}
	// New table empty => all entries in old table were unmounted
	if newTabSize == 0 {
		for _, entry := range oldTab.entries {
			diffTable = diffTable.AddDiffEntry(entry, nil, MountActionUmount)
		}
		return diffTable
	}

	// Search newly mounted or modified entries
	for _, newFs := range newTab.entries {
		oldFs := oldTab.Find(SourceFilter(newFs.GetSource()), TargetFilter(newFs.GetTarget()))
		if oldFs == nil {
			diffTable = diffTable.AddDiffEntry(nil, newFs, MountActionMount)
		} else if oldFs.GetVFSOptions() != newFs.GetVFSOptions() ||
			oldFs.GetFSOptions() != newFs.GetFSOptions() {
			diffTable = diffTable.AddDiffEntry(oldFs, newFs, MountActionRemount)
		}
	}

	// Search umounted or moved entries
	for _, oldFs := range oldTab.entries {
		newFs := newTab.Find(SourceFilter(oldFs.GetSource()), TargetFilter(oldFs.GetTarget()))
		if newFs == nil {
			de := diffTable.getMountEntry(oldFs.GetSource(), oldFs.GetID())
			// fs umounted
			if de == nil {
				diffTable = diffTable.AddDiffEntry(oldFs, nil, MountActionUmount)
			} else {
				// else moved
				de.oldFs = oldFs
				de.action = MountActionMove
			}
		}
	}

	return diffTable
}

func (mde *MountTabDiffEntry) GetAction() MountAction {
	return mde.action
}

func (mde *MountTabDiffEntry) GetOldFs() *Filesystem {
	if mde.oldFs == nil {
		return nil
	}
	fs := *mde.oldFs
	return &fs
}

func (mde *MountTabDiffEntry) GetNewFs() *Filesystem {
	if mde.newFs == nil {
		return nil
	}
	fs := *mde.newFs
	return &fs
}
