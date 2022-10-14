package configset_controller

import (
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/configset"
)

// runningControllerState implements configset state
type runningControllerState struct {
	id   string
	err  error
	conf configset.ControllerConfig
	ctrl controller.Controller
}

// GetId returns the controller id.
func (s *runningControllerState) GetId() string {
	return s.id
}

// GetControllerConfig returns the current controller config in use.
func (s *runningControllerState) GetControllerConfig() configset.ControllerConfig {
	return s.conf
}

// GetController returns the controller instance if running.
// Returns nil otherwise.
func (s *runningControllerState) GetController() controller.Controller {
	return s.ctrl
}

// GetError returns any error processing the controller config.
func (s *runningControllerState) GetError() error {
	return s.err
}

// Equals checks if the two states are equal.
func (s *runningControllerState) Equals(other *runningControllerState) bool {
	switch {
	case s.id != other.id:
	case s.err != other.err:
	case s.conf != other.conf:
	case s.ctrl != other.ctrl:
	default:
		return true
	}
	return false
}

// _ is a type assertion
var _ configset.State = ((*runningControllerState)(nil))
