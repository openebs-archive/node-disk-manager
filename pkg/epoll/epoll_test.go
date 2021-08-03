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

package epoll

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ep, err := New()
	if err != nil {
		assert.Equal(t, Epoll{}, ep)
	}
	ep.Close()
}

func TestStart(t *testing.T) {
	_, err := New()
	if err != nil {
		t.Fatal("failed to create epoll instance: ", err)
	}
}

func TestAddWatcher(t *testing.T) {
	ep, err := New()
	if err != nil {
		t.Fatal("failed to create epoll instance: ", err)
	}
	t.Cleanup(ep.Close)
	testfileName := "/proc/self/mounts"
	eventTypes := []EventType{EPOLLERR, EPOLLIN}
	// Note: order of test cases matters.
	testCases := []struct {
		name        string
		watcher     Watcher
		expectErr   bool
		expectedErr error
	}{
		{
			name: "add a new valid watcher",
			watcher: Watcher{
				FileName:   testfileName,
				EventTypes: eventTypes,
			},
			expectErr:   false,
			expectedErr: nil,
		},
		{
			name: "invalid event type",
			watcher: Watcher{
				FileName:   testfileName,
				EventTypes: []EventType{EPOLLERR, 10},
			},
			expectErr:   true,
			expectedErr: ErrInvalidEventType,
		},
		{
			name: "file already being watched",
			watcher: Watcher{
				FileName:   testfileName,
				EventTypes: eventTypes,
			},
			expectErr:   true,
			expectedErr: ErrFileAlreadyWatched,
		},
		{
			name: "file not present",
			watcher: Watcher{
				FileName:   filepath.Join(t.TempDir(), "non-existent"),
				EventTypes: eventTypes,
			},
			expectErr:   true,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ep.AddWatcher(tc.watcher)
			if tc.expectErr {
				assert.NotNil(t, err)
				if tc.expectedErr != nil {
					assert.Equal(t, tc.expectedErr, err)
				}
			} else {
				assert.Nil(t, err)
			}
		})
	}

}

func TestDeleteWatcher(t *testing.T) {
	ep, err := New()
	if err != nil {
		t.Fatal("failed to create epoll instance: ", err)
	}
	t.Cleanup(ep.Close)
	fileName := "/proc/self/mounts"
	watcher := Watcher{
		FileName:   fileName,
		EventTypes: []EventType{EPOLLIN},
	}

	t.Run("watcher not present", func(t *testing.T) {
		if ep.DeleteWatcher(fileName) != ErrWatcherNotFound {
			t.Fail()
		}
	})

	if ep.AddWatcher(watcher) != nil {
		t.Fatal("failed to add watcher")
	}

	t.Run("delete watcher", func(t *testing.T) {
		if ep.DeleteWatcher(fileName) != nil {
			t.Fail()
		}
	})
}
