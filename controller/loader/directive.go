package loader

import (
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// ExecController is a directive indicating a controller instance should be
// constructed and executed given a factory and a configuration.
// If/when the controller exits, the directive is canceled with the error.
type ExecController interface {
	// Directive indicates this is a directive.
	directive.Directive

	// GetExecControllerFactory returns the factory desired to load.
	GetExecControllerFactory() controller.Factory

	// GetExecControllerConfig returns the config to load the controller with.
	GetExecControllerConfig() config.Config
}

// ExecControllerValue is the value emitted to satisfy the ExecController
// directive.
type ExecControllerValue = controller.Controller

// _ is a type assertion
var _ directive.Value = ((ExecControllerValue)(nil))
