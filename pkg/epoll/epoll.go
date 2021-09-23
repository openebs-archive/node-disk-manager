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
	"errors"
	"os"
	"syscall"

	"k8s.io/klog"
)

const (
	EPOLLIN  EventType = syscall.EPOLLIN
	EPOLLPRI EventType = syscall.EPOLLPRI
	EPOLLERR EventType = syscall.EPOLLERR
	EPOLLOUT EventType = syscall.EPOLLOUT
	EPOLLHUP EventType = syscall.EPOLLHUP

	// timeout is the duration in milliseconds after which the EpollWait system
	// call returns even if there are no events. Setting it to -1 makes the call
	// return only when there is an event.
	timeout int = -1
)

var (
	ErrFileAlreadyWatched error = errors.New("file being watched already")
	ErrInvalidEventType   error = errors.New("invalid event type")
	ErrWatcherNotFound    error = errors.New("watcher not found")
)

type Epoll struct {
	watchers      []Watcher
	fileToWatcher map[string]*Watcher
	fdToWatcher   map[int32]*Watcher
	eventChan     chan Event
	eventChanSize int
	epfd          int
	active        bool
}

type Event struct {
	fileName string
	events   uint32
}

type EventType uint32
type Watcher struct {
	fd *os.File
	// The complete path of the file to watch. The file name is also used to
	// uniquely identify the watcher
	FileName string
	// The event types to watch for
	EventTypes []EventType
}

type NewOpt func(*Epoll)

func BufferSize(size int) NewOpt {
	return func(e *Epoll) {
		e.eventChanSize = size
	}
}

func New(opts ...NewOpt) (Epoll, error) {
	epfd, err := syscall.EpollCreate(1)
	if err != nil {
		return Epoll{}, err
	}
	e := Epoll{
		watchers:      make([]Watcher, 0),
		fileToWatcher: make(map[string]*Watcher),
		fdToWatcher:   make(map[int32]*Watcher),
		epfd:          epfd,
		eventChanSize: 0,
	}
	for _, opt := range opts {
		opt(&e)
	}
	return e, nil
}

func (e *Epoll) Start() (<-chan Event, error) {
	if e.epfd == 0 {
		return nil, errors.New("invalid epoll struct")
	}

	if e.active {
		return nil, errors.New("epoll already started")
	}
	e.eventChan = make(chan Event, e.eventChanSize)
	e.active = true
	go e.listen()
	return e.eventChan, nil
}

func (e *Epoll) Stop() {
	if !e.active {
		return
	}
	e.active = false
	close(e.eventChan)
}

func (e *Epoll) Close() {
	e.Stop()
	syscall.Close(e.epfd)
	for i := range e.watchers {
		e.watchers[i].fd.Close()
	}
	e.fdToWatcher = nil
	e.fileToWatcher = nil
	e.eventChan = nil
	e.watchers = nil
}

func (e *Epoll) AddWatcher(w Watcher) error {
	for _, evt := range w.EventTypes {
		if !isValidEventType(evt) {
			return ErrInvalidEventType
		}
	}

	if _, ok := e.fileToWatcher[w.FileName]; ok {
		return ErrFileAlreadyWatched
	}
	fd, err := os.Open(w.FileName)
	if err != nil {
		return err
	}
	w.fd = fd
	if err = e.addToEpoll(fd, w.EventTypes); err != nil {
		return err
	}

	e.watchers = append(e.watchers, w)
	n := len(e.watchers) - 1
	e.fileToWatcher[w.FileName] = &e.watchers[n]
	e.fdToWatcher[int32(fd.Fd())] = &e.watchers[n]
	return nil
}

func (e *Epoll) DeleteWatcher(fileName string) error {
	w, ok := e.fileToWatcher[fileName]
	if !ok {
		return ErrWatcherNotFound
	}
	err := e.deleteFromEpoll(w.fd)
	if err != nil {
		return err
	}
	delete(e.fileToWatcher, fileName)
	delete(e.fdToWatcher, int32(w.fd.Fd()))
	return nil
}

func (e *Epoll) addToEpoll(fd *os.File, events []EventType) error {
	ifd := int32(fd.Fd())
	var evs EventType
	for _, e := range events {
		evs |= e
	}
	return syscall.EpollCtl(e.epfd, syscall.EPOLL_CTL_ADD, int(ifd),
		&syscall.EpollEvent{
			Events: uint32(evs),
			Fd:     ifd,
		})
}

func (e *Epoll) deleteFromEpoll(fd *os.File) error {
	return syscall.EpollCtl(e.epfd, syscall.EPOLL_CTL_DEL, int(fd.Fd()), nil)
}

func (e *Epoll) dispatchEvent(ev *syscall.EpollEvent) {
	w, ok := e.fdToWatcher[ev.Fd]
	// Skip if no watcher found
	if !ok {
		return
	}
	klog.V(4).Infof("epoll event for file %s dispatched", w.FileName)

	e.eventChan <- Event{
		fileName: w.FileName,
		events:   ev.Events,
	}
}

func (e *Epoll) listen() {
	events := make([]syscall.EpollEvent, 2)
	for e.active {
		// TODO: Handle errors here
		klog.Info("waiting for epoll events...")
		count, _ := syscall.EpollWait(e.epfd, events, timeout)
		klog.Infof("received %d events from epoll. dispatching...", count)
		for i := 0; e.active && i < count; i++ {
			e.dispatchEvent(&events[i])
		}
	}
}

func (ev *Event) IsType(evt EventType) bool {
	return ev.events&uint32(evt) != 0
}

func (ev *Event) FileName() string {
	return ev.fileName
}

func isValidEventType(evt EventType) bool {
	switch evt {
	case EPOLLERR, EPOLLHUP, EPOLLIN, EPOLLOUT, EPOLLPRI:
		return true
	}

	return false
}
