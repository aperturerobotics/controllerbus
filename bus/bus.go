package bus

import (
	"context"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/broadcast"
)

// ControllerCloseError reports a Controller.Close failure.
//
// AddController reports this error through its callback. ExecuteController and
// an AddController attachment failure may include it in their returned error.
// Callers can distinguish close failures from execution failures with
// errors.As.
type ControllerCloseError struct {
	Err error
}

// Error returns the controller close failure.
func (e *ControllerCloseError) Error() string {
	return "controller close failed: " + e.Err.Error()
}

// Unwrap returns the underlying close failure.
func (e *ControllerCloseError) Unwrap() error {
	return e.Err
}

// Bus manages running controllers. It has an attached directive controller,
// which is used to build declarative state requests between controllers.
type Bus interface {
	// Controller is the directive controller.
	directive.Controller

	// GetControllers returns a list of all currently active controllers.
	GetControllers() []controller.Controller
	// GetControllersBroadcast returns the broadcast that is signaled when
	// controllers are added or removed from the bus.
	GetControllersBroadcast() *broadcast.Broadcast

	// AddController attaches a controller and calls Execute asynchronously.
	// The controller receives directive callbacks until it is released,
	// Execute returns an error, or Execute panics.
	//
	// The returned release function cancels the Execute context, detaches
	// directive handling, waits for Execute to return, calls Close once, calls
	// cb, and then returns. Repeated release calls are idempotent.
	//
	// cb receives the execution error, a ControllerCloseError, both errors
	// joined together, or nil when release completes without either error.
	// cb can be nil.
	AddController(ctx context.Context, ctrl controller.Controller, cb func(exitErr error)) (func(), error)

	// ExecuteController attaches a controller and calls Execute synchronously.
	// A nil Execute return leaves the controller attached until
	// RemoveController is called. An Execute error or panic detaches the
	// controller, calls Close once, and is returned, joined with any
	// ControllerCloseError.
	// RemoveController may be called concurrently to cancel and finalize the
	// synchronous execution.
	ExecuteController(context.Context, controller.Controller) error

	// RemoveController synchronously releases one attached instance of the
	// controller. It cancels that instance's Execute context, detaches
	// directive handling, waits for Execute to return, and calls Close once.
	RemoveController(controller.Controller)
}
