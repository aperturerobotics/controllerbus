package configset_controller

import (
	"context"
	"sync"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
)

// Version is the version of the controller implementation.
var Version = semver.MustParse("0.0.1")

// ControllerID is the ID of the controller.
const ControllerID = "controllerbus/configset/1"

// Controller is the ConfigSet controller.
type Controller struct {
	// le is the root logger
	le *logrus.Entry
	// bus is the controller bus
	bus bus.Bus
	// wakeCh wakes the controller
	wakeCh chan struct{}

	// mtx guards the controllers map
	mtx         sync.Mutex
	controllers map[string]*runningController
}

// NewController constructs a new peer controller.
// If privKey is nil, one will be generated.
func NewController(le *logrus.Entry, bus bus.Bus) (*Controller, error) {
	return &Controller{
		le:          le,
		bus:         bus,
		wakeCh:      make(chan struct{}, 1),
		controllers: make(map[string]*runningController),
	}, nil
}

// Execute executes the given controller.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *Controller) Execute(ctx context.Context) error {
	execControllers := make(map[string]context.CancelFunc)
ExecLoop:
	for {
		select {
		case <-ctx.Done():
			break ExecLoop
		case <-c.wakeCh:
		}

		c.mtx.Lock()
		for k, cancel := range execControllers {
			if _, ok := c.controllers[k]; !ok {
				cancel()
				delete(execControllers, k)
			}
		}
		for k, c := range c.controllers {
			if _, ok := execControllers[k]; !ok {
				nctx, nctxCancel := context.WithCancel(ctx)
				execControllers[k] = nctxCancel
				go c.Execute(nctx)
			}
		}
		c.mtx.Unlock()
	}

	c.mtx.Lock()
	for k := range c.controllers {
		delete(c.controllers, k)
		if cancel, kiOk := execControllers[k]; kiOk {
			cancel()
			delete(execControllers, k)
		}
	}
	c.mtx.Unlock()
	return nil
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (c *Controller) HandleDirective(
	ctx context.Context,
	di directive.Instance,
) (directive.Resolver, error) {
	dir := di.GetDirective()
	switch d := dir.(type) {
	case configset.ApplyConfigSet:
		return c.resolveApplyConfigSet(ctx, di, d), nil
	}

	return nil, nil
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() controller.Info {
	return controller.NewInfo(
		ControllerID,
		Version,
		"configset actuation controller",
	)
}

// PushControllerConfig pushes a controller configuration for a given key, if
// the version is newer. Returns a reference to the configset, or an error.
func (c *Controller) PushControllerConfig(
	ctx context.Context,
	key string,
	conf configset.ControllerConfig,
) (configset.Reference, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	existing, existingOk := c.controllers[key]
	if existingOk {
		if existing.ApplyConfig(conf) {
			c.wake()
		}
	} else {
		existing = newRunningController(c, key, conf)
		c.controllers[key] = existing
		c.wake()
	}

	ref := newRunningControllerRef(existing)
	existing.AddReference(ref)
	return ref, nil
}

// resolveApplyConfigSet resolves the ApplyConfigSet directive
func (c *Controller) resolveApplyConfigSet(
	ctx context.Context,
	di directive.Instance,
	dir configset.ApplyConfigSet,
) directive.Resolver {
	return newApplyConfigSetResolver(c, ctx, di, dir)
}

// wake wakes the controller
func (c *Controller) wake() {
	select {
	case c.wakeCh <- struct{}{}:
	default:
	}
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (c *Controller) Close() error {
	return nil
}

// _ is a type assertion
var _ configset.Controller = ((*Controller)(nil))
