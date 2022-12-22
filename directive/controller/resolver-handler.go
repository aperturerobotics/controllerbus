package controller

import (
	"context"

	"github.com/aperturerobotics/controllerbus/directive"
)

// resolverHandler handles resolver values.
type resolverHandler struct {
	// r is the resolver
	r *resolver
	// ctx contains the context for this handler
	ctx context.Context
}

// AddValue adds a value to the result, returning success and an ID. If
// AddValue returns false, value was rejected. A rejected value should be
// released immediately. If the value limit is reached, the value may not be
// accepted. The value may be accepted, immediately before the resolver is
// canceled (limit reached). It is always safe to call RemoveValue with the
// ID at any time, even if the resolver is cancelled.
func (r *resolverHandler) AddValue(val directive.Value) (uint32, bool) {
	r.r.di.c.mtx.Lock()
	defer r.r.di.c.mtx.Unlock()
	if r.r.ctx != r.ctx {
		return 0, false
	}
	return r.r.di.addValueLocked(r.r, val)
}

// RemoveValue removes a value from the result, returning found.
// It is safe to call this function even if the resolver is canceled.
func (r *resolverHandler) RemoveValue(id uint32) (val directive.Value, found bool) {
	r.r.di.c.mtx.Lock()
	defer r.r.di.c.mtx.Unlock()
	if r.r.ctx != r.ctx {
		return nil, false
	}
	return r.r.di.removeValueLocked(r.r, id)
}

// MarkIdle marks the resolver as idle.
// If the resolver returns nil or an error, it's also marked as idle.
func (r *resolverHandler) MarkIdle() {
	r.r.di.c.mtx.Lock()
	defer r.r.di.c.mtx.Unlock()
	if r.r.ctx != r.ctx {
		return
	}
	r.r.markIdleLocked()
}

// ClearValues removes any values that were set by this ResolverHandler.
// Returns list of value IDs that were removed.
func (r *resolverHandler) ClearValues() []uint32 {
	r.r.di.c.mtx.Lock()
	defer r.r.di.c.mtx.Unlock()
	if r.r.ctx != r.ctx {
		return nil
	}
	removed := make([]uint32, len(r.r.vals))
	for i := len(r.r.vals) - 1; i >= 0; i-- {
		val := r.r.vals[i]
		removed[i] = val.id
		r.r.vals = r.r.vals[:i]
		r.r.di.onValueRemovedLocked(r.r, val)
	}
	return removed
}

// executeResolver is the goroutine to execute the resolver.
func (r *resolverHandler) executeResolver() {
	err := r.r.res.Resolve(r.ctx, r)

	r.r.di.c.mtx.Lock()
	defer r.r.di.c.mtx.Unlock()
	if r.r.ctx != r.ctx {
		return
	}
	r.r.err = err
	r.r.markIdleLocked()
}

// _ is a type assertion.
var _ directive.ResolverHandler = ((*resolverHandler)(nil))
