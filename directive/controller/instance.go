package controller

import (
	"context"
	"sync"
	"time"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// DirectiveInstance implements the directive instance interface.
type DirectiveInstance struct {
	// ctx is the parent directive controller ctx
	ctx context.Context
	// le is the logger
	le *logrus.Entry
	// dir is the underlying directive.
	dir directive.Directive
	// valueOpts are the directive options
	valueOpts directive.ValueOptions

	// mtx guards below fields
	mtx sync.Mutex

	// refs is the list of references
	refs []*directiveInstanceReference

	// nvalID stores the next value id
	nvalID uint32
	// vals is the map of emitted values
	vals map[uint32]*attachedValue

	// nrelID stores the next rel id
	nrelID uint32
	// rel is the released cb list
	rel map[uint32]func()
	// released flags the instance as released
	released bool
	// unrefDestroyTimer is the timer to call Close()
	unrefDestroyTimer *time.Timer

	attachedResolvers []*attachedResolver
	// attachedResolverCtx is the current context for resolvers
	attachedResolverCtx context.Context
	// attachedResolverCtxCancel is the cancellation func for the resolvers ctx
	attachedResolverCtxCancel context.CancelFunc
	// runningResolvers is the number of executing resolvers
	runningResolvers int
	// nidleID is the next callback id
	nidleID uint32
	// idleCallbacks are the list of idle callback functions
	idleCallbacks map[uint32]func(errs []error)
}

// NewDirectiveInstance constructs a new directive instance with an initial reference.
func NewDirectiveInstance(
	ctx context.Context,
	le *logrus.Entry,
	dir directive.Directive,
	cb directive.ReferenceHandler,
	released func(di *DirectiveInstance),
) (*DirectiveInstance, directive.Reference) {
	i := &DirectiveInstance{
		ctx: ctx,
		dir: dir,
		le:  le,
	}
	i.attachedResolverCtx, i.attachedResolverCtxCancel = context.WithCancel(ctx)
	i.vals = make(map[uint32]*attachedValue)
	i.rel = make(map[uint32]func())
	i.idleCallbacks = make(map[uint32]func(errs []error))
	if released != nil {
		i.rel[0] = func() {
			released(i)
		}
		i.nrelID++
	}
	i.valueOpts = dir.GetValueOptions()
	ref := &directiveInstanceReference{di: i, valCb: cb}
	i.refs = append(i.refs, ref)
	return i, ref
}

// AddReference adds a reference to the directive.
// Will return nil if the directive is already expired.
// Weak references do not contribute to the reference count.
func (r *DirectiveInstance) AddReference(
	cb directive.ReferenceHandler,
	weakRef bool,
) directive.Reference {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.released {
		return nil
	}

	ref := &directiveInstanceReference{
		di:      r,
		valCb:   cb,
		weakRef: weakRef,
	}

	r.markReferenced()

	r.refs = append(r.refs, ref)
	if cb != nil {
		for _, v := range r.vals {
			callHandleValueAdded(r.le, r, v, cb.HandleValueAdded)
		}
	}

	return ref
}

// emitValue emits a new value to all listeners
// returns the value ID and boolean OK
// does NOT expect mtx to be locked by the caller.
func (r *DirectiveInstance) emitValue(v directive.Value) (uint32, bool) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.released {
		return 0, false
	}

	var reject bool
	var cancelOthers bool
	valCount := len(r.vals)
	if maxCount := r.valueOpts.MaxValueCount; maxCount != 0 {
		if valCount >= maxCount && r.valueOpts.MaxValueHardCap {
			// reject value
			reject = true
		} else if valCount+1 >= maxCount {
			cancelOthers = true
		}
	}
	r.nvalID++
	nvid := r.nvalID
	var nav *attachedValue
	if !reject {
		nav = newAttachedValue(nvid, v)
		r.vals[nvid] = nav
	}

	if reject {
		return 0, false
	}

	for _, ref := range r.refs {
		if ref != nil && ref.valCb != nil {
			callHandleValueAdded(r.le, r, nav, ref.valCb.HandleValueAdded)
		}
	}

	if cancelOthers {
		r.cancelResolvers()
	}

	return nvid, true
}

