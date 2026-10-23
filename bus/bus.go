package bus

import (
	"context"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/broadcast"
	"github.com/pkg/errors"
)

// ErrClosed rejects controller admission after bus shutdown begins.
var ErrClosed = errors.New("controller bus is closed")

// ControllerCloseError reports a Controller.Close failure.
//
// AddController reports this error through its callback.
// ExecuteController may include it in its returned error.
// An AddController attachment failure may include it in its returned error.
// Use errors.As to tell a close failure from an execution failure.
type ControllerCloseError struct {
	// Err is the controller's cleanup failure.
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

// Bus manages running controllers.
// Its directive controller carries declarative state requests between controllers.
type Bus interface {
	// Controller is the directive controller.
	directive.Controller

	// GetControllers returns a list of all currently active controllers.
	GetControllers() []controller.Controller
	// GetControllersBroadcast returns the broadcast that is signaled when
	// controllers are added or removed from the bus.
	GetControllersBroadcast() *broadcast.Broadcast

	// AddController attaches a controller and calls Execute asynchronously.
	// The controller receives directive callbacks until it is released.
	// An Execute error or panic also ends directive callbacks.
	//
	// The release function cancels the Execute context.
	// It then detaches directive handling and waits for Execute to return.
	// It then calls Close once, calls cb, and returns.
	// Repeated release calls are idempotent.
	//
	// cb receives the execution error, a ControllerCloseError, or both joined.
	// cb receives nil if release completes with neither error.
	// cb can be nil.
	AddController(ctx context.Context, ctrl controller.Controller, cb func(exitErr error)) (func(), error)

	// ExecuteController attaches a controller and calls Execute synchronously.
	// A nil Execute return leaves the controller attached until RemoveController.
	// An Execute error or panic detaches the controller and calls Close once.
	// Returns the Execute error joined with any ControllerCloseError.
	// RemoveController may be called concurrently to cancel and finalize it.
	ExecuteController(context.Context, controller.Controller) error

	// RemoveController synchronously releases one attached instance of the controller.
	// It cancels that instance's Execute context and detaches directive handling.
	// It then waits for Execute to return and calls Close once.
	RemoveController(controller.Controller)

	// Close rejects new controllers and cancels admitted controllers.
	// It waits for their Execute, Close, and completion callbacks.
	// It also waits for releases already in progress.
	// Returns the joined ControllerCloseError values.
	// Repeated calls wait for the same shutdown and return the same result.
	// Call Close from the owning host, never from an owned controller lifecycle.
	Close() error
}
