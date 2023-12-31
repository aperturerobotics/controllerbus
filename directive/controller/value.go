package controller

import (
	"sync/atomic"

	"github.com/aperturerobotics/controllerbus/directive"
)

// value contains an attached resolver value
type value struct {
	// id is the value id
	id uint32
	// val is the directive value
	val directive.Value
	// removeCallbackCtr is the counter for remove callback id
	removeCallbackCtr uint32
	// removeCallbacks is a set of callbacks to call when removed
	removeCallbacks []*valueRemoveCallback
}

// valueRemoveCallback is a callback to call when value is removed
type valueRemoveCallback struct {
	released atomic.Bool
	id       uint32
	cb       func()
}

// GetValueID returns the value ID.
func (v *value) GetValueID() uint32 {
	return v.id
}

// GetValue returns the value.
func (v *value) GetValue() directive.Value {
	return v.val
}

// _ is a type assertion
var _ directive.AttachedValue = ((*value)(nil))
