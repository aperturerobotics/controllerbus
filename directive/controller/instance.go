package controller

import (
	"context"
	"net/url"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// directiveInstance implements the directive instance interface.
type directiveInstance struct {
	// released indicates this instance is released
	released atomic.Bool
	// c is the parent directive controller
	c *Controller
	// ctx is canceled when the directive instance expires
	ctx context.Context
	// ctxCancel cancels ctx
	ctxCancel context.CancelFunc
	// id is the id of this instance
	// incremented by 1 each time a directive is added
	id uint32
	// dir is the directive.
	dir directive.Directive
	// valueOpts are the directive options
	valueOpts directive.ValueOptions
	// ident contains the identifier string
	// unset until GetDirectiveIdent is called for the first time.
	ident atomic.Pointer[string]
	// ready indicates we called the initial set of handlers.
	// until ready=false, the directive instance is NOT idle.
	ready bool

	// c.mtx guards below fields

	// destroyTimer is the timer to destroy after 0 refs
	destroyTimer *time.Timer
	// rels contains all release callbacks
	rels []*callback[func()]
	// idles contains all idle callbacks
	idles []*callback[directive.IdleCallback]
	// callingRefs indicates we are currently calling a ref callback
	callingRefCbs bool
	// refCbs is the callbacks queue
	// called when callingRefCbs=true and unlocked
	refCbs []func()
	// refs contains all directive instance references
	// all weak references are at the beginning of the list
	refs []*dirRef
	// res contains attached resolvers
	res []*resolver
	// valCtr is the ID of the next value
	// note: the first value ID is 1, not 0
	valCtr uint32
}

// newDirectiveInstance constructs a new directive instance with an initial reference.
func newDirectiveInstance(
	c *Controller,
	id uint32,
	dir directive.Directive,
	h directive.ReferenceHandler,
) (*directiveInstance, directive.Reference) {
	i := &directiveInstance{
		c:         c,
		id:        id,
		dir:       dir,
		valueOpts: dir.GetValueOptions(),
	}
	i.ctx, i.ctxCancel = context.WithCancel(c.ctx)
	return i, i.addReferenceLocked(h, false)
}

// GetContext returns a context that is canceled when Instance is released.
func (i *directiveInstance) GetContext() context.Context {
	return i.ctx
}

// GetDirective returns the underlying directive.
func (i *directiveInstance) GetDirective() directive.Directive {
	return i.dir
}

// GetDirectiveIdent returns a human-readable string identifying the directive.
//
// Ex: DoSomething or DoSomething<param=foo>
func (i *directiveInstance) GetDirectiveIdent() string {
	existing := i.ident.Load()
	if existing != nil {
		return *existing
	}
	var dirDebugStr string
	debuggable, isDebuggable := i.dir.(directive.Debuggable)
	if isDebuggable {
		if debugVals := debuggable.GetDebugVals(); debugVals != nil {
			dirDebugStr = url.Values(debugVals).Encode()
			dirDebugStr, _ = url.PathUnescape(dirDebugStr)
		}
	}
	dirNameDebugStr := i.dir.GetName()
	if dirDebugStr != "" {
		dirNameDebugStr += "<" + dirDebugStr + ">"
	}
	i.ident.CompareAndSwap(nil, &dirNameDebugStr)
	return dirNameDebugStr
}

// GetResolverErrors returns a snapshot of any errors returned by resolvers.
func (i *directiveInstance) GetResolverErrors() []error {
	i.c.mtx.Lock()
	errs := i.getResolverErrsLocked()
	i.c.mtx.Unlock()
	return errs
}

// getResolverErrsLocked returns the list of resolver errors while c.mtx is locked.
func (i *directiveInstance) getResolverErrsLocked() []error {
	var errs []error
	for _, resolver := range i.res {
		if err := resolver.err; err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// AddReference adds a reference to the directive.
// cb is called for each value.
// cb calls should return immediately.
// returns nil if the directive is already expired.
// If marked as a weak ref, the handler will not count towards the ref count.
func (i *directiveInstance) AddReference(cb directive.ReferenceHandler, weakRef bool) directive.Reference {
	i.c.mtx.Lock()
	defer i.c.mtx.Unlock()
	return i.addReferenceLocked(cb, weakRef)
}

// addReferenceLocked adds a reference while i.c.mtx is locked.
// returns nil if the directive is already expired.
func (i *directiveInstance) addReferenceLocked(cb directive.ReferenceHandler, weakRef bool) directive.Reference {
	ref := &dirRef{di: i, h: cb, weak: weakRef}
	if i.released.Load() {
		ref.released.Store(true)
		if ref.h != nil {
			i.callCallbacksLocked(func() {
				ref.h.HandleInstanceDisposed(i)
			})
		}
		return ref
	}
	firstRef := len(i.refs) == 0
	firstNonWeakRef := !firstRef && !weakRef && i.refs[len(i.refs)-1].weak
	if weakRef {
		i.refs = append([]*dirRef{ref}, i.refs...)
	} else {
		i.refs = append(i.refs, ref)
	}
	var cbs []func()
	if cb != nil {
		for _, res := range i.res {
			for _, val := range res.vals {
				val := val
				cbs = append(cbs, func() {
					cb.HandleValueAdded(i, val)
				})
			}
		}
	}
	if firstRef || firstNonWeakRef {
		i.handleReferencedLocked(weakRef)
	}
	i.callCallbacksLocked(cbs...)
	return ref
}

// removeReferenceLocked removes a reference while i.c.mtx is locked.
func (i *directiveInstance) removeReferenceLocked(ref *dirRef) {
	for idx, iref := range i.refs {
		if iref == ref {
			i.refs = append(i.refs[:idx], i.refs[idx+1:]...)
			ref.released.Store(true)
			onlyWeakRefs := len(i.refs) != 0 && i.refs[len(i.refs)-1].weak
			if onlyWeakRefs || len(i.refs) == 0 {
				i.handleUnreferencedLocked(onlyWeakRefs)
			}
			break
		}
	}
}

// handleReferencedLocked handles when we reach 1 reference while i.c.mtx is locked
// also called when we reach 1 non-weak reference.
func (i *directiveInstance) handleReferencedLocked(weakRef bool) {
	if i.released.Load() {
		return
	}
	// cancel dispose callback, but only if this isn't a "weak" ref
	if !weakRef && i.destroyTimer != nil {
		_ = i.destroyTimer.Stop()
		i.destroyTimer = nil
	}
	// start all resolvers
	for _, res := range i.res {
		res.updateContextLocked(&i.ctx)
	}
}

// countRunningResolversLocked counts the number of non-idle running resolvers.
func (i *directiveInstance) countRunningResolversLocked() int {
	var count int
	for _, res := range i.res {
		if !res.idle && res.ctx != nil && !res.exited {
			count++
		}
	}
	return count
}

// anyValuesLocked checks if any values have been associated with the resolvers.
func (i *directiveInstance) anyValuesLocked() bool {
	for _, res := range i.res {
		if len(res.vals) != 0 {
			return true
		}
	}
	return false
}

// handleUnreferencedLocked handles when we reach 0 references while i.c.mtx is locked.
func (i *directiveInstance) handleUnreferencedLocked(anyWeakRefs bool) {
	if i.released.Load() || i.destroyTimer != nil {
		return
	}
	disposeDur := i.valueOpts.UnrefDisposeDur
	disposeEmptyImmediate := i.valueOpts.UnrefDisposeEmptyImmediate
	if disposeDur == 0 || (disposeEmptyImmediate && !i.anyValuesLocked()) {
		i.removeLocked(-1)
	} else {
		var destroyTimer *time.Timer
		destroyTimer = time.AfterFunc(disposeDur, func() {
			if i.released.Load() {
				return
			}
			i.c.mtx.Lock()
			if !i.released.Load() && destroyTimer == i.destroyTimer {
				i.destroyTimer = nil
				i.removeLocked(-1)
			}
			i.c.mtx.Unlock()
		})
		i.destroyTimer = destroyTimer

		// cancel all resolvers that have no values
		// ... but only if there are not any "weak" references
		if !anyWeakRefs {
			for _, res := range i.res {
				if len(res.vals) == 0 {
					res.updateContextLocked(nil)
				}
			}
		}
	}
}

// handleIdleLocked is called when the instance becomes idle while i.c.mtx is locked
func (i *directiveInstance) handleIdleLocked() {
	if len(i.idles) == 0 {
		return
	}
	errs := i.getResolverErrsLocked()
	var cbs []func()
	for _, idle := range i.idles {
		if !idle.released.Load() && idle.cb != nil {
			idle := idle
			cbs = append(cbs, func() {
				idle.cb(errs)
			})
		}
	}
	i.callCallbacksLocked(cbs...)
}

// removeReleaseCallbackLocked removes a release callback while i.c.mtx is locked.
func (i *directiveInstance) removeReleaseCallbackLocked(cb *callback[func()]) {
	removeFromCallbacks(i.rels, cb)
}

// removeIdleCallbackLocked removes a idle callback while i.c.mtx is locked.
func (i *directiveInstance) removeIdleCallbackLocked(cb *callback[directive.IdleCallback]) {
	removeFromCallbacks(i.idles, cb)
}

// addValueLocked adds a value to the instance while i.c.mtx is locked
func (i *directiveInstance) addValueLocked(res *resolver, val directive.Value) (uint32, bool) {
	maxVals := i.valueOpts.MaxValueCount
	if maxVals != 0 && i.valueOpts.MaxValueHardCap && len(res.vals) >= maxVals {
		return 0, false
	}

	i.valCtr++
	vid := i.valCtr
	v := &value{id: vid, val: val}
	res.vals = append(res.vals, v)

	var cbs []func()
	for _, ref := range i.refs {
		ref := ref
		if !ref.released.Load() && ref.h != nil {
			cbs = append(cbs, func() {
				ref.h.HandleValueAdded(i, v)
			})
		}
	}
	i.callCallbacksLocked(cbs...)

	if maxVals != 0 && len(res.vals) >= maxVals {
		// we have enough values, reaching the maxVals cap
		// mark all resolvers as idle (so ExecOneOff returns)
		// cancel contexts for resolvers with no values
		runningResolvers := i.countRunningResolversLocked()
		if runningResolvers != 0 {
			for _, res := range i.res {
				res.idle = true
			}
			i.handleIdleLocked()
		}
	}

	return vid, true
}

// removeValueLocked removes a value from the instance while i.c.mtx is locked
func (i *directiveInstance) removeValueLocked(res *resolver, valID uint32) (directive.Value, bool) {
	for idx := 0; idx < len(res.vals); idx++ {
		val := res.vals[idx]
		if val.id == valID {
			res.vals = append(res.vals[:idx], res.vals[idx+1:]...)
			i.onValueRemovedLocked(res, val)
			return val.val, true
		}
	}
	return nil, false
}

// onValueRemovedLocked is called by removeValueLocked after removing an item.
func (i *directiveInstance) onValueRemovedLocked(res *resolver, val *value) {
	maxVals := i.valueOpts.MaxValueCount
	if maxVals != 0 && len(res.vals)+1 == maxVals {
		// restart resolvers
		for _, res := range i.res {
			res.updateContextLocked(&i.ctx)
		}
	}

	var cbs []func()
	for _, ref := range i.refs {
		ref := ref
		if !ref.released.Load() && ref.h != nil {
			cbs = append(cbs, func() {
				ref.h.HandleValueRemoved(i, val)
			})
		}
	}
	i.callCallbacksLocked(cbs...)
}

// AddDisposeCallback adds a callback that will be called when the instance
// is disposed, either when Close() is called, or when the reference count
// drops to zero. The callback may occur immediately if the instance is
// already disposed, but will be made in a new goroutine.
// Returns a callback release function.
func (i *directiveInstance) AddDisposeCallback(cb func()) func() {
	i.c.mtx.Lock()
	if i.released.Load() {
		i.c.mtx.Unlock()
		cb()
		return func() {}
	}
	rel := newCallback(cb)
	i.rels = append(i.rels, rel)
	i.c.mtx.Unlock()
	return func() {
		if !rel.released.Swap(true) {
			i.c.mtx.Lock()
			i.removeReleaseCallbackLocked(rel)
			i.c.mtx.Unlock()
		}
	}
}

// AddIdleCallback adds a callback that will be called when idle.
// The callback is called exactly once.
// Returns a callback release function.
func (i *directiveInstance) AddIdleCallback(cb directive.IdleCallback) func() {
	i.c.mtx.Lock()
	rel := newCallback(cb)
	i.idles = append(i.idles, rel)
	isIdle := i.ready && i.countRunningResolversLocked() == 0
	var errs []error
	if isIdle {
		errs = i.getResolverErrsLocked()
	}
	i.c.mtx.Unlock()
	if isIdle {
		go cb(errs)
	}
	return func() {
		if !rel.released.Swap(true) {
			i.c.mtx.Lock()
			i.removeIdleCallbackLocked(rel)
			i.c.mtx.Unlock()
		}
	}
}

// Close cancels the directive instance.
func (i *directiveInstance) Close() {
	i.c.mtx.Lock()
	defer i.c.mtx.Unlock()
	if !i.released.Swap(true) {
		i.removeLocked(-1)
	}
}

// CloseIfUnreferenced cancels the directive instance if there are no refs.
//
// This bypasses the unref dispose timer.
// If inclWeakRefs=true, keeps the instance if there are any weak refs.
// Returns if the directive instance was closed.
func (i *directiveInstance) CloseIfUnreferenced(inclWeakRefs bool) bool {
	i.c.mtx.Lock()
	defer i.c.mtx.Unlock()
	if i.released.Load() {
		return true
	}
	hasRefs := len(i.refs) != 0
	if hasRefs && !inclWeakRefs {
		// if the last ref in the list is a weak ref
		// there are only weak refs, mark as unreferenced.
		hasRefs = !i.refs[len(i.refs)-1].weak
	}
	if !hasRefs && !i.released.Swap(true) {
		i.removeLocked(-1)
	}
	return i.released.Load()
}

// callHandlerUnlocked calls the HandleDirective function while i.c.mtx is not locked.
//
// expects c.mtx to not be locked
func (i *directiveInstance) callHandlerUnlocked(handler *handler) ([]*resolver, error) {
	resolvers, err := handler.h.HandleDirective(i.ctx, i)
	if err != nil {
		return nil, err
	}
	out := make([]*resolver, len(resolvers))
	for x, resolver := range resolvers {
		out[x] = newResolver(i, handler, resolver)
	}
	return out, nil
}

// attachStartResolverLocked attaches and starts a resolver while i.c.mtx is locked
func (i *directiveInstance) attachStartResolverLocked(res *resolver) {
	i.res = append(i.res, res)
	// start resolver if len(refs) != 0
	if len(i.refs) != 0 {
		res.updateContextLocked(&i.ctx)
	}
}

// removeLocked removes the directive instance while i.c.mtx is locked.
// calls the directive removed callbacks
// if diIdx != -1, uses the index as the one to remove.
func (i *directiveInstance) removeLocked(diIdx int) {
	// mark released
	i.released.Store(true)
	i.ctxCancel()
	if i.destroyTimer != nil {
		_ = i.destroyTimer.Stop()
		i.destroyTimer = nil
	}
	// determine index in list (if necessary)
	if diIdx < 0 {
		for idx, di := range i.c.dir {
			if di == i {
				diIdx = idx
				break
			}
		}
		if diIdx < 0 {
			// not found
			return
		}
	}
	// remove from list of instances
	i.logger().Debug("removed directive")
	i.c.dir = append(i.c.dir[:diIdx], i.c.dir[diIdx+1:]...)
	// clear everything else
	var cbs []func()
	for _, ref := range i.refs {
		ref := ref
		if !ref.released.Swap(true) && ref.h != nil {
			cbs = append(cbs, func() {
				ref.h.HandleInstanceDisposed(i)
			})
		}
	}
	for _, cb := range i.rels {
		if !cb.released.Swap(true) && cb.cb != nil {
			cbs = append(cbs, cb.cb)
		}
	}
	i.rels = nil
	i.callCallbacksLocked(cbs...)
	i.refs = nil
	for _, res := range i.res {
		if res.ctxCancel != nil {
			res.ctxCancel()
		}
		res.ctx, res.ctxCancel = nil, nil
	}
	i.res = nil
	for _, idle := range i.idles {
		idle.released.Store(true)
	}
	i.idles = nil
}

// removeHandlerLocked removes all resolvers associated with the handler.
// caller locks c.mtx
func (i *directiveInstance) removeHandlerLocked(hnd *handler) {
	for idx := 0; idx < len(i.res); idx++ {
		res := i.res[idx]
		if res.hnd == hnd {
			i.removeResolverLocked(idx, res)
			idx--
		}
	}
}

// removeResolverLocked removes the given resolver while c.mtx is locked.
// cancels the resolver and removes all values associated with it.
// if resIdx >= 0 removes that index from i.res, otherwise searches.
func (i *directiveInstance) removeResolverLocked(resIdx int, rres *resolver) {
	// search for the resolver in the list if necessary
	if resIdx < 0 {
		for i, res := range i.res {
			if res == rres {
				resIdx = i
				break
			}
		}
		if resIdx < 0 {
			return
		}
	}

	// remove the resolver from the list
	i.res = append(i.res[:resIdx], i.res[resIdx+1:]...)
	// cancel the resolver
	rres.updateContextLocked(nil)
	// remove values associated with the resolver
	var cbs []func()
	for _, val := range rres.vals {
		val := val
		for _, ref := range i.refs {
			ref := ref
			if ref.h != nil && !ref.released.Load() {
				cbs = append(cbs, func() {
					ref.h.HandleValueRemoved(i, val)
				})
			}
		}
	}
	i.callCallbacksLocked(cbs...)
	rres.vals = nil
}

// callCallbacksLocked calls the refsCbs list or adds to queue
func (i *directiveInstance) callCallbacksLocked(cbs ...func()) {
	if len(cbs)+len(i.refCbs) == 0 {
		return
	}
	i.refCbs = append(i.refCbs, cbs...)
	if i.callingRefCbs {
		return
	}
	i.callingRefCbs = true
	cbs = i.refCbs
	i.refCbs = nil
	for len(cbs) != 0 {
		i.c.mtx.Unlock()
		for _, cb := range cbs {
			func() {
				defer func() {
					if rerr := recover(); rerr != nil {
						debug.PrintStack()
						e, eOk := rerr.(error)
						le := i.logger()
						if eOk {
							le = le.WithError(e)
						} else {
							le = le.WithField("error", rerr)
						}
						le.Error("callback paniced")
					}
				}()
				cb()
			}()
		}
		i.c.mtx.Lock()
		cbs = i.refCbs
		i.refCbs = nil
	}
	i.callingRefCbs = false
}

// logger returns the logger for this instance
// no locks required
func (i *directiveInstance) logger() *logrus.Entry {
	return i.c.le.WithField("directive", i.GetDirectiveIdent())
}

// _ is a type assertion
var _ directive.Instance = ((*directiveInstance)(nil))
