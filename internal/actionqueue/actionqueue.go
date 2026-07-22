package actionQueue

import (
	"sync"
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
)

// Add appends a new action to the FIFO buffer.
//
// This function is safe for concurrent use by multiple goroutines. It
// performs a simple append which preserves FIFO ordering.
func Add(action string) {
	mu.Lock()
	defer mu.Unlock()

	// Append preserves FIFO order.
	queue = append(queue, action)
}

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