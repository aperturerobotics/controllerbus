package controller

import (
	"context"
	"sync"

	"github.com/aperturerobotics/controllerbus/directive"
)

// DirectiveInstance implements the directive instance interface.
type DirectiveInstance struct {
	// ctx is the parent directive controller ctx
	ctx context.Context
	// dir is the underlying directive.
	dir directive.Directive
	// valueOpts are the directive options
	valueOpts directive.ValueOptions

	// refsMtx guards refs
	refsMtx sync.Mutex
	// refs is the list of references
	refs []*directiveInstanceReference

	// valsMtx guards vals
	valsMtx sync.Mutex
	// nvalID stores the next value id
	nvalID uint32
	// vals is the map of emitted values
	vals map[uint32]directive.Value

	// relMtx guards rel
	relMtx sync.Mutex
	// nrelID stores the next rel id
	nrelID uint32
	// rel is the released cb list
	rel map[uint32]func()
	// released flags the instance as released
	released bool

	// attachedResolversMtx guards attachedResolvers
	attachedResolversMtx sync.Mutex
	attachedResolvers    []*attachedResolver
	// attachedResolverCtx is the current context for resolvers
	attachedResolverCtx context.Context
	// attachedResolverCtxCancel is the cancellation func for the resolvers ctx
	attachedResolverCtxCancel context.CancelFunc
	// runningResolvers is the number of executing resolvers
	runningResolvers int
	// nidleID is the next callback id
	nidleID uint32
	// idleCallbacks are the list of idle callback functions
	idleCallbacks map[uint32]func()
}

