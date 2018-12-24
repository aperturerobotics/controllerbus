package configset

import (
	"context"

	"github.com/aperturerobotics/controllerbus/controller"
)

// Controller is a configset controller.
type Controller interface {
	// Controller indicates this is a controllerbus controller.
	controller.Controller

	// PushControllerConfig pushes a controller configuration for a given key, if
	// the version is newer. Returns a reference to the configset, or an error.
	PushControllerConfig(
		ctx context.Context,
		key string,
		conf *ControllerConfig,
	) (Reference, error)
}

// Reference is a reference to a pushed controller config. The reference is used
// to monitor the state of the controller and release the reference when no
// longer needed, so that the configured controllers can be properly terminated
// when no longer needed.
type Reference interface {
	// GetConfigKey returns the configset key for this controller.
	GetConfigKey() string
	// GetState returns the current state object.
	GetState() State
	// AddStateCb adds a callback that is called when the state changes.
	// Should not block.
	// Will be called with the initial state.
	AddStateCb(func(State))
	// AddChangeCb adds a callback that is called when controllerconfig, error,
	// Release releases the reference.
	Release()
}

// State contains controller state.
type State interface {
	// GetControllerConfig returns the current controller config in use.
	GetControllerConfig() *ControllerConfig
	// GetError returns any error processing the controller config.
	GetError() error
}
