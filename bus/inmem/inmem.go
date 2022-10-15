package inmem

import (
	"context"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/pkg/errors"
)

// Bus is an in-memory controller bus.
type Bus struct {
	// Controller is the directive controller.
	directive.Controller

	// controllersMtx guards controllers
	controllersMtx sync.Mutex
	// controllers is the controllers set
	controllers []controller.Controller
}

// NewBus constructs a new in-memory Bus with a directive controller.
func NewBus(dc directive.Controller) *Bus {
	return &Bus{Controller: dc}
}

// GetControllers returns a list of all currently active controllers.
func (b *Bus) GetControllers() []controller.Controller {
	b.controllersMtx.Lock()
	c := make([]controller.Controller, len(b.controllers))
	copy(c, b.controllers)
	b.controllersMtx.Unlock()
	return c
}

// AddController adds a controller to the bus and calls Execute().
// Returns a release function for the controller instance.
// Any fatal error in the controller is written to the callback.
// The controller will receive directive callbacks until removed.
// cb can be nil
func (b *Bus) AddController(ctx context.Context, ctrl controller.Controller, cb func(exitErr error)) (func(), error) {
	b.addController(ctrl)

	relCh := make(chan struct{})
	var released atomic.Bool
	relFunc := func() {
		if !released.Swap(true) {
			close(relCh)
		}
	}

	errCh := make(chan error, 1)
	go func() {
		var err error
		defer func() {
			b.handleControllerPanic(&err)
			if err != nil {
				errCh <- err
			}
		}()
		err = ctrl.Execute(ctx)
	}()
	go func() {
		select {
		case <-ctx.Done():
			if cb != nil {
				cb(context.Canceled)
			}
		case err := <-errCh:
			if cb != nil {
				cb(err)
			}
		case <-relCh:
			if cb != nil {
				cb(nil)
			}
		}
	}()

	return relFunc, nil
}

// handleControllerPanic handles recover for a paniced controller.
func (b *Bus) handleControllerPanic(outErr *error) {
	if rerr := recover(); rerr != nil {
		e, eOk := rerr.(error)
		if eOk {
			debug.PrintStack()
			if outErr != nil {
				*outErr = errors.Wrap(e, "controller paniced")
			}
		} else if outErr != nil && *outErr == nil {
			*outErr = errors.New("controller paniced")
		}
	}
}

// ExecuteController adds a controller to the bus and calls Execute().
// Any fatal error in the controller is returned.
// The controller will receive directive callbacks.
// If the controller returns nil, call RemoveController to remove the controller.
func (b *Bus) ExecuteController(ctx context.Context, c controller.Controller) (err error) {
	b.addController(c)

	defer func() {
		b.handleControllerPanic(&err)
		if err != nil {
			b.removeController(c)
		}
	}()

	return c.Execute(ctx)
}

// RemoveController removes the controller from the bus.
func (b *Bus) RemoveController(c controller.Controller) {
	b.removeController(c)
}

// addController adds a controller to the bus
func (b *Bus) addController(c controller.Controller) {
	b.controllersMtx.Lock()
	b.controllers = append(b.controllers, c)
	b.controllersMtx.Unlock()

	_ = b.Controller.AddHandler(c)
}

// removeController removes a controller from the bus
func (b *Bus) removeController(c controller.Controller) {
	b.controllersMtx.Lock()
	for i, ci := range b.controllers {
		if ci == c {
			b.controllers[i] = b.controllers[len(b.controllers)-1]
			b.controllers[len(b.controllers)-1] = nil
			b.controllers = b.controllers[:len(b.controllers)-1]
			defer b.Controller.RemoveHandler(ci)
			break
		}
	}
	b.controllersMtx.Unlock()
}

// _ is a type assertion
var _ bus.Bus = ((*Bus)(nil))
