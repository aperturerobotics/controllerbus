package refcount

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/aperturerobotics/controllerbus/util/ccontainer"
)

// RefCount is a refcount driven object container.
// Wraps a ccontainer with a ref count mechanism.
// When there are no references, the container contents are released.
type RefCount[T comparable] struct {
	// ctx is the root context
	ctx context.Context
	// target is the target ccontainer
	target *ccontainer.CContainer[T]
	// targetErr is the destination for resolution errors
	targetErr *ccontainer.CContainer[*error]
	// resolver is the resolver function
	// returns the value and a release function
	resolver func(ctx context.Context) (T, func(), error)
	// mtx guards below fields
	mtx sync.Mutex
	// refs is the list of references.
	refs map[*Ref[T]]struct{}
	// resolveCtx is the resolution context.
	resolveCtx context.Context
	// resolveCtxCancel cancels resolveCtx
	resolveCtxCancel context.CancelFunc
	// valueRel releases the current value.
	valueRel func()
}

// Ref is a reference to a RefCount.
type Ref[T comparable] struct {
	rc  *RefCount[T]
	rel atomic.Bool
	cb  func(val T, err error)
}

// Release releases the reference.
func (k *Ref[T]) Release() {
	if k.rel.Swap(true) {
		return
	}
	k.rc.removeRef(k)
}

// NewRefCount builds a new RefCount.
// target and targetErr can be empty
func NewRefCount[T comparable](
	ctx context.Context,
	target *ccontainer.CContainer[T],
	targetErr *ccontainer.CContainer[*error],
	resolver func(ctx context.Context) (T, func(), error),
) *RefCount[T] {
	return &RefCount[T]{
		ctx:       ctx,
		target:    target,
		targetErr: targetErr,
		resolver:  resolver,
		refs:      make(map[*Ref[T]]struct{}),
	}
}

// AddRef adds a reference to the RefCount container.
// cb is an optional callback to call when the value changes.
func (r *RefCount[T]) AddRef(cb func(val T, err error)) *Ref[T] {
	r.mtx.Lock()
	nref := &Ref[T]{rc: r, cb: cb}
	r.refs[nref] = struct{}{}
	if len(r.refs) == 1 {
		r.startResolve()
	}
	r.mtx.Unlock()
	return nref
}

// removeRef removes a reference and shuts down if no refs remain.
func (r *RefCount[T]) removeRef(ref *Ref[T]) {
	r.mtx.Lock()
	lenBefore := len(r.refs)
	delete(r.refs, ref)
	lenAfter := len(r.refs)
	if lenAfter < lenBefore && lenAfter == 0 {
		r.shutdown()
	}
	r.mtx.Unlock()
}

// shutdown shuts down the resolver, if any.
// expects mtx is locked by caller
func (r *RefCount[T]) shutdown() {
	if r.resolveCtxCancel != nil {
		r.resolveCtxCancel()
		r.resolveCtx, r.resolveCtxCancel = nil, nil
	}
	if r.valueRel != nil {
		r.valueRel()
		r.valueRel = nil
	}
	var empty T
	r.target.SetValue(empty)
}

// startResolve starts the resolve goroutine.
// expects caller to lock mutex.
func (r *RefCount[T]) startResolve() {
	r.resolveCtx, r.resolveCtxCancel = context.WithCancel(r.ctx)
	go r.resolve(r.resolveCtx)
}

// resolve is the goroutine to resolve the value to the container.
func (r *RefCount[T]) resolve(ctx context.Context) {
	val, valRel, err := r.resolver(ctx)
	r.mtx.Lock()
	// assert we are still the resolver
	if r.resolveCtx != ctx {
		if valRel != nil {
			valRel()
		}
		r.mtx.Unlock()
	}

	// store the value and/or error
	if err != nil {
		r.targetErr.SetValue(&err)
	} else {
		r.targetErr.SetValue(nil)
		if r.target != nil {
			r.target.SetValue(val)
		}
	}
	r.valueRel = valRel

	refs := make([]*Ref[T], 0, len(r.refs))
	for ref := range r.refs {
		if ref.cb != nil {
			refs = append(refs, ref)
		}
	}
	r.mtx.Unlock()

	for _, ref := range refs {
		ref.cb(val, err)
	}
}
