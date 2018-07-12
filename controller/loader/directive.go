package loader

import (
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// LoadController is a directive indicating a controller instance should be
// constructed given a factory and a configuration.
type LoadController interface {
	// Directive indicates this is a directive.
	directive.Directive

	// GetLoadControllerFactory returns the factory desired to load.
	GetLoadControllerFactory() controller.Factory

	// GetLoadControllerConfig returns the config to load the controller with.
	GetLoadControllerConfig() config.Config
}

// LoadControllerValue is the value emitted to satisfy the LoadController
// directive.
type LoadControllerValue = controller.Controller

// _ is a type assertion
var _ directive.Value = ((LoadControllerValue)(nil))
