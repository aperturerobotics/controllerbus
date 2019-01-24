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
	if err := dir.Validate(); err != nil {
		return nil, nil, err
	}

	var dirDebugStr string
	if debugVals := dir.GetDebugVals(); debugVals != nil {
		dirDebugStr = url.Values(debugVals).Encode()
		dirDebugStr, _ = url.PathUnescape(dirDebugStr)
	}
	dirNameDebugStr := dir.GetName()
	if dirDebugStr != "" {
		dirNameDebugStr += "<" + dirDebugStr + ">"
	}
	le := c.le.WithField("directive", dirNameDebugStr)

	c.directivesMtx.Lock() // lock first
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
				continue
			} else {
				return di, ref, nil
			}
		}
	}

	// Build new reference
	di, ref := NewDirectiveInstance(c.ctx, c.le, dir, cb, func(di *DirectiveInstance) {
		le.Debug("removed directive")
		c.directivesMtx.Lock() // lock first
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
	le.Debug("added directive")
	c.directives = append(c.directives, di)
	c.handlersMtx.Lock()
	for _, handler := range c.handlers {
		if err := c.callHandler(handler, di); err != nil {
			le.WithError(err).Warn("unable to call directive handler")
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

	dirs := make([]*DirectiveInstance, len(c.directives))
	copy(dirs, c.directives)
	c.directivesMtx.Unlock()

	for _, dir := range dirs {
		_ = c.callHandler(ahnd, dir)
	}

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

// GetDirectives returns a list of all currently active directives.
func (c *DirectiveController) GetDirectives() []directive.Instance {
	c.directivesMtx.Lock()
	d := make([]directive.Instance, len(c.directives))
	for i := range c.directives {
		d[i] = c.directives[i]
	}
	c.directivesMtx.Unlock()
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
