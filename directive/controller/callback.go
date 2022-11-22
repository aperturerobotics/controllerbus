package controller

import (
	"sync/atomic"
)

// callback contains a func to call that can be released
type callback[T any] struct {
	// released indicates this was released
	released atomic.Bool
	// cb is the callback
	cb T
}

// newCallback constructs a new callback
func newCallback[T any](cb T) *callback[T] {
	return &callback[T]{cb: cb}
}

// removeFromCallbacks removes the callback from the list.
func removeFromCallbacks[T any](cbs []*callback[T], cb *callback[T]) []*callback[T] {
	for i := 0; i < len(cbs); i++ {
		if cbs[i] == cb {
			cbs = append(cbs[:i], cbs[i+1:]...)
			break
		}
	}
	return cbs
}
