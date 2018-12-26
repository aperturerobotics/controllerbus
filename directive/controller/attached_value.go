package controller

import (
	"github.com/aperturerobotics/controllerbus/directive"
)

// attachedValue implements directive AttachedValue.
type attachedValue struct {
	id uint32
	v  directive.Value
}

func newAttachedValue(id uint32, v directive.Value) *attachedValue {
	return &attachedValue{id: id, v: v}
}

// GetValueID returns the value ID.
func (a *attachedValue) GetValueID() uint32 {
	return a.id
}

// GetValue returns the value.
func (a *attachedValue) GetValue() directive.Value {
	return a.v
}

// _ is a type assertion
var _ directive.AttachedValue = ((*attachedValue)(nil))
