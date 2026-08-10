package inmem

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/broadcast"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Bus is an in-memory controller bus.
type Bus struct {
	// Controller is the directive controller.
	directive.Controller

	// le reports controller lifecycle diagnostics.
	le *logrus.Entry
	// bcast is signaled when controllers are added or removed.
	bcast broadcast.Broadcast
	// mtx guards below fields
	mtx sync.Mutex
	// controllers is the set of attached controllers
	controllers []*attachedCtrl
}

// NewBus constructs a new in-memory Bus with a directive controller.
func NewBus(dc directive.Controller) *Bus {
	return NewBusWithLogger(dc, nil)
}

// NewBusWithLogger constructs a new in-memory Bus with a directive controller
// and lifecycle logger.
func NewBusWithLogger(dc directive.Controller, le *logrus.Entry) *Bus {
	if le == nil {
		le = logrus.NewEntry(logrus.New())
	}
	return &Bus{Controller: dc, le: le}
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

// GetControllersBroadcast returns the broadcast that is signaled when
// controllers are added or removed from the bus.
func (b *Bus) GetControllersBroadcast() *broadcast.Broadcast {
	return &b.bcast
}

// AddController attaches a controller and calls Execute asynchronously.
// Its release function cancels, detaches, waits for Execute, closes once, and
// reports the lifecycle result through cb before returning.
func (b *Bus) AddController(ctx context.Context, ctrl controller.Controller, cb func(exitErr error)) (func(), error) {
	subCtx, subCtxCancel := context.WithCancel(ctx)
	attached, err := b.attachController(ctrl, subCtxCancel, cb)
	if err != nil {
		subCtxCancel()
		return nil, joinControllerErrors(err, ctrl.Close())
	}

	go b.executeAttached(subCtx, attached)
	return func() {
		attached.finalize(b)
	}, nil
}

// executeAttached runs Execute and finalizes terminal failures.
func (b *Bus) executeAttached(ctx context.Context, attached *attachedCtrl) {
	err := b.executeController(ctx, attached.ctrl)
	attached.finishExecution(err)
	if err != nil {
		attached.finalize(b)
	}
}

// executeController calls Execute and converts a panic into an error.
func (b *Bus) executeController(ctx context.Context, ctrl controller.Controller) (err error) {
	defer b.handleControllerPanic(&err)
	return ctrl.Execute(ctx)
}

// handleControllerPanic handles recovery for a controller panic.
func (b *Bus) handleControllerPanic(outErr *error) {
	if rerr := recover(); rerr != nil {
		debug.PrintStack()
		e, eOk := rerr.(error)
		if eOk {
			if outErr != nil {
				*outErr = errors.Wrap(e, "controller panicked")
			}
		} else if outErr != nil && *outErr == nil {
			*outErr = errors.New("controller panicked")
		}
	}
}

// ExecuteController attaches a controller and calls Execute synchronously.
// A nil return leaves the controller attached for RemoveController. A terminal
// failure finalizes the attachment before returning.
func (b *Bus) ExecuteController(ctx context.Context, c controller.Controller) error {
	subCtx, subCtxCancel := context.WithCancel(ctx)
	attached, err := b.attachController(c, subCtxCancel, nil)
	if err != nil {
		subCtxCancel()
		return joinControllerErrors(err, c.Close())
	}

	err = b.executeController(subCtx, c)
	attached.finishExecution(err)
	if err != nil {
		return attached.finalize(b)
	}
	return nil
}

// RemoveController synchronously finalizes one attached controller instance.
func (b *Bus) RemoveController(c controller.Controller) {
	attached := b.findController(c)
	if attached != nil {
		attached.finalize(b)
	}
}

// attachController registers and records one attached controller instance.
func (b *Bus) attachController(
	c controller.Controller,
	cancel context.CancelFunc,
	cb func(error),
) (*attachedCtrl, error) {
	// AddHandler may call HandleDirective, so it must run outside b.mtx.
	rel, err := b.AddHandler(c)
	if err != nil {
		return nil, err
	}
	attached := newAttachedCtrl(c, rel, cancel, cb)

	b.mtx.Lock()
	b.controllers = append(b.controllers, attached)
	b.mtx.Unlock()
	b.bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
		broadcast()
	})
	return attached, nil
}

// findController returns one attached instance of c.
func (b *Bus) findController(c controller.Controller) *attachedCtrl {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	for _, attached := range b.controllers {
		if attached.ctrl == c {
			return attached
		}
	}
	return nil
}

// detachController removes an exact attached instance and then releases its
// directive handler without holding b.mtx.
func (b *Bus) detachController(attached *attachedCtrl) {
	var removed bool
	b.mtx.Lock()
	for i, candidate := range b.controllers {
		if candidate == attached {
			b.controllers[i] = b.controllers[len(b.controllers)-1]
			b.controllers[len(b.controllers)-1] = nil
			b.controllers = b.controllers[:len(b.controllers)-1]
			removed = true
			break
		}
	}
	b.mtx.Unlock()

	if !removed {
		return
	}
	attached.rel()
	b.bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
		broadcast()
	})
}

// _ is a type assertion
var _ bus.Bus = ((*Bus)(nil))
