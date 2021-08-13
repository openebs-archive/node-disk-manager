/*
Copyright 2018 OpenEBS Authors.

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

package udevevent

import (
	"reflect"
	"testing"
)

func TestNewMonitor(t *testing.T) {
	monitor, err := newMonitor()
	if err != nil {
		t.Error(err)
	}
	defer monitor.free()
	if monitor.udev == nil {
		t.Errorf("udev should not be nil")
	}
	if monitor.udevMonitor == nil {
		t.Errorf("udevMonitor should not be nil")
	}
}

func TestSetup(t *testing.T) {
	monitor, err := newMonitor()
	if err != nil {
		t.Error(err)
	}
	defer monitor.free()
	fd, err := monitor.setup()
	if err != nil {
		t.Error(err)
	}
	if fd < 3 {
		t.Errorf("fd value should be greater than 2")
	}
}

func TestSubscribe(t *testing.T) {
	type args struct {
		eventTypes []UdevEventType
	}
	tests := []struct {
		name string
		args args
		want *Subscription
	}{
		{
			name: "no event types passed",
			want: &Subscription{
				subscribedTypes: []UdevEventType{EventTypeAdd,
					EventTypeRemove, EventTypeChange},
			},
		},
		{
			name: "event type passed",
			args: args{
				eventTypes: []UdevEventType{EventTypeChange},
			},
			want: &Subscription{
				subscribedTypes: []UdevEventType{EventTypeChange},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Subscribe(tt.args.eventTypes...)
			if !reflect.DeepEqual(got.subscribedTypes,
				tt.want.subscribedTypes) {
				t.Errorf("Subscribe() = %v, want %v", got, tt.want)
			}
			if len(subscriptions) != 1 {
				t.Error("subscriptions array not updated")
			}
			subscriptions = nil
		})
	}
}

func TestUnsubscribe(t *testing.T) {
	type args struct {
		sub *Subscription
	}

	dummySubscription := Subscription{
		targetChannel:   make(chan UdevEvent),
		subscribedTypes: []UdevEventType{EventTypeAdd},
	}

	tests := []struct {
		name          string
		args          args
		subscriptions []*Subscription
		wantErr       bool
	}{
		{
			name:          "valid subscription",
			subscriptions: []*Subscription{&dummySubscription},
			args:          args{&dummySubscription},
			wantErr:       false,
		},
		{
			name:    "invalid subscription, argument nil",
			wantErr: true,
		},
		{
			name: "invalid subscription, target channel nil",
			args: args{&Subscription{
				subscribedTypes: []UdevEventType{EventTypeAdd},
			}},
			wantErr: true,
		},
		{
			name: "invalid subscription, subscribed types nil",
			args: args{&Subscription{
				targetChannel: make(chan UdevEvent),
			}},
			wantErr: true,
		},
		{
			name:    "invalid subscription, all fields nil",
			args:    args{&Subscription{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscriptions = tt.subscriptions
			if err := Unsubscribe(tt.args.sub); (err != nil) != tt.wantErr {
				t.Errorf("Unsubscribe() error = %v, wantErr %v", err, tt.wantErr)
			}
			subscriptions = nil
		})
	}
}
