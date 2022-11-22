package controller

import "github.com/aperturerobotics/controllerbus/directive"

// value contains an attached resolver value
type value struct {
	// id is the value id
	id uint32
	// val is the directive value
	val directive.Value
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