// NewDirectiveInstance constructs a new directive instance with an initial reference.
func NewDirectiveInstance(
	ctx context.Context,
	dir directive.Directive,
	cb directive.ReferenceHandler,
	released func(),
) (*DirectiveInstance, directive.Reference) {
	i := &DirectiveInstance{ctx: ctx, dir: dir}
	i.attachedResolverCtx, i.attachedResolverCtxCancel = context.WithCancel(ctx)
	i.vals = make(map[uint32]directive.Value)
	i.rel = make(map[uint32]func())
	i.idleCallbacks = make(map[uint32]func())
	if released != nil {
		i.rel[0] = released
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
	r.relMtx.Lock()
	defer r.relMtx.Unlock()

	if r.released {
		return nil
	}

	ref := &directiveInstanceReference{
		di:      r,
		valCb:   cb,
		weakRef: weakRef,
	}

	r.refsMtx.Lock()
	r.valsMtx.Lock()
	// avoid calling cb() before returning addreference
	defer func() {
		go func() {
			// cb should not block
			for _, v := range r.vals {
				cb.HandleValueAdded(r, v)
			}
			r.valsMtx.Unlock()
		}()
	}()

	r.refs = append(r.refs, ref)
	r.refsMtx.Unlock()

	return ref
}

// emitValue emits a new value to all listeners
// returns the value ID and boolean OK
func (r *DirectiveInstance) emitValue(v directive.Value) (uint32, bool) {
	r.relMtx.Lock()
	defer r.relMtx.Unlock()

	if r.released {
		return 0, false
	}

	var reject bool
	var cancelOthers bool
	r.valsMtx.Lock()
	valCount := len(r.vals)
	if maxCount := r.valueOpts.MaxValueCount; maxCount != 0 {
		if valCount >= maxCount && r.valueOpts.MaxValueHardCap {
			// reject value
			reject = true
		} else if valCount+1 >= maxCount {
			cancelOthers = true
		}
	}
	nvid := r.nvalID
	if !reject {
		r.nvalID++
		r.vals[nvid] = v
	}
	r.valsMtx.Unlock()

	if reject {
		return 0, false
	}

	r.refsMtx.Lock()
	for _, ref := range r.refs {
		if ref != nil && ref.valCb != nil {
			ref.valCb.HandleValueAdded(r, v)
		}
	}
	r.refsMtx.Unlock()

	if cancelOthers {
		r.cancelResolvers()
	}

	return nvid, true
}

// purgeEmittedValue deletes and returns an emitted value
// returns the value ID and boolean OK
func (r *DirectiveInstance) purgeEmittedValue(id uint32) (directive.Value, bool) {
	r.relMtx.Lock()
	defer r.relMtx.Unlock()

	if r.released {
		return 0, false
	}

	r.valsMtx.Lock()
	if r.vals == nil {
		r.valsMtx.Unlock()
		return 0, false
	}
	val, ok := r.vals[id]
	delete(r.vals, id)
	valCount := len(r.vals)
	r.valsMtx.Unlock()

	if maxCount := r.valueOpts.MaxValueCount; maxCount != 0 {
		if valCount+1 == maxCount {
			r.restartResolvers()
		}
	}

	if ok {
		r.refsMtx.Lock()
		for _, ref := range r.refs {
			if ref.valCb != nil {
				ref.valCb.HandleValueRemoved(r, val)
			}
		}
		r.refsMtx.Unlock()
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

// AddDisposeCallback adds a callback that will be called when the instance
// is disposed, either when Close() is called, or when the reference count
// drops to zero. The callback may occur immediately if the instance is
// already disposed, but will be made in a new goroutine.
// Returns a callback release function.
func (r *DirectiveInstance) AddDisposeCallback(cb func()) func() {
	r.relMtx.Lock()
	defer r.relMtx.Unlock()

	if r.released {
		go cb()
		return func() {}
	}

	relid := r.nrelID
	r.nrelID++
	r.rel[relid] = cb
	return func() {
		r.relMtx.Lock()
		if r.rel != nil {
			delete(r.rel, relid)
		}
		r.relMtx.Unlock()
	}
}

// AddIdleCallback adds a callback that will be called when idle.
// The callback is called exactly once.
// Returns a callback release function.
func (r *DirectiveInstance) AddIdleCallback(cb func()) func() {
	r.relMtx.Lock()
	defer r.relMtx.Unlock()

	if r.released {
		go cb()
		return func() {}
	}

	r.attachedResolversMtx.Lock()
	defer r.attachedResolversMtx.Unlock()

	if r.runningResolvers == 0 {
		go cb()
		return func() {}
	}

	relid := r.nidleID
	r.nidleID++
	r.idleCallbacks[relid] = cb

	return func() {
		r.attachedResolversMtx.Lock()
		if r.idleCallbacks != nil {
			delete(r.idleCallbacks, relid)
		}
		r.attachedResolversMtx.Unlock()
	}
}

// Close cancels the directive instance.
func (r *DirectiveInstance) Close() {
	r.callRel()
}

// callRel calls the release callbacks.
func (r *DirectiveInstance) callRel() {
	r.relMtx.Lock()
	defer r.relMtx.Unlock()

	if r.released {
		return
	}

	r.released = true
	r.refsMtx.Lock()
	for _, ref := range r.refs {
		go ref.valCb.HandleInstanceDisposed(r)
	}
	r.refs = nil
	r.refsMtx.Unlock()

	r.valsMtx.Lock()
	r.vals = nil
	r.valsMtx.Unlock()

	for _, ref := range r.rel {
		go ref()
	}
	r.rel = nil
}

// attachResolver calls and attaches a directive resolver.
func (r *DirectiveInstance) attachResolver(handlerCtx context.Context, res directive.Resolver) {
	inst := r
	r.attachedResolversMtx.Lock()
	ares := newAttachedResolver(inst, res)
	r.attachedResolvers = append(r.attachedResolvers, ares)
	r.runningResolvers++
	ares.pushHandlerContext(r.attachedResolverCtx)
	go func() {
		err := ares.execResolver(handlerCtx)
		// TODO: handle resolver error (log it)
		_ = err
		r.attachedResolversMtx.Lock()
		for i, ai := range r.attachedResolvers {
			if ai == ares {
				r.attachedResolvers[i] = r.attachedResolvers[len(r.attachedResolvers)-1]
				r.attachedResolvers[len(r.attachedResolvers)-1] = nil
				r.attachedResolvers = r.attachedResolvers[:len(r.attachedResolvers)-1]
				break
			}
		}
		r.attachedResolversMtx.Unlock()
	}()
	r.attachedResolversMtx.Unlock()
}

// cancelResolvers cancels all children resolvers.
// requires restartResolvers to resume them.
func (r *DirectiveInstance) cancelResolvers() {
	r.attachedResolverCtxCancel()
}

// restartResolvers pushes a fresh resolver context to child resolvers.
// if the resolver context is not already canceled, does nothing
func (r *DirectiveInstance) restartResolvers() {
	r.attachedResolversMtx.Lock()
	select {
	default:
		r.attachedResolversMtx.Unlock()
		return
	case <-r.attachedResolverCtx.Done():
	}
	r.attachedResolverCtx, r.attachedResolverCtxCancel = context.WithCancel(r.ctx)
	for _, re := range r.attachedResolvers {
		re.pushHandlerContext(r.attachedResolverCtx)
	}
	r.attachedResolversMtx.Unlock()
}

// incrementRunningResolvers is called when a resolver starts.
func (r *DirectiveInstance) incrementRunningResolvers() {
	r.attachedResolversMtx.Lock()
	r.runningResolvers++
	r.attachedResolversMtx.Unlock()
}

// decrementRunningResolvers is called when a resolver exits.
func (r *DirectiveInstance) decrementRunningResolvers() {
	r.attachedResolversMtx.Lock()
	if r.runningResolvers != 0 {
		r.runningResolvers--
		if r.runningResolvers == 0 {
			for id, idleCb := range r.idleCallbacks {
				if idleCb != nil {
					go idleCb()
				}
				delete(r.idleCallbacks, id)
			}
		}
	}
	r.attachedResolversMtx.Unlock()
}

// _ is a type assertion
var _ directive.Instance = ((*DirectiveInstance)(nil))
