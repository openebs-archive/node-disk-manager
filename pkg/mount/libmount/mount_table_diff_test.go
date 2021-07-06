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
	"reflect"
	"testing"
)

func TestMountTabDiff_getMountEntry(t *testing.T) {
	type args struct {
		source string
		id     int
	}

	requiredArgs := args{"src1", 1}
	requiredFs := &Filesystem{
		id:     requiredArgs.id,
		source: requiredArgs.source,
	}
	requiredMTDEntry := &MountTabDiffEntry{
		action: MountActionMount,
		newFs:  requiredFs,
	}

	tests := []struct {
		name string
		md   MountTabDiff
		want *MountTabDiffEntry
	}{
		{
			name: "no mount action entry present",
			md: []*MountTabDiffEntry{
				{
					action: MountActionMove,
					newFs:  requiredFs,
				},
				{
					action: MountActionRemount,
					newFs:  requiredFs,
				},
				{
					action: MountActionUmount,
					newFs:  requiredFs,
				},
			},
			want: nil,
		},
		{
			name: "no matching id",
			md: []*MountTabDiffEntry{
				{
					action: MountActionMount,
					newFs: &Filesystem{
						id:     2,
						source: "src1",
					},
				},
			},
			want: nil,
		},
		{
			name: "no matching source",
			md: []*MountTabDiffEntry{
				{
					action: MountActionMount,
					newFs: &Filesystem{
						id:     1,
						source: "src2",
					},
				},
			},
			want: nil,
		},
		{
			name: "action, source and id match",
			md: []*MountTabDiffEntry{
				{
					action: MountActionMount,
					newFs: &Filesystem{
						id:     1,
						source: "src2",
					},
				},
				requiredMTDEntry,
			},
			want: requiredMTDEntry,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.md.getMountEntry(requiredArgs.source, requiredArgs.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MountTabDiff.getMountEntry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateDiff(t *testing.T) {
	type args struct {
		oldTab *MountTab
		newTab *MountTab
	}

	dummyFs := NewFilesystem()
	dummyMountTab := &MountTab{}
	tests := []struct {
		name string
		args args
		want MountTabDiff
	}{
		{
			name: "both args nil",
			args: args{nil, nil},
			want: NewMountTabDiff(),
		},
		{
			name: "both tabs empty",
			args: args{&MountTab{}, &MountTab{}},
			want: NewMountTabDiff(),
		},
		{
			name: "oldTab nil",
			args: args{nil, &MountTab{entries: []*Filesystem{dummyFs}}},
			want: MountTabDiff{{action: MountActionMount, oldFs: nil, newFs: dummyFs}},
		},
		{
			name: "new mounts only",
			args: args{dummyMountTab, &MountTab{entries: []*Filesystem{dummyFs}}},
			want: MountTabDiff{{action: MountActionMount, oldFs: nil, newFs: dummyFs}},
		},
		{
			name: "newFs nil",
			args: args{&MountTab{entries: []*Filesystem{dummyFs}}, nil},
			want: MountTabDiff{{action: MountActionUmount, oldFs: dummyFs, newFs: nil}},
		},
		{
			name: "all fs unmounted",
			args: args{&MountTab{entries: []*Filesystem{dummyFs}}, dummyMountTab},
			want: MountTabDiff{{action: MountActionUmount, oldFs: dummyFs, newFs: nil}},
		},

		// TODO: Add test for other types of changes
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateDiff(tt.args.oldTab, tt.args.newTab); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateDiff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMountTabDiffEntry_GetAction(t *testing.T) {
	type fields struct {
		action MountAction
	}
	tests := []struct {
		name   string
		fields fields
		want   MountAction
	}{
		{
			name:   "mount",
			fields: fields{action: MountActionMount},
			want:   MountActionMount,
		},
		{
			name:   "umount",
			fields: fields{action: MountActionUmount},
			want:   MountActionUmount,
		},
		{
			name:   "move",
			fields: fields{action: MountActionMove},
			want:   MountActionMove,
		},
		{
			name:   "remount",
			fields: fields{action: MountActionRemount},
			want:   MountActionRemount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mde := &MountTabDiffEntry{
				action: tt.fields.action,
			}
			if got := mde.GetAction(); got != tt.want {
				t.Errorf("MountTabDiffEntry.GetAction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMountTabDiffEntry_GetOldFs(t *testing.T) {
	type fields struct {
		oldFs *Filesystem
	}

	tests := []struct {
		name   string
		fields fields
		want   *Filesystem
	}{
		{
			name:   "oldFs nil",
			fields: fields{nil},
			want:   nil,
		},
		{
			name:   "sanity check",
			fields: fields{&Filesystem{id: 1}},
			want:   &Filesystem{id: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mde := &MountTabDiffEntry{
				oldFs: tt.fields.oldFs,
			}
			if got := mde.GetOldFs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MountTabDiffEntry.GetOldFs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMountTabDiffEntry_GetNewFs(t *testing.T) {
	type fields struct {
		newFs *Filesystem
	}
	tests := []struct {
		name   string
		fields fields
		want   *Filesystem
	}{
		{
			name:   "newFs nil",
			fields: fields{nil},
			want:   nil,
		},
		{
			name:   "sanity check",
			fields: fields{&Filesystem{id: 1}},
			want:   &Filesystem{id: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mde := &MountTabDiffEntry{
				newFs: tt.fields.newFs,
			}
			if got := mde.GetNewFs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MountTabDiffEntry.GetNewFs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMountTabDiff_ListSources(t *testing.T) {
	fs1 := NewFilesystem()
	fs2 := NewFilesystem()
	fs3 := NewFilesystem()

	fs1.SetSource("src1")
	fs2.SetSource("src2")
	fs3.SetSource("src3")

	tests := []struct {
		name string
		md   MountTabDiff
		want []string
	}{
		{
			name: "list sources",
			md: MountTabDiff{
				{
					action: MountActionMount,
					newFs:  fs1,
				},
				{
					action: MountActionUmount,
					oldFs:  fs1,
				},
				{
					action: MountActionMove,
					newFs:  fs2,
					oldFs:  fs2,
				},
				{
					action: MountActionRemount,
					newFs:  fs3,
					oldFs:  fs3,
				},
			},
			want: []string{"src1", "src2", "src3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.md.ListSources(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MountTabDiff.ListSources() = %v, want %v", got, tt.want)
			}
		})
	}
}
