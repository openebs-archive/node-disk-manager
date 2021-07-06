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
	"math/rand"
	"reflect"
	"testing"
)

func TestMountTab_applyFilters(t *testing.T) {
	type fields struct {
		format       MountTabFormat
		entries      []*Filesystem
		allowFilters []FsFilter
		denyFilters  []FsFilter
	}
	type args struct {
		fs *Filesystem
	}

	allowAll := allowAllFilter()
	denyAll := denyAllFilter()
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "no apply/deny filters",
			want: true,
		},
		{
			name: "allow filters only, fs allowed",
			fields: fields{
				allowFilters: []FsFilter{allowAll},
			},
			want: true,
		},
		{
			name: "allow filters only, fs not allowed",
			fields: fields{
				allowFilters: []FsFilter{denyAll},
			},
			want: false,
		},
		{
			name: "deny filters only, fs denied",
			fields: fields{
				denyFilters: []FsFilter{allowAll},
			},
			want: false,
		},
		{
			name: "deny filters only, fs not denied",
			fields: fields{
				denyFilters: []FsFilter{denyAll},
			},
			want: true,
		},
		{
			name: "both allow and deny filters, fs allowed and denied",
			fields: fields{
				allowFilters: []FsFilter{allowAll},
				denyFilters:  []FsFilter{allowAll},
			},
			want: false,
		},
		{
			name: "both allow and deny filters, fs allowed but not denied",
			fields: fields{
				allowFilters: []FsFilter{allowAll},
				denyFilters:  []FsFilter{denyAll},
			},
			want: true,
		},
		{
			name: "both allow and deny filters, fs denied but not allowed",
			fields: fields{
				allowFilters: []FsFilter{denyAll},
				denyFilters:  []FsFilter{allowAll},
			},
			want: false,
		},
		{
			name: "both allow and deny filters, fs neither allowed nor denied",
			fields: fields{
				allowFilters: []FsFilter{denyAll},
				denyFilters:  []FsFilter{denyAll},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := &MountTab{
				format:       tt.fields.format,
				entries:      tt.fields.entries,
				allowFilters: tt.fields.allowFilters,
				denyFilters:  tt.fields.denyFilters,
			}
			if got := mt.applyFilters(tt.args.fs); got != tt.want {
				t.Errorf("MountTab.applyFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMountTab_AddFilesystem(t *testing.T) {
	type fields struct {
		format       MountTabFormat
		entries      []*Filesystem
		allowFilters []FsFilter
		denyFilters  []FsFilter
	}
	type args struct {
		fs *Filesystem
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "filesystem arg nil",
			args:    args{nil},
			wantErr: true,
		},
		{
			name:    "filesystem busy",
			args:    args{&Filesystem{tab: &MountTab{}}},
			wantErr: true,
		},
		{
			name: "filesystem denied by filters",
			fields: fields{
				denyFilters: []FsFilter{allowAllFilter()},
			},
			args:    args{NewFilesystem()},
			wantErr: true,
		},
		{
			name:    "filesystem accepted and added",
			args:    args{NewFilesystem()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := &MountTab{
				format:       tt.fields.format,
				entries:      tt.fields.entries,
				allowFilters: tt.fields.allowFilters,
				denyFilters:  tt.fields.denyFilters,
			}
			if err := mt.AddFilesystem(tt.args.fs); (err != nil) != tt.wantErr {
				t.Errorf("MountTab.AddFilesystem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMountTab_Size(t *testing.T) {
	type fields struct {
		format       MountTabFormat
		entries      []*Filesystem
		allowFilters []FsFilter
		denyFilters  []FsFilter
	}
	// get a random length for the slice
	randSize := rand.Intn(10)
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "random size",
			fields: fields{
				entries: make([]*Filesystem, randSize),
			},
			want: randSize,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := &MountTab{
				format:       tt.fields.format,
				entries:      tt.fields.entries,
				allowFilters: tt.fields.allowFilters,
				denyFilters:  tt.fields.denyFilters,
			}
			if got := mt.Size(); got != tt.want {
				t.Errorf("MountTab.Size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMountTab_Find(t *testing.T) {
	type fields struct {
		format       MountTabFormat
		entries      []*Filesystem
		allowFilters []FsFilter
		denyFilters  []FsFilter
	}
	type args struct {
		filters []FsFilter
	}

	fs := NewFilesystem()
	fsEntries := []*Filesystem{fs}
	matchAll := allowAllFilter()
	matchNone := denyAllFilter()
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Filesystem
	}{
		{
			name:   "not matching fs present",
			fields: fields{entries: fsEntries},
			args:   args{filters: []FsFilter{matchNone}},
			want:   nil,
		},
		{
			name:   "matching fs present",
			fields: fields{entries: fsEntries},
			args:   args{filters: []FsFilter{matchAll}},
			want:   fs,
		},
		{
			name:   "multiple filters passed, one or more filters doen't match",
			fields: fields{entries: fsEntries},
			args:   args{filters: []FsFilter{matchAll, matchNone, matchAll}},
			want:   nil,
		},
		{
			name: "fs entries empty",
			args: args{filters: []FsFilter{matchAll}},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := &MountTab{
				format:       tt.fields.format,
				entries:      tt.fields.entries,
				allowFilters: tt.fields.allowFilters,
				denyFilters:  tt.fields.denyFilters,
			}
			if got := mt.Find(tt.args.filters...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MountTab.Find() = %v, want %v", got, tt.want)
			}
		})
	}
}

func allowAllFilter() FsFilter {
	return func(f *Filesystem) bool {
		return true
	}
}

func denyAllFilter() FsFilter {
	return func(f *Filesystem) bool {
		return false
	}
}
