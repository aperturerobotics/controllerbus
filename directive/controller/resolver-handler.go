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
//
// Value IDs start at 1.
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

// MarkIdle marks the resolver as idle or not idle.
// If the resolver returns nil or an error, it's also marked as idle.
func (r *resolverHandler) MarkIdle(idle bool) {
	r.r.di.c.mtx.Lock()
	defer r.r.di.c.mtx.Unlock()
	if r.r.ctx != r.ctx {
		return
	}
	r.r.setIdleLocked(idle)
}

// CountValues returns the number of values that were set.
// if allResolvers=false, returns the number set by this ResolverHandler.
// if allResolvers=true, returns the number set by all resolvers.
func (r *resolverHandler) CountValues(allResolvers bool) int {
	r.r.di.c.mtx.Lock()
	defer r.r.di.c.mtx.Unlock()
	if allResolvers {
		return r.r.di.countValuesLocked()
	} else {
		return len(r.r.vals)
	}
}

// ClearValues removes any values that were set by this ResolverHandler.
// Returns list of value IDs that were removed.
func (r *resolverHandler) ClearValues() []uint32 {
	r.r.di.c.mtx.Lock()
	defer r.r.di.c.mtx.Unlock()
	if r.r.ctx != r.ctx {
		return nil
	}
	vals := r.r.vals
	removed := make([]uint32, len(vals))
	for i := range vals {
		removed[i] = vals[i].id
	}
	r.r.vals = nil
	r.r.di.onValuesRemovedLocked(vals...)
	return removed
}

// AddValueRemovedCallback adds a callback that will be called when the given
// value id is disposed or removed. The callback might be called immediately if
// the value was already removed.
func (r *resolverHandler) AddValueRemovedCallback(id uint32, cb func()) func() {
	emptyFn := func() {}
	if cb == nil {
		return emptyFn
	}

	r.r.di.c.mtx.Lock()
	var relFn func()
	var found bool
	if r.r.ctx == r.ctx {
		_, relFn, found = r.r.di.addValueRemovedCallbackLocked(r.r, id, cb)
	}
	r.r.di.c.mtx.Unlock()
	if !found {
		cb()
		return emptyFn
	}
	return relFn
}

// AddResolverRemovedCallback adds a callback that will be called when the
// directive resolver is removed.
func (r *resolverHandler) AddResolverRemovedCallback(cb func()) func() {
	emptyFn := func() {}
	if cb == nil {
		return emptyFn
	}

	r.r.di.c.mtx.Lock()
	var relFn func()
	var found bool
	if r.r.ctx == r.ctx {
		relFn, found = r.r.di.addResolverRemovedCallbackLocked(r.r, cb)
	}
	r.r.di.c.mtx.Unlock()
	if !found {
		cb()
		return emptyFn
	}
	return relFn
}

// AddResolver adds a resolver as a child of the current resolver.
//
// The child resolver will be removed if the parent handler is removed.
//
// The callback will be called if the child resolver is removed for any
// reason, including if the parent resolver, handler, or directive are
// removed.
//
// The callback might be called immediately if the child resolver was
// already removed or not created.
//
// Returns a release function to clear and stop the resolver early.
// Does nothing if res == nil returning an empty release func.
func (r *resolverHandler) AddResolver(res directive.Resolver, removedCb func()) func() {
	emptyFn := func() {}
	if res == nil {
		if removedCb != nil {
			removedCb()
		}
		return emptyFn
	}

	r.r.di.c.mtx.Lock()
	// only start the sub-resolver if this one has not been canceled.
	var relFn func()
	var found bool
	if r.r.ctx == r.ctx {
		relFn, found = r.r.di.attachStartSubResolverLocked(r.r, res, removedCb)
	}
	r.r.di.c.mtx.Unlock()
	if !found {
		if removedCb != nil {
			removedCb()
		}
		return emptyFn
	}
	return relFn
}

// executeResolver is the goroutine to execute the resolver.
func (r *resolverHandler) executeResolver(ctx context.Context, exitedCh chan<- struct{}, waitCh <-chan struct{}) {
	defer close(exitedCh)
	if waitCh != nil {
		select {
		case <-ctx.Done():
			return
		case <-waitCh:
		}
	}

	err := r.r.res.Resolve(r.ctx, r)

	r.r.di.c.mtx.Lock()
	defer r.r.di.c.mtx.Unlock()
	if r.r.ctx != r.ctx {
		return
	}

	defer r.r.di.deferCheckStateChanged()()
	r.r.exited = true
	r.r.setErrLocked(err)
	r.r.setIdleLocked(true)
}

// _ is a type assertion.
var _ directive.ResolverHandler = ((*resolverHandler)(nil))
