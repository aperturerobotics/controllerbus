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
	// ctx contains the root context
	// can be nil
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
// ctx, target and targetErr can be empty
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

// AccessRefCount adds a reference to the RefCount while the fn executes.
func AccessRefCount[T comparable](
	ctx context.Context,
	rc *RefCount[T],
	cb func(T) error,
) error {
	valCh := make(chan T, 1)
	errCh := make(chan error, 1)
	ref := rc.AddRef(func(val T, err error) {
		if err != nil {
			select {
			case errCh <- err:
			default:
			}
		} else {
			select {
			case valCh <- val:
			default:
			}
		}
	})
	defer ref.Release()

	select {
	case <-ctx.Done():
		return context.Canceled
	case err := <-errCh:
		return err
	case val := <-valCh:
		return cb(val)
	}
}

// WaitRefCount waits for a RefCount container handling errors.
// targetErr can be nil
func WaitRefCount[T comparable](
	ctx context.Context,
	target *ccontainer.CContainer[T],
	targetErr *ccontainer.CContainer[*error],
) (T, error) {
	var errCh chan error
	if targetErr != nil {
		errCh = make(chan error, 1)
		go func() {
			outErr, _ := targetErr.WaitValue(ctx, errCh)
			if outErr != nil && *outErr != nil {
				select {
				case errCh <- *outErr:
				default:
				}
			}
		}()
	}
	return target.WaitValue(ctx, errCh)
}

// SetContext updates the context to use for the RefCount container resolution.
// If ctx=nil the RefCount will wait until ctx != nil to start.
// This also restarts resolution, if there are any refs.
func (r *RefCount[T]) SetContext(ctx context.Context) {
	r.mtx.Lock()
	if r.ctx != ctx {
		r.ctx = ctx
		r.startResolve()
	}
	r.mtx.Unlock()
}

// AddRef adds a reference to the RefCount container.
// cb is an optional callback to call when the value changes.
func (r *RefCount[T]) AddRef(cb func(val T, err error)) *Ref[T] {
	r.mtx.Lock()
	nref := &Ref[T]{rc: r, cb: cb}
	r.refs[nref] = struct{}{}
	if len(r.refs) == 1 {
		r.startResolve()
	} else {
		var empty T
		if val := r.target.GetValue(); val != empty {
			nref.cb(val, nil)
		} else if err := r.targetErr.GetValue(); err != nil && *err != nil {
			nref.cb(empty, *err)
		}
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
	if r.resolveCtxCancel != nil {
		r.resolveCtxCancel()
	}
	if r.ctx == nil || len(r.refs) == 0 {
		r.resolveCtxCancel = nil
		return
	}
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
		if r.targetErr != nil {
			r.targetErr.SetValue(&err)
		}
	} else {
		if r.targetErr != nil {
			r.targetErr.SetValue(nil)
		}
		if r.target != nil {
			r.target.SetValue(val)
		}
	}
	r.valueRel = valRel
	for ref := range r.refs {
		if ref.cb != nil {
			ref.cb(val, err)
		}
	}
	r.mtx.Unlock()
}
