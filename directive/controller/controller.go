package controller

import (
	"context"
	"errors"
	"sync"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// DirectiveController is the directive controller.
type DirectiveController struct {
	// ctx is the directive context
	ctx context.Context
	// ctxCancel cancels the directive context.
	ctxCancel context.CancelFunc

	// le is the log entry
	le *logrus.Entry

	// handlersMtx guards handlers
	handlersMtx sync.Mutex
	handlers    []*attachedHandler

	// directivesMtx guards directives
	directivesMtx sync.Mutex
	directives    []*DirectiveInstance
}

// NewDirectiveController builds a new directive controller.
func NewDirectiveController(ctx context.Context, le *logrus.Entry) *DirectiveController {
	nctx, nctxCancel := context.WithCancel(ctx)
	return &DirectiveController{
		ctx:       nctx,
		ctxCancel: nctxCancel,
		le:        le,
	}
}

// AddDirective adds a directive to the controller.
// This call de-duplicates equivalent directives.
// Returns the instance, new reference, and any error.
func (c *DirectiveController) AddDirective(
	dir directive.Directive,
	cb directive.ReferenceHandler,
) (directive.Instance, directive.Reference, error) {
	if dir == nil {
		return nil, nil, errors.New("directive cannot be nil")
	}

	c.directivesMtx.Lock()
	defer c.directivesMtx.Unlock()

	for ii := 0; ii < len(c.directives); ii++ {
		di := c.directives[ii]
		d := di.GetDirective()
		if d.IsEquivalent(dir) {
			if dir.Superceeds(d) {
				di.Close()
				break
			}

			ref := di.AddReference(cb, false)
			if ref == nil {
				c.directives[ii] = c.directives[len(c.directives)-1]
				c.directives[len(c.directives)-1] = nil
				c.directives = c.directives[:len(c.directives)-1]
				ii--
				continue
			} else {
				return di, ref, nil
			}
		}
	}

	// Build new reference
	var di *DirectiveInstance
	var ref directive.Reference
	di, ref = NewDirectiveInstance(c.ctx, dir, cb, func() {
		// c.le.Debugf("removed directive: %#v", dir)
		c.directivesMtx.Lock()
		for i, d := range c.directives {
			if d == di {
				c.directives[i] = c.directives[len(c.directives)-1]
				c.directives[len(c.directives)-1] = nil
				c.directives = c.directives[:len(c.directives)-1]
				break
			}
		}
		c.directivesMtx.Unlock()
	})
	// c.le.Debugf("added directive: %#v", dir)
	c.directives = append(c.directives, di)
	c.handlersMtx.Lock()
	for _, handler := range c.handlers {
		if err := c.callHandler(handler, di); err != nil {
			c.le.WithError(err).Warn("unable to call directive handler")
		}
	}
	c.handlersMtx.Unlock()
	return di, ref, nil
}

// AddHandler adds a directive handler.
// The handler will receive calls for all existing directives (initial set).
func (c *DirectiveController) AddHandler(hnd directive.Handler) error {
	c.directivesMtx.Lock() // lock first
	c.handlersMtx.Lock()
	ahnd := newAttachedHandler(c.ctx, hnd)
	c.handlers = append(c.handlers, ahnd)
	c.handlersMtx.Unlock()

	for _, dir := range c.directives {
		_ = c.callHandler(ahnd, dir)
	}
	c.directivesMtx.Unlock()

	return nil
}

// RemoveHandler removes a directive handler.
func (c *DirectiveController) RemoveHandler(hnd directive.Handler) {
	c.handlersMtx.Lock()
	for i, h := range c.handlers {
		if h.Handler == hnd {
			h.Cancel()
			c.handlers[i] = c.handlers[len(c.handlers)-1]
			c.handlers[len(c.handlers)-1] = nil
			c.handlers = c.handlers[:len(c.handlers)-1]
			break
		}
	}
	c.handlersMtx.Unlock()
}

// callHandler calls the directive handler with the directive, managing the resolver if returned.
func (c *DirectiveController) callHandler(ahnd *attachedHandler, inst *DirectiveInstance) error {
	// HandleDirective asks if the handler can resolve the directive.
	// If it can, it returns a resolver. If not, returns nil.
	// Any exceptional errors are returned for logging.
	// It is safe to add a reference to the directive during this call.
	hnd := ahnd.Handler
	resolver, err := hnd.HandleDirective(ahnd.Context, inst)
	if err != nil {
		return err
	}

	if resolver != nil {
		// attach resolver
		inst.attachResolver(ahnd.Context, resolver)
	}

	return nil
}

// Close closes the directive instance.
func (c *DirectiveController) Close() {
	c.ctxCancel()
}

// _ is a type assertion
var _ directive.Controller = ((*DirectiveController)(nil))
