package loader

import (
	"time"

	"github.com/aperturerobotics/controllerbus/controller"
)

// execControllerValue is the value type for ExecController
type execControllerValue struct {
	updatedTimestamp time.Time
	retryTimestamp   time.Time
	ctrl             controller.Controller
	err              error
}

// NewExecControllerValue builds a new ExecControllerValue
func NewExecControllerValue(
	updatedTimestamp time.Time,
	retryTimestamp time.Time,
	ctrl controller.Controller,
	err error,
) ExecControllerValue {
	return &execControllerValue{
		updatedTimestamp: updatedTimestamp,
		retryTimestamp:   retryTimestamp,
		ctrl:             ctrl,
		err:              err,
	}
}

// GetUpdatedTimestamp returns the last time this info changed.
func (v *execControllerValue) GetUpdatedTimestamp() time.Time {
	return v.updatedTimestamp
}

// GetNextRetryTimestamp returns the next time this controller will be attempted.
func (v *execControllerValue) GetNextRetryTimestamp() time.Time {
	return v.retryTimestamp
}

// GetController returns the current controller object.
func (v *execControllerValue) GetController() controller.Controller {
	return v.ctrl
}

// GetError returns the error running the controller.
// Controller may still be set in this case.
func (v *execControllerValue) GetError() error {
	return v.err
}

// _ is a type assertion
var _ ExecControllerValue = ((*execControllerValue)(nil))
