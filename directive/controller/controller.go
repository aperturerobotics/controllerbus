package controller

import (
	"context"
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
	handlers    []directive.Handler

	// directivesMtx guards directives
	directivesMtx sync.Mutex
	directives    []*DirectiveInstance
}

// NewDirectiveController builds a new directive controller.
func NewDirectiveController(ctx context.Context, le *logrus.Entry) *DirectiveController {
	nctx, nctxCancel := context.WithCancel(ctx)
	return &DirectiveController{ctx: nctx, ctxCancel: nctxCancel, le: le}
}

// AddDirective adds a directive to the controller.
// This call de-duplicates equivalent directives.
// Returns the instance, new reference, and any error.
func (c *DirectiveController) AddDirective(
	dir directive.Directive,
	cb func(val directive.Value),
) (directive.Instance, directive.Reference, error) {
	c.directivesMtx.Lock()
	defer c.directivesMtx.Unlock()

	for _, di := range c.directives {
		d := di.GetDirective()
		if d.IsEquivalent(dir) {
			if dir.Superceeds(d) {
				di.Close()
				break
			}

			return di, di.AddReference(cb), nil
		}
	}

	// Build new reference
	var di *DirectiveInstance
	var ref directive.Reference
	di, ref = NewDirectiveInstance(dir, cb, func() {
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
	var err error
	c.directivesMtx.Lock()
	defer c.directivesMtx.Unlock()
	for _, dir := range c.directives {
		err = c.callHandler(hnd, dir)
		if err != nil {
			return err
		}
	}

	c.handlersMtx.Lock()
	c.handlers = append(c.handlers, hnd)
	c.handlersMtx.Unlock()

	return nil
}

// RemoveHandler removes a directive handler.
func (c *DirectiveController) RemoveHandler(hnd directive.Handler) {
	c.handlersMtx.Lock()
	for i, h := range c.handlers {
		if h == hnd {
			c.handlers[i] = c.handlers[len(c.handlers)-1]
			c.handlers[len(c.handlers)-1] = nil
			c.handlers = c.handlers[:len(c.handlers)-1]

			// TODO: expire all resolvers attached to handler.
			break
		}
	}
	c.handlersMtx.Unlock()
}

// callHandler calls the directive handler with the directive, managing the resolver if returned.
func (c *DirectiveController) callHandler(hnd directive.Handler, inst *DirectiveInstance) error {
	// HandleDirective asks if the handler can resolve the directive.
	// If it can, it returns a resolver. If not, returns nil.
	// Any exceptional errors are returned for logging.
	// It is safe to add a reference to the directive during this call.
	resolver, err := hnd.HandleDirective(inst)
	if err != nil {
		return err
	}

	if resolver != nil {
		go c.callResolver(resolver, inst)
	}

	return nil
}

// callResolver calls a directive resolver.
func (c *DirectiveController) callResolver(res directive.Resolver, inst *DirectiveInstance) {
	valCh := make(chan directive.Value, 10)
	errCh := make(chan error, 1)
	go func() {
		errCh <- res.Resolve(c.ctx, valCh)
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case val := <-valCh:
			c.le.Debugf("emitting value: %#v", val)
			inst.emitValue(val)
		case err := <-errCh:
			if err != nil {
				c.le.
					WithError(err).
					Warn("resolver exited with error")
			}
			return
		}
	}
}

// Close closes the directive instance.
func (c *DirectiveController) Close() {
	c.ctxCancel()
}

// _ is a type assertion
var _ directive.Controller = ((*DirectiveController)(nil))
