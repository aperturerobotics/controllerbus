package controller

import (
	"context"

	"github.com/aperturerobotics/controllerbus/directive"
)

// resolver tracks a resolver and its values.
type resolver struct {
	di  *directiveInstance
	hnd *handler
	res directive.Resolver

	// di.c.mtx guards below fields

	// err is the current error returned by the resolver
	err error
	// ctx is the current resolver context
	ctx context.Context
	// ctxCancel cancels ctx
	ctxCancel context.CancelFunc
	// vals are the attached values
	// sorted by id
	vals []*value
	// idle indicates the resolver is currently idle
	idle bool
}

// newResolver constructs a new resolver.
func newResolver(di *directiveInstance, hnd *handler, res directive.Resolver) *resolver {
	return &resolver{
		di:  di,
		hnd: hnd,
		res: res,
	}
}

// updateContextLocked updates the resolver context while di.c.mtx is locked
//
// if ctx is nil, stops the resolver.
func (r *resolver) updateContextLocked(ctx *context.Context) {
	if r.ctxCancel != nil {
		r.ctxCancel()
	}
	if ctx == nil {
		r.ctx, r.ctxCancel = nil, nil
		r.idle = true
	} else {
		// start resolver with new context
		r.ctx, r.ctxCancel = context.WithCancel(*ctx)
		r.idle = false
		hnd := &resolverHandler{r: r, ctx: r.ctx}
		go hnd.executeResolver()
	}
}

// markIdleLocked marks the resolver as idle while di.c.mtx is locked
func (r *resolver) markIdleLocked() {
	if !r.idle {
		r.idle = true
		r.di.runningResolvers--
		if r.di.runningResolvers == 0 {
			r.di.handleIdleLocked()
		}
	}
}
