package controller

import (
	"context"
	"errors"
	"net/url"
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

	// mtx guards below fields
	mtx sync.Mutex

	handlers   []*attachedHandler
	directives []*DirectiveInstance
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
	if err := dir.Validate(); err != nil {
		return nil, nil, err
	}

	var dirDebugStr string
	debuggable, isDebuggable := dir.(directive.Debuggable)
	if isDebuggable {
		if debugVals := debuggable.GetDebugVals(); debugVals != nil {
			dirDebugStr = url.Values(debugVals).Encode()
			dirDebugStr, _ = url.PathUnescape(dirDebugStr)
		}
	}
	dirNameDebugStr := dir.GetName()
	if dirDebugStr != "" {
		dirNameDebugStr += "<" + dirDebugStr + ">"
	}
	le := c.le.WithField("directive", dirNameDebugStr)

	c.mtx.Lock() // lock first
	defer c.mtx.Unlock()

	for ii := 0; ii < len(c.directives); ii++ {
		di := c.directives[ii]
		d := di.GetDirective()
		if dir.IsEquivalent(d) {
			if dir.Superceeds(d) {
				di.Close()
				break
			}

			ref := di.AddReference(cb, false)
			if ref == nil {
				continue
			} else {
				return di, ref, nil
			}
		}
	}

	// Build new reference
	di, ref := NewDirectiveInstance(c.ctx, le, dir, cb, func(di *DirectiveInstance) {
		le.Debug("removed directive")
		c.mtx.Lock() // lock first
		for i, d := range c.directives {
			if d == di {
				c.directives[i] = c.directives[len(c.directives)-1]
				c.directives[len(c.directives)-1] = nil
				c.directives = c.directives[:len(c.directives)-1]
				break
			}
		}
		c.mtx.Unlock()
	})
	le.Debug("added directive")
	c.directives = append(c.directives, di)
	for _, handler := range c.handlers {
		if err := c.callHandler(handler, di); err != nil {
			le.WithError(err).Warn("unable to call directive handler")
		}
	}
	return di, ref, nil
}

// AddHandler adds a directive handler.
// The handler will receive calls for all existing directives (initial set).
func (c *DirectiveController) AddHandler(hnd directive.Handler) error {
	c.mtx.Lock() // lock first
	defer c.mtx.Unlock()

	ahnd := newAttachedHandler(c.ctx, hnd)
	c.handlers = append(c.handlers, ahnd)

	for _, dir := range c.directives {
		err := c.callHandler(ahnd, dir)
		if err != nil {
			c.le.WithError(err).Warn("handler returned error")
		}
	}

	return nil
}

// RemoveHandler removes a directive handler.
func (c *DirectiveController) RemoveHandler(hnd directive.Handler) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	for i, h := range c.handlers {
		if h.Handler == hnd {
			h.Cancel()
			c.handlers[i] = c.handlers[len(c.handlers)-1]
			c.handlers[len(c.handlers)-1] = nil
			c.handlers = c.handlers[:len(c.handlers)-1]
			break
		}
	}
	// TODO: ensure removing resolvers from handler.
}

// GetDirectives returns a list of all currently active directives.
func (c *DirectiveController) GetDirectives() []directive.Instance {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	d := make([]directive.Instance, len(c.directives))
	for i := range c.directives {
		d[i] = c.directives[i]
	}
	return d
}

// callHandler calls the directive handler with the directive, managing the resolver if returned.
func (c *DirectiveController) callHandler(ahnd *attachedHandler, inst *DirectiveInstance) error {
	// HandleDirective asks if the handler can resolve the directive.
	// If it can, it returns a resolver. If not, returns nil.
	// Any exceptional errors are returned for logging.
	// It is safe to add a reference to the directive during this call.
	hnd := ahnd.Handler
	handleCtx, handleCtxCancel := context.WithCancel(ahnd.Context)

	// go is needed here due to relMtx being locked
	go inst.AddDisposeCallback(handleCtxCancel)
	// inst.AddDisposeCallback(handleCtxCancel)

	resolver, err := hnd.HandleDirective(handleCtx, inst)
	if err != nil {
		return err
	}
	if resolver != nil {
		// attach resolver
		inst.attachResolver(handleCtx, resolver)
	}

	return nil
}

// Close closes the directive instance.
func (c *DirectiveController) Close() {
	c.ctxCancel()
}

// _ is a type assertion
var _ directive.Controller = ((*DirectiveController)(nil))
