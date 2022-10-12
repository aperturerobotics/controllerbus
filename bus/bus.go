package bus

import (
	"context"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// Bus manages running controllers. It has an attached directive controller,
// which is used to build declarative state requests between controllers.
type Bus interface {
	// Controller is the directive controller.
	directive.Controller

	// GetControllers returns a list of all currently active controllers.
	GetControllers() []controller.Controller

	// AddController adds a controller to the bus and calls Execute().
	// The controller will exit if ctx is canceled.
	// Returns a release function for the controller reference.
	// The controller will receive directive callbacks until removed.
	// Any fatal error in the controller is written to cb.
	// If the controller is released, cb will be called with nil.
	// cb can be nil
	AddController(ctx context.Context, ctrl controller.Controller, cb func(exitErr error)) (func(), error)

	// ExecuteController adds a controller to the bus and calls Execute().
	// The controller will exit if ctx is canceled.
	// Any fatal error in the controller is returned.
	// The controller will receive directive callbacks.
	// If this function returns nil, call RemoveController to remove the controller.
	ExecuteController(context.Context, controller.Controller) error

	// RemoveController removes the controller from the bus.
	// The controller will no longer receive callbacks.
	// Note: this might not cancel the Execute() context automatically.
	RemoveController(controller.Controller)
}
