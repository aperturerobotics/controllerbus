package controller

import (
	"context"
	"sync"

	"github.com/aperturerobotics/controllerbus/directive"
)

// attachedResolver tracks a resolver and its values.
// it also manages starting and stopping the resolver.
type attachedResolver struct {
	diMtx sync.Mutex
	di    *DirectiveInstance

	valsMtx sync.Mutex
	vals    []uint32

	res       directive.Resolver
	handlerCh chan context.Context
}

// newAttachedResolver constructs a new attachedResolver.
func newAttachedResolver(di *DirectiveInstance, res directive.Resolver) *attachedResolver {
	return &attachedResolver{
		di:        di,
		res:       res,
		handlerCh: make(chan context.Context, 1),
	}
}

// pushHandlerContext pushes the next handler context.
func (r *attachedResolver) pushHandlerContext(ctx context.Context) {
PushLoop:
	for {
		select {
		case r.handlerCh <- ctx:
			break PushLoop
		default:
		}

		select {
		case <-r.handlerCh:
		default:
		}
	}
}

// handlerCtx is canceled when the handler is removed or di canceled.
func (r *attachedResolver) execResolver(handlerCtx context.Context) error {
	errCh := make(chan error, 1)
	rctx := <-r.handlerCh

	nctx, nctxCancel := context.WithCancel(handlerCtx)
	defer nctxCancel()

	for {
		go func() {
			errCh <- r.res.Resolve(nctx, r)
		}()

		var gerr bool
		select {
		case err := <-errCh:
			gerr = true
			if err != context.Canceled {
				return err
			}
		case <-nctx.Done():
			return nctx.Err()
		case <-rctx.Done():
		}

		// ensure we wait for resolve to return
		if !gerr {
			select {
			case _ = <-errCh:
			case <-nctx.Done():
				return nctx.Err()
			}
		}

		// wait for signal to restart resolver
		select {
		case <-nctx.Done():
			return nctx.Err()
		case rctx = <-r.handlerCh:
		}
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
func (r *attachedResolver) AddValue(val directive.Value) (id uint32, accepted bool) {
	r.diMtx.Lock()
	di := r.di
	r.diMtx.Unlock()

	if di == nil {
		return 0, false
	}

	r.valsMtx.Lock()
	id, accepted = di.emitValue(val)
	if !accepted {
		return
	}

	r.vals = append(r.vals, id)
	r.valsMtx.Unlock()
	return
}

// RemoveValue removes a value from the result, returning found.
// It is safe to call this function even if the resolver is canceled.
func (r *attachedResolver) RemoveValue(id uint32) (val directive.Value, found bool) {
	r.diMtx.Lock()
	di := r.di
	r.diMtx.Unlock()

	if di == nil {
		return nil, false
	}

	r.valsMtx.Lock()
	for i, valID := range r.vals {
		if valID == id {
			found = true
			r.vals[i] = r.vals[len(r.vals)-1]
			r.vals[len(r.vals)-1] = 0
			r.vals = r.vals[:len(r.vals)-1]
			break
		}
	}
	r.valsMtx.Unlock()

	if !found {
		return
	}

	return di.purgeEmittedValue(id)
}

// _ is a type assertion.
var _ directive.ResolverHandler = ((*attachedResolver)(nil))
