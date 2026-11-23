// devicechange.go - Shared device-change notification subscriptions
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package devicechange distributes system device-change notifications.
package devicechange

import "sync"

var subscribers = struct {
	sync.Mutex
	nextID uint64
	chans  map[uint64]chan struct{}
}{
	chans: make(map[uint64]chan struct{}),
}

// Subscribe registers for device-change notifications. The returned function
// removes the subscription; notifications may be coalesced.
func Subscribe() (<-chan struct{}, func()) {
	subscribers.Lock()
	subscribers.nextID++
	id := subscribers.nextID
	channel := make(chan struct{}, 1)
	subscribers.chans[id] = channel
	subscribers.Unlock()

	return channel, func() {
		subscribers.Lock()
		delete(subscribers.chans, id)
		subscribers.Unlock()
	}
}

// Notify broadcasts a device-change notification to current subscribers.
func Notify() {
	subscribers.Lock()
	channels := make([]chan struct{}, 0, len(subscribers.chans))
	for _, channel := range subscribers.chans {
		channels = append(channels, channel)
	}
	subscribers.Unlock()

	for _, channel := range channels {
		select {
		case channel <- struct{}{}:
		default:
		}
	}
}

// End.
