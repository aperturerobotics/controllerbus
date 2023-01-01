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
	// exitedCh is closed when the routine running with ctx exits
	// may be nil if ctx == nil
	exitedCh <-chan struct{}
	// vals are the attached values
	// sorted by id
	vals []*value
	// idle indicates the resolver is currently idle
	idle bool
	// exited indicates the resolver has exited
	exited bool
}

// newResolver constructs a new resolver.
func newResolver(di *directiveInstance, hnd *handler, res directive.Resolver) *resolver {
	return &resolver{
		di:   di,
		hnd:  hnd,
		res:  res,
		idle: true,
	}
}

// updateContextLocked updates the resolver context while di.c.mtx is locked
//
// if ctx is nil, stops the resolver.
func (r *resolver) updateContextLocked(ctx *context.Context) {
	if ctx != nil && r.ctx != nil && !r.exited {
		select {
		case <-r.ctx.Done():
		default:
			// resolver is still running
			return
		}
	}
	if r.ctxCancel != nil {
		r.ctxCancel()
	}
	if ctx == nil {
		if !r.idle {
			r.markIdleLocked()
		}
		r.idle, r.exited = true, true
		r.ctx, r.ctxCancel = nil, nil
	} else {
		// start resolver with new context
		exitedCh := make(chan struct{})
		waitCh := r.exitedCh
		r.exitedCh = exitedCh
		r.err = nil
		r.idle, r.exited = false, false
		r.ctx, r.ctxCancel = context.WithCancel(*ctx)
		hnd := &resolverHandler{r: r, ctx: r.ctx}
		go hnd.executeResolver(r.ctx, exitedCh, waitCh)
	}
}

// markIdleLocked marks the resolver as idle while di.c.mtx is locked
func (r *resolver) markIdleLocked() {
	if !r.idle {
		r.idle = true
		if r.di.countRunningResolversLocked() == 0 {
			r.di.handleIdleLocked()
		}
	}
}
