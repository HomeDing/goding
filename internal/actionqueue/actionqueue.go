// actionqueue.go - In-memory FIFO action queue
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package actionQueue implements a simple in-process FIFO queue for action strings.

// For queuing and dispatching actions more functionality like look-ahead is required so
// the build-in go channel mechanism cannot be used and the
// [/internal/actionqueue](/internal/actionqueue/actionqueue.go) is implementing a similar
// mechanism FIFO mechanism especially for actions.

package actionQueue

import (
	"log/slog"
	"strings"
	"sync"

	"github.com/HomeDing/goding/internal/elements/registry"
)

// Package-level internal state for the action queue.
//
// This is a simple, in-memory, FIFO queue for action strings. It is
// intended for short-lived process-local use and is protected by a mutex
// for basic concurrent access from multiple goroutines. It does not
// provide persistence, bounds checking, or blocking semantics — callers
// should handle those concerns if needed.
var (
	// mu guards the queue from concurrent access by multiple goroutines.
	mu sync.Mutex
	// queue is the static FIFO buffer for action strings.
	queue []string
	// workerOnce starts the dispatcher only when the first action is queued.
	workerOnce sync.Once
	// wake signals the dispatcher that the queue may contain work.
	wake = make(chan struct{}, 1)
)

// Add appends a new action to the FIFO buffer for asynchronous dispatch.
//
// This function is safe for concurrent use by multiple goroutines. It
// preserves FIFO ordering and starts the dispatcher when needed.
func Add(action string) {
	mu.Lock()
	// Append preserves FIFO order.
	queue = append(queue, action)
	mu.Unlock()

	// start the worker only once, when the first action is added. Subsequent calls will not start additional workers.
	workerOnce.Do(func() {
		go dispatchWorker()
	})

	// Wake the worker without blocking if it is already scheduled to run.
	select {
	case wake <- struct{}{}:
	default:
	}
}

// AddOnce removes queued actions for the same target before appending action.
// The target is the portion before '?', so parameter changes do not create
// additional pending actions for the same element. It then schedules dispatch.
func AddOnce(action string) {
	target := actionTarget(action)

	mu.Lock()
	queue = removeActionsForTarget(queue, target)
	mu.Unlock()
	Add(action)
}

func actionTarget(action string) string {
	target, _, _ := strings.Cut(action, "?")
	return target
}

// removeActionsForTarget removes all actions from the queue that match the given target.
// to avoid repeated actions for the same target.
// It returns a new slice with only the actions that do not match the target.
func removeActionsForTarget(actions []string, target string) []string {
	kept := actions[:0]
	for _, action := range actions {
		if actionTarget(action) != target {
			kept = append(kept, action)
		}
	}
	clear(actions[len(kept):])
	return kept
} // removeActionsForTarget()

// dispatchWorker is a long-running goroutine that processes actions from the queue.
// It waits for new actions to be added and dispatches them in FIFO order.
// The worker will continue running until the program exits, but it will
// only wake up when there are actions to process.
func dispatchWorker() {
	for {
		action, ok := GetNext()
		if !ok {
			<-wake
			continue
		}
		DispatchNow(action)
	}
} // dispatchWorker()

// GetNext removes and returns the oldest action from the buffer.
//
// Returns the action string and true if an element was available, or an
// empty string and false if the buffer is empty. This operation is safe
// for concurrent use. Note: removal re-slices the underlying slice which
// may retain the backing array; if the queue grows large and memory
// retention is a concern consider copying or using a ring buffer.
func GetNext() (string, bool) {
	mu.Lock()
	defer mu.Unlock()

	if len(queue) == 0 {
		return "", false
	}

	// Retrieve the first element and remove it from the slice.
	action := queue[0]
	queue = queue[1:]

	return action, true
}

// TODO: create test cases for the action queue functions to ensure correct behavior under concurrent access and edge cases.
func GetlatestFor(target string) (string, bool) {
	mu.Lock()
	defer mu.Unlock()

	// Iterate backwards to find the latest action for the target.
	for i := len(queue) - 1; i >= 0; i-- {
		action := queue[i]
		if strings.HasPrefix(action, target+"/") {
			return action, true
		}
	}

	return "", false
}

// action : type/id?key=value
func DispatchNow(action string) {
	slog.Debug("DispatchNow", slog.String("action", action))
	var action1, action2 string
	// var err error
	var found bool

	action1, action2, found = strings.Cut(action, "?")
	if !found {
		return
	}

	eType, eID, found := strings.Cut(action1, "/")
	if !found {
		return
	}

	e := registry.Find(eType, eID)
	if e == nil {
		slog.Warn("element not found", slog.String("t", eType), slog.String("id", eID))
		return
	}

	if len(action2) > 0 {
		aKey, aValue, found := strings.Cut(action2, "=")
		if found {
			e.Set(aKey, aValue)
		}
	}
}