// purgeEmittedValue deletes and returns an emitted value
// returns the value ID and boolean OK
// does NOT expect caller to lock mtx
func (r *DirectiveInstance) purgeEmittedValue(id uint32) (directive.Value, bool) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.released {
		return 0, false
	}

	if r.vals == nil {
		return 0, false
	}
	val, ok := r.vals[id]
	if ok {
		delete(r.vals, id)
		valCount := len(r.vals)
		for _, ref := range r.refs {
			if ref.valCb != nil {
				callHandleValueRemoved(r.le, r, val, ref.valCb.HandleValueRemoved)
			}
		}
		if maxCount := r.valueOpts.MaxValueCount; maxCount != 0 {
			if valCount < maxCount {
				r.restartResolvers()
			}
		} else {
			r.restartResolvers()
		}
	}

	return val, true
}

// GetDirective returns the underlying directive.
func (r *DirectiveInstance) GetDirective() directive.Directive {
	if r == nil {
		return nil
	}

	return r.dir
}

// GetResolverErrors returns a snapshot of any errors returned by resolvers.
func (r *DirectiveInstance) GetResolverErrors() []error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	return r.getResolverErrs()
}

// AddDisposeCallback adds a callback that will be called when the instance
// is disposed, either when Close() is called, or when the reference count
// drops to zero. The callback may occur immediately if the instance is
// already disposed, but will be made in a new goroutine.
// Returns a callback release function.
func (r *DirectiveInstance) AddDisposeCallback(cb func()) func() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.released {
		cb()
		return func() {}
	}

	relid := r.nrelID
	r.nrelID++
	r.rel[relid] = cb
	return func() {
		r.mtx.Lock()
		if r.rel != nil {
			delete(r.rel, relid)
		}
		r.mtx.Unlock()
	}
}

// AddIdleCallback adds a callback that will be called when idle.
// The callback is called exactly once.
// Errs is the list of non-nil resolver errors.
// Returns a callback release function.
func (r *DirectiveInstance) AddIdleCallback(cb directive.IdleCallback) func() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	errs := r.getResolverErrs()

	if r.released {
		cb(errs)
		return func() {}
	}

	if r.runningResolvers == 0 {
		cb(errs)
		return func() {}
	}

	relid := r.nidleID
	r.nidleID++
	r.idleCallbacks[relid] = cb

	return func() {
		r.mtx.Lock()
		if r.idleCallbacks != nil {
			delete(r.idleCallbacks, relid)
		}
		r.mtx.Unlock()
	}
}

// Close cancels the directive instance.
func (r *DirectiveInstance) Close() {
	r.mtx.Lock()
	r.callRel()
	r.mtx.Unlock()
}

// callRel calls the release callbacks.
// caller should lock mtx.
func (r *DirectiveInstance) callRel() {
	if r.released {
		return
	}
	r.released = true

	for _, ref := range r.refs {
		if ref.valCb != nil {
			go ref.valCb.HandleInstanceDisposed(r)
		}
	}
	r.refs = nil
	r.vals = nil

	rel := r.rel
	r.rel = nil
	for _, ref := range rel {
		// go ref()
		go ref()
	}
}

// attachResolver calls and attaches a directive resolver.
func (r *DirectiveInstance) attachResolver(handlerCtx context.Context, res directive.Resolver) {
	inst := r
	r.mtx.Lock()
	defer r.mtx.Unlock()
	ares := newAttachedResolver(inst, res)
	r.attachedResolvers = append(r.attachedResolvers, ares)
	r.runningResolvers++
	ares.pushHandlerContext(r.attachedResolverCtx)
	go ares.execResolver(handlerCtx)
}

// cancelResolvers cancels all children resolvers.
// requires restartResolvers to resume them.
func (r *DirectiveInstance) cancelResolvers() {
	r.attachedResolverCtxCancel()
}

