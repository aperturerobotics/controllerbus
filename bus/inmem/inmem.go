package inmem

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/pkg/errors"
)

// Bus is an in-memory controller bus.
type Bus struct {
	// Controller is the directive controller.
	directive.Controller

	// mtx guards below fields
	mtx sync.Mutex
	// controllers is the set of attached controllers
	controllers []*attachedCtrl
}

// NewBus constructs a new in-memory Bus with a directive controller.
func NewBus(dc directive.Controller) *Bus {
	return &Bus{Controller: dc}
}

// GetControllers returns a list of all currently active controllers.
func (b *Bus) GetControllers() []controller.Controller {
	b.mtx.Lock()
	c := make([]controller.Controller, len(b.controllers))
	for i := range b.controllers {
		c[i] = b.controllers[i].ctrl
	}
	b.mtx.Unlock()
	return c
}

// AddController adds a controller to the bus and calls Execute().
// Returns a release function for the controller instance.
// Any fatal error in the controller is written to the callback.
// The controller will receive directive callbacks until removed.
// cb can be nil
func (b *Bus) AddController(ctx context.Context, ctrl controller.Controller, cb func(exitErr error)) (func(), error) {
	subCtx, subCtxCancel := context.WithCancel(ctx)
	relFunc := func() {
		subCtxCancel()
		b.removeController(ctrl)
	}
	if err := b.addController(ctrl); err != nil {
		subCtxCancel()
		_ = ctrl.Close()
		return nil, err
	}
	go func() {
		var err error
		defer func() {
			b.handleControllerPanic(&err)
			if err != nil {
				subCtxCancel()
				b.removeController(ctrl)
				if cb != nil {
					cb(err)
				}
			}
		}()
		err = ctrl.Execute(subCtx)
	}()
	return relFunc, nil
}

// handleControllerPanic handles recover for a paniced controller.
func (b *Bus) handleControllerPanic(outErr *error) {
	if rerr := recover(); rerr != nil {
		debug.PrintStack()
		e, eOk := rerr.(error)
		if eOk {
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
	if err := b.addController(c); err != nil {
		return err
	}

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
func (b *Bus) addController(c controller.Controller) error {
	b.mtx.Lock()
	rel, err := b.AddHandler(c)
	if err == nil {
		b.controllers = append(b.controllers, &attachedCtrl{
			ctrl: c,
			rel:  rel,
		})
	}
	b.mtx.Unlock()
	return err
}

// removeController removes a controller from the bus
func (b *Bus) removeController(c controller.Controller) {
	b.mtx.Lock()
	for i, ci := range b.controllers {
		if ci.ctrl == c {
			b.controllers[i] = b.controllers[len(b.controllers)-1]
			b.controllers[len(b.controllers)-1] = nil
			b.controllers = b.controllers[:len(b.controllers)-1]
			ci.rel()
			break
		}
	}
	b.mtx.Unlock()
}

// _ is a type assertion
var _ bus.Bus = ((*Bus)(nil))
