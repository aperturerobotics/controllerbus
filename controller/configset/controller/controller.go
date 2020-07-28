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
	mtx                  sync.Mutex
	controllers          map[string]*runningController
	persistentRefCounter uint32
	persistentRefs       map[uint32]*runningControllerRef
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
			cv, ok := c.controllers[k]
			if ok {
				cv.mtx.Lock()
				if len(cv.refs) == 0 { // garbage collect
					delete(c.controllers, k)
					ok = false
				}
				cv.mtx.Unlock()
			}
			if !ok {
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
		for _, pref := range c.persistentRefs {
			if pref.id == key {
				existing.ApplyReference(pref)
			}
		}
	}

	ref := newRunningControllerRef(key, existing)
	existing.AddReference(ref)
	return ref, nil
}

// AddConfigSetReference adds a persistent reference to a config set which will
// be re-applied across iterations. This reference type will wait until a
// ApplyConfigSet specifies the configuration for the controller, and will add
// to the reference count for the controller such that the controller will
// continue to execute after the ApplyConfigSet call exits.
func (c *Controller) AddConfigSetReference(
	ctx context.Context,
	key string,
) (configset.Reference, error) {
	c.mtx.Lock()
	id := c.persistentRefCounter
	c.persistentRefCounter++
	rcRef := newRunningControllerRef(key, nil)
	rcRef.relInternalCallback = func() {
		c.mtx.Lock()
		delete(c.persistentRefs, id)
		c.mtx.Unlock()
	}
	c.persistentRefs[id] = rcRef // rc may be nil
	rc := c.controllers[key]
	if rc != nil {
		rc.ApplyReference(rcRef)
	}
	c.mtx.Unlock()
	return rcRef, nil
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