// restartResolvers pushes a fresh resolver context to child resolvers.
// restarts any resolvers that exited before the call.
// expects caller to lock mtx
func (r *DirectiveInstance) restartResolvers() {
	// re-use ctx if not canceled
	nextCtx := r.attachedResolverCtx
	select {
	case <-nextCtx.Done():
		r.attachedResolverCtx, r.attachedResolverCtxCancel = context.WithCancel(r.ctx)
		nextCtx = r.attachedResolverCtx
	default:
	}
	// restart any canceled resolvers
	for _, re := range r.attachedResolvers {
		re.pushHandlerContext(nextCtx)
	}
}

// incrementRunningResolvers is called when a resolver starts.
// does NOT expect caller to lock mtx
func (r *DirectiveInstance) incrementRunningResolvers() {
	r.mtx.Lock()
	r.runningResolvers++
	r.mtx.Unlock()
}

// decrementRunningResolvers is called when a resolver exits.
// does NOT expect caller to lock mtx
func (r *DirectiveInstance) decrementRunningResolvers() {
	r.mtx.Lock()
	if r.runningResolvers != 0 {
		r.runningResolvers--
		if r.runningResolvers == 0 {
			// call idle callbacks
			errs := r.getResolverErrs()
			for id, idleCb := range r.idleCallbacks {
				if idleCb != nil {
					go idleCb(errs)
				}
				delete(r.idleCallbacks, id)
			}
		}
	}
	r.mtx.Unlock()
}

// releaseReference releases a instance reference.
// does NOT expect caller to lock mtx
func (r *DirectiveInstance) releaseReference(dr *directiveInstanceReference) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	found := false
	nonWeakRefCount := 0
	for i := 0; i < len(r.refs); i++ {
		ref := r.refs[i]
		if !found && ref == dr {
			found = true
			r.refs[i] = r.refs[len(r.refs)-1]
			r.refs[len(r.refs)-1] = nil
			r.refs = r.refs[:len(r.refs)-1]
			i--
		} else if ref != nil && !ref.weakRef {
			nonWeakRefCount++
		}

		if found && nonWeakRefCount != 0 {
			break
		}
	}

	if nonWeakRefCount == 0 {
		r.markUnreferenced()
	} else {
		r.markReferenced()
	}
}

// markUnreferenced requires refsMtx is locked, and starts the Close() timer
// caller should lock r.mtx
func (r *DirectiveInstance) markUnreferenced() {
	if r.unrefDestroyTimer == nil {
		udd := r.valueOpts.UnrefDisposeDur
		if udd == 0 {
			// release immediately
			r.callRel()
		} else {
			r.unrefDestroyTimer = time.AfterFunc(udd, r.Close)
		}
	}
}

// getResolverErrs lists resolver errs while caller has mtx locked.
func (r *DirectiveInstance) getResolverErrs() []error {
	errs := make([]error, 0, len(r.attachedResolvers))
	for _, res := range r.attachedResolvers {
		if err := res.valErr; err != nil && err != context.Canceled {
			errs = append(errs, err)
		}
	}
	return errs
}

// markReferenced requires refsMtx is locked, and stops the Close() timer
func (r *DirectiveInstance) markReferenced() {
	if r.unrefDestroyTimer != nil {
		r.unrefDestroyTimer.Stop()
		r.unrefDestroyTimer = nil
	}
}

// callHandleValueAdded calls a handle value added function
func callHandleValueAdded(
	le *logrus.Entry,
	r *DirectiveInstance,
	v *attachedValue,
	fn func(r directive.Instance, v directive.AttachedValue),
) {
	defer func() {
		if err := recover(); err != nil {
			r.le.Errorf("handle value added paniced: %v", err)
		}
	}()
	fn(r, v)
}

// callHandleValueRemoved calls a handle value removed function
func callHandleValueRemoved(
	le *logrus.Entry,
	r *DirectiveInstance,
	v *attachedValue,
	fn func(r directive.Instance, v directive.AttachedValue),
) {
	defer func() {
		if err := recover(); err != nil {
			r.le.Errorf("handle value removed paniced: %v", err)
		}
	}()
	fn(r, v)
}

// _ is a type assertion
var _ directive.Instance = ((*DirectiveInstance)(nil))
