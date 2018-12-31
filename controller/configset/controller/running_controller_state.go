package configset_controller

import (
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/configset"
)

// runningControllerState implements configset state
type runningControllerState struct {
	err  error
	conf configset.ControllerConfig
	ctrl controller.Controller
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

// _ is a type assertion
var _ configset.State = ((*runningControllerState)(nil))
