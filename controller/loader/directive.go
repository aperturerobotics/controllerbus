package loader

import (
	"time"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/cenkalti/backoff/v4"
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
	// GetExecControllerRetryBackoff returns the backoff to use for retries.
	// If empty / nil, uses the default.
	GetExecControllerRetryBackoff() func() backoff.BackOff
}

// ExecControllerValue is the value emitted to satisfy the ExecController
// directive.
type ExecControllerValue interface {
	// GetUpdatedTimestamp returns the last time this info changed.
	GetUpdatedTimestamp() time.Time
	// GetNextRetryTimestamp returns the next time this controller will be attempted.
	GetNextRetryTimestamp() time.Time
	// GetController returns the current controller object.
	GetController() controller.Controller
	// GetError returns the error running the controller.
	// Controller may still be set in this case.
	GetError() error
}

// _ is a type assertion
var _ directive.Value = ((ExecControllerValue)(nil))
