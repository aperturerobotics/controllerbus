package controller

import (
	"context"
	"sync"

	"github.com/aperturerobotics/controllerbus/directive"
)

// attachedResolver tracks a resolver and its values.
// it also manages starting and stopping the resolver.
type attachedResolver struct {
	mtx  sync.Mutex
	di   *DirectiveInstance
	vals []uint32

	res             directive.Resolver
	resolutionCtxCh chan context.Context
}

// newAttachedResolver constructs a new attachedResolver.
func newAttachedResolver(di *DirectiveInstance, res directive.Resolver) *attachedResolver {
	return &attachedResolver{
		di:              di,
		res:             res,
		resolutionCtxCh: make(chan context.Context, 1),
	}
}

// pushHandlerContext pushes the next handler context.
func (r *attachedResolver) pushHandlerContext(ctx context.Context) {
PushLoop:
	for {
		select {
		case r.resolutionCtxCh <- ctx:
			break PushLoop
		default:
		}

		select {
		case <-r.resolutionCtxCh:
		default:
		}
	}
}

// handlerCtx is canceled when the handler is removed or di canceled.
func (r *attachedResolver) execResolver(handlerCtx context.Context) error {
	errCh := make(chan error, 1)
	rctx := <-r.resolutionCtxCh

	// the running resolver count is incremented before execResolver is called.
	// rctx is the context for all currently running resolvers, may be canceled
	// resolutionCtxCh will emit the next rctx context when the previous is canceled
	// handlerCtx is the context for this specific handler and may be canceled separately from rctx

	// Resolve() needs to exit if either rctx or handlerCtx is canceled.
	// Create a new sub-context and cancel it when execOnce returns.

	for {
		nctx, nctxCancel := context.WithCancel(handlerCtx)
		go func(ctx context.Context) {
			// if skipCtx then rctx was already canceled
			var skipCtx bool
			select {
			case <-ctx.Done():
				skipCtx = true
			default:
			}
			if !skipCtx {
				errCh <- r.res.Resolve(ctx, r)
			} else {
				errCh <- nil
			}
			r.di.decrementRunningResolvers()
		}(nctx)

		var gerr bool
		select {
		case err := <-errCh:
			nctxCancel()
			gerr = true
			if err != nil && err != context.Canceled {
				return err
			}
		case <-handlerCtx.Done():
			nctxCancel()
			return handlerCtx.Err()
		case <-rctx.Done():
			nctxCancel()
		}

		// ensure we wait for resolve to return
		if !gerr {
			select {
			case _ = <-errCh:
			case <-handlerCtx.Done():
				return handlerCtx.Err()
			}
		}

		// wait for signal to restart resolver
		select {
		case <-handlerCtx.Done():
			return handlerCtx.Err()
		case rctx = <-r.resolutionCtxCh:
		}

		r.di.incrementRunningResolvers()
	}
}

// Constructed: when resolver is returned from a handler.
//
// When DI destroyed:

// AddValue adds a value to the result, returning success and an ID. If
// AddValue returns false, value was rejected. A rejected value should be
// released immediately. If the value limit is reached, the value may not be
// accepted. The value may be accepted, immediately before the resolver is
// canceled (limit reached). It is always safe to call RemoveValue with the
// ID at any time, even if the resolver is cancelled.
func (r *attachedResolver) AddValue(val directive.Value) (uint32, bool) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	di := r.di
	if di == nil {
		return 0, false
	}

	id, accepted := di.emitValue(val)
	if !accepted {
		return 0, accepted
	}

	r.vals = append(r.vals, id)
	return id, accepted
}

// RemoveValue removes a value from the result, returning found.
// It is safe to call this function even if the resolver is canceled.
func (r *attachedResolver) RemoveValue(id uint32) (val directive.Value, found bool) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	di := r.di
	if di == nil {
		return nil, false
	}

	for i, valID := range r.vals {
		if valID == id {
			found = true
			r.vals[i] = r.vals[len(r.vals)-1]
			r.vals[len(r.vals)-1] = 0
			r.vals = r.vals[:len(r.vals)-1]
			break
		}
	}

	if !found {
		return
	}

	return di.purgeEmittedValue(id)
}

// _ is a type assertion.
var _ directive.ResolverHandler = ((*attachedResolver)(nil))
