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

	// ExecuteController adds a controller to the bus and calls Execute().
	// Any fatal error in the controller is returned.
	// The controller will receive directive callbacks until this returns.
	// If the controller execute returns nil, waits until context is canceled to return.
	ExecuteController(context.Context, controller.Controller) error
	// RemoveController removes the controller from the bus.
	// The controller will no longer receive callbacks.
	// Note: this might not cancel the Execute() context automatically.
	RemoveController(context.Context, controller.Controller) error
}
