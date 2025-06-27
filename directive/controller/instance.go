package controller

import (
	"context"
	"net/url"
	"runtime/debug"
	"slices"
	"sync/atomic"
	"time"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/pkg/errors"
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

	// c.mtx guards below fields

	// ready indicates we called the initial set of handlers.
	// until ready=false, the directive instance is NOT idle.
	ready bool
	// destroyTimer is the timer to destroy after 0 refs
	destroyTimer *time.Timer
	// rels contains all release callbacks
	rels []*callback[func()]
	// idles contains all idle callbacks
	idles []*callback[directive.IdleCallback]
	// stateChanged contains all state changed callbacks
	stateChanged []*callback[directive.StateCallback]
	// stateChangedSnapshot contains the current state snapshot for stateChanged callbacks.
	// only populated when len(stateChanged) != 0 for performance
	// when this changes, call the stateChanged callbacks
	stateChangedSnapshot directiveStateSnapshot
	// stateChangedSkip indicates we are already checking for state changes in this call stack
	stateChangedSkip bool
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
	// idle indicates there are no running resolvers
	// call the idle callbacks if/when this changes
	idle bool
	// full indicates we have more than MaxValueCap values.
	full bool
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

// directiveStateSnapshot is a snapshot of the state for a StateCallback
type directiveStateSnapshot struct {
	set          bool // set to false if not populated
	idle         bool
	errs         []error
	attachedVals []directive.AttachedValue
}

// getDirectiveStateSnapshotLocked returns the latest snapshot or populates it if empty.
func (i *directiveInstance) getDirectiveStateSnapshotLocked() directiveStateSnapshot {
	if i.stateChangedSnapshot.set {
		return i.stateChangedSnapshot
	}

	// populate values
	i.populateDirectiveStateSnapshotLocked()
	return i.stateChangedSnapshot
}

// populateDirectiveStateSnapshotLocked re-populates the entire state snapshot field.
func (i *directiveInstance) populateDirectiveStateSnapshotLocked() {
	i.stateChangedSnapshot.idle = i.ready && i.idle
	i.stateChangedSnapshot.errs = i.getResolverErrsLocked()
	i.stateChangedSnapshot.attachedVals = i.getResolverAttachedValsLocked()
	i.stateChangedSnapshot.set = true
}

// deferCheckStateChanged returns a function to defer state change checking.
// Usage: defer i.deferCheckStateChanged()()
func (i *directiveInstance) deferCheckStateChanged() func() {
	// If already skipping or no state callbacks, return empty function
	if i.stateChangedSkip || len(i.stateChanged) == 0 {
		return func() {}
	}

	// Set skip flag and return function to check state changes
	i.stateChangedSkip = true
	return func() {
		i.maybeCallStateChanged()
		i.stateChangedSkip = false
	}
}

// maybeCallStateChanged checks if state changed and calls callbacks if so.
func (i *directiveInstance) maybeCallStateChanged() {
	if len(i.stateChanged) == 0 || i.stateChangedSnapshot.set {
		return
	}

	// Populate current state snapshot
	i.populateDirectiveStateSnapshotLocked()

	// Call callbacks
	var cbs []func()
	for _, stateCb := range i.stateChanged {
		if !stateCb.released.Load() && stateCb.cb != nil {
			stateCbFn := stateCb.cb
			snapshot := i.stateChangedSnapshot
			cbs = append(cbs, func() {
				stateCbFn(snapshot.idle, snapshot.errs, snapshot.attachedVals)
			})
		}
	}
	i.callCallbacksLocked(cbs...)
}

// getResolverAttachedValsLocked returns the list of values while c.mtx is locked.
func (i *directiveInstance) getResolverAttachedValsLocked() []directive.AttachedValue {
	// count size for slice
	var numVals int
	for _, resolver := range i.res {
		numVals += len(resolver.vals)
	}

	vals := make([]directive.AttachedValue, 0, numVals)
	for _, resolver := range i.res {
		for _, val := range resolver.vals {
			vals = append(vals, val)
		}
	}

	return vals
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
	if !weakRef && (firstRef || firstNonWeakRef) {
		i.handleReferencedLocked()
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
			anyNonWeakRefs := len(i.refs) != 0 && !i.refs[len(i.refs)-1].weak
			if !anyNonWeakRefs {
				i.handleUnreferencedLocked()
			}
			break
		}
	}
}

// handleReferencedLocked handles when we reach 1 non-weak reference while i.c.mtx is locked
func (i *directiveInstance) handleReferencedLocked() {
	if i.released.Load() {
		return
	}
	// cancel dispose callback
	if i.destroyTimer != nil {
		_ = i.destroyTimer.Stop()
		i.destroyTimer = nil
	}
}

// countRunningResolversLocked counts the number of non-idle running resolvers.
func (i *directiveInstance) countRunningResolversLocked() int {
	var count int
	for _, res := range i.res {
		if !res.idle && !res.exited {
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

// countValuesLocked counts the number of values attached to the resolvers.
func (i *directiveInstance) countValuesLocked() int {
	var count int
	for _, res := range i.res {
		count += len(res.vals)
	}
	return count
}

// handleUnreferencedLocked handles when we reach 0 non-weak references while i.c.mtx is locked.
func (i *directiveInstance) handleUnreferencedLocked() {
	if i.released.Load() {
		return
	}
	disposeDur := i.valueOpts.UnrefDisposeDur
	disposeEmptyImmediate := i.valueOpts.UnrefDisposeEmptyImmediate
	if disposeDur == 0 || (disposeEmptyImmediate && !i.anyValuesLocked()) {
		i.removeLocked(-1)
	} else if i.destroyTimer == nil {
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
	}
}

// handleFullLocked handles reaching the max value cap.
func (i *directiveInstance) handleFullLocked() {
	// already full
	if i.full {
		return
	}

	// mark as full
	i.full = true

	// we have enough values, reaching the maxVals cap
	// mark resolvers with values as idle
	// cancel non-idle resolvers with no values
	for _, res := range i.res {
		if res.idle {
			continue
		}
		if !res.exited && len(res.vals) == 0 {
			res.stopped = true
			res.updateContextLocked(nil)
		} else {
			res.idle = true
		}
	}

	// check if the directive became idle
	i.handleIdleStateLocked()
}

// handleNotFullLocked handles going below the max value cap.
func (i *directiveInstance) handleNotFullLocked() {
	if !i.full {
		return
	}

	i.full = false
	// restart resolvers that exited with errors and/or were killed
	for _, res := range i.res {
		if (res.exited && res.err != nil) || res.stopped {
			res.updateContextLocked(&i.ctx)
		}
	}
}

// checkFullLocked checks if we have reached the maximum value cap.
func (i *directiveInstance) checkFullLocked() {
	maxVals := i.valueOpts.MaxValueCount
	if maxVals == 0 {
		return
	}

	valueCount := i.countValuesLocked()
	isFull := valueCount >= maxVals
	if isFull {
		i.handleFullLocked()
	} else {
		i.handleNotFullLocked()
	}
}

// handleIdleStateLocked checks if the idle state of the instance changes & handles that if so.
func (i *directiveInstance) handleIdleStateLocked() {
	// true if there are no running resolvers and the instance is ready.
	idle := i.countRunningResolversLocked() == 0 && i.ready
	if i.idle == idle {
		return
	}

	defer i.deferCheckStateChanged()()
	i.idle = idle
	i.stateChangedSnapshot = directiveStateSnapshot{} // mark state as changed

	if len(i.idles) == 0 {
		return
	}

	errs := i.getResolverErrsLocked()
	var cbs []func()
	for _, idleCb := range i.idles {
		if !idleCb.released.Load() && idleCb.cb != nil {
			idleCbFn := idleCb.cb
			cbs = append(cbs, func() {
				idleCbFn(idle, errs)
			})
		}
	}
	i.callCallbacksLocked(cbs...)
}

// removeReleaseCallbackLocked removes a release callback while i.c.mtx is locked.
func (i *directiveInstance) removeReleaseCallbackLocked(cb *callback[func()]) {
	i.rels = removeFromCallbacks(i.rels, cb)
}

// removeIdleCallbackLocked removes a idle callback while i.c.mtx is locked.
func (i *directiveInstance) removeIdleCallbackLocked(cb *callback[directive.IdleCallback]) {
	i.idles = removeFromCallbacks(i.idles, cb)
}

// removeStateCallbackLocked removes a state callback while i.c.mtx is locked.
func (i *directiveInstance) removeStateCallbackLocked(cb *callback[directive.StateCallback]) {
	i.stateChanged = removeFromCallbacks(i.stateChanged, cb)
}

// addValueLocked adds a value to the instance while i.c.mtx is locked
func (i *directiveInstance) addValueLocked(res *resolver, val directive.Value) (uint32, bool) {
	defer i.deferCheckStateChanged()()

	maxVals := i.valueOpts.MaxValueCount
	if maxVals != 0 && i.countValuesLocked() >= maxVals && i.valueOpts.MaxValueHardCap {
		return 0, false
	}

	i.valCtr++
	vid := i.valCtr

	v := &value{id: vid, val: val}
	res.vals = append(res.vals, v)
	i.stateChangedSnapshot = directiveStateSnapshot{} // mark state as changed

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
	i.checkFullLocked()

	return vid, true
}

// removeValueLocked removes a value from the instance while i.c.mtx is locked
func (i *directiveInstance) removeValueLocked(res *resolver, valID uint32) (directive.Value, bool) {
	for idx := 0; idx < len(res.vals); idx++ {
		val := res.vals[idx]
		if val.id == valID {
			res.vals = append(res.vals[:idx], res.vals[idx+1:]...)
			i.onValuesRemovedLocked(val)
			return val.val, true
		}
	}
	return nil, false
}

// addValueRemovedCallbackLocked adds a callback to be called when the given value id is removed.
// returns nil, nil, false if the value was not found
func (i *directiveInstance) addValueRemovedCallbackLocked(res *resolver, valID uint32, cb func()) (directive.Value, func(), bool) {
	for idx := 0; idx < len(res.vals); idx++ {
		val := res.vals[idx]
		if val.id == valID {
			val.removeCallbackCtr++
			cbID := val.removeCallbackCtr
			vrc := &valueRemoveCallback{id: cbID, cb: cb}
			val.removeCallbacks = append(val.removeCallbacks, vrc)
			relLocked := func() {
				if !vrc.released.Swap(true) {
					for k := 0; k < len(val.removeCallbacks); k++ {
						if val.removeCallbacks[k].id == cbID {
							val.removeCallbacks = append(val.removeCallbacks[:k], val.removeCallbacks[k+1:]...)
							break
						}
					}
				}
			}
			return val.val, func() {
				i.c.mtx.Lock()
				defer i.c.mtx.Unlock()
				relLocked()
			}, true
		}
	}

	// not found
	return nil, nil, false
}

// addResolverRemovedCallbackLocked adds a callback to be called when the given resolver is removed.
// returns nil, nil, false if the resolver was already removed
func (i *directiveInstance) addResolverRemovedCallbackLocked(res *resolver, cb func()) (func(), bool) {
	if !slices.Contains(i.res, res) {
		return nil, false
	}

	rel := newCallback(cb)
	res.rels = append(res.rels, rel)
	relLocked := func() {
		if !rel.released.Swap(true) {
			for k := 0; k < len(res.rels); k++ {
				if res.rels[k] == rel {
					res.rels = append(res.rels[:k], res.rels[k+1:]...)
					break
				}
			}
		}
	}
	return func() {
		i.c.mtx.Lock()
		defer i.c.mtx.Unlock()
		relLocked()
	}, true
}

// attachStartSubResolverLocked adds and starts a sub-resolver for a parent resolver.
// returns nil, nil, false if the resolver was already removed
func (i *directiveInstance) attachStartSubResolverLocked(res *resolver, subRes directive.Resolver, removedCb func()) (func(), bool) {
	if !slices.Contains(i.res, res) || res.hnd.rel.Load() || i.released.Load() {
		return nil, false
	}

	subResReg := newResolver(i, res.hnd, subRes)
	if removedCb != nil {
		subResReg.rels = append(subResReg.rels, newCallback(removedCb))
	}
	i.attachStartResolverLocked(subResReg)

	return func() {
		i.c.mtx.Lock()
		defer i.c.mtx.Unlock()
		i.removeResolverLocked(-1, subResReg)
	}, true
}

// onValuesRemovedLocked is called after removing values from a resolver.
func (i *directiveInstance) onValuesRemovedLocked(vals ...*value) {
	if len(vals) == 0 {
		return
	}

	defer i.deferCheckStateChanged()()
	i.stateChangedSnapshot = directiveStateSnapshot{} // mark state as changed

	var cbs []func()
	for _, val := range vals {
		for _, removedCallback := range val.removeCallbacks {
			if !removedCallback.released.Swap(true) {
				cbs = append(cbs, removedCallback.cb)
			}
		}
		val.removeCallbacks = nil

		for _, ref := range i.refs {
			if !ref.released.Load() && ref.h != nil {
				cbs = append(cbs, func() {
					ref.h.HandleValueRemoved(i, val)
				})
			}
		}
	}
	i.callCallbacksLocked(cbs...)
	i.checkFullLocked()
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

// AddIdleCallback adds a callback that will be called when the idle state changes.
// Called immediately with the initial state.
// Returns a callback release function.
func (i *directiveInstance) AddIdleCallback(cb directive.IdleCallback) func() {
	i.c.mtx.Lock()
	defer i.c.mtx.Unlock()

	cbCtr := newCallback(cb)
	i.idles = append(i.idles, cbCtr)
	isIdle := i.ready && i.idle
	errs := i.getResolverErrsLocked()
	i.callCallbacksLocked(func() {
		cb(isIdle, errs)
	})

	return func() {
		if !cbCtr.released.Swap(true) {
			i.c.mtx.Lock()
			i.removeIdleCallbackLocked(cbCtr)
			i.c.mtx.Unlock()
		}
	}
}

// AddStateCallback adds a callback that will be called when the directive state changes.
// Called when any of the parameters of the callback change.
// This is more expensive than AddIdleCallback or AddDisposeCallback.
// Returns a callback release function.
func (i *directiveInstance) AddStateCallback(cb directive.StateCallback) func() {
	i.c.mtx.Lock()
	defer i.c.mtx.Unlock()

	cbCtr := newCallback(cb)
	i.stateChanged = append(i.stateChanged, cbCtr)
	stateSnapshot := i.getDirectiveStateSnapshotLocked()

	i.callCallbacksLocked(func() {
		cb(stateSnapshot.idle, stateSnapshot.errs, stateSnapshot.attachedVals)
	})

	return func() {
		if !cbCtr.released.Swap(true) {
			i.c.mtx.Lock()
			i.removeStateCallbackLocked(cbCtr)
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
func (i *directiveInstance) callHandlerUnlocked(handler *handler) (res []*resolver, err error) {
	defer func() {
		if rerr := recover(); rerr != nil {
			perr := handlePanic(i.logger(), rerr)
			if err == nil {
				err = perr
			}
		}
	}()

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
	if i.full {
		// already full => already idle => don't call handleIdleStateLocked.
		res.idle = true
		res.updateContextLocked(nil)
	} else {
		// we possibly just transitioned from idle => not idle.
		res.updateContextLocked(&i.ctx)
		i.handleIdleStateLocked()
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
		diIdx = slices.Index(i.c.dir, i)
		if diIdx < 0 {
			// not found
			return
		}
	}

	// remove from list of instances
	i.logger().Debug("removed directive")
	i.c.dir = append(i.c.dir[:diIdx], i.c.dir[diIdx+1:]...)

	// Clear all idle callbacks
	for _, idle := range i.idles {
		idle.released.Store(true)
	}
	i.idles = nil

	// Remove all resolvers and all values emitted by those resolvers.
	for len(i.res) != 0 {
		resIdx := len(i.res) - 1
		i.removeResolverLocked(resIdx, i.res[resIdx])
	}
	for _, res := range i.res {
		if res.ctxCancel != nil {
			res.ctxCancel()
		}
		res.ctx, res.ctxCancel = nil, nil
	}
	i.res = nil

	// Call directive release callbacks and ref release callbacks
	var cbs []func()
	for _, cb := range i.rels {
		if !cb.released.Swap(true) && cb.cb != nil {
			cbs = append(cbs, cb.cb)
		}
	}
	for _, ref := range i.refs {
		ref := ref
		if !ref.released.Swap(true) && ref.h != nil {
			cbs = append(cbs, func() {
				ref.h.HandleInstanceDisposed(i)
			})
		}
	}
	i.rels = nil
	i.callCallbacksLocked(cbs...)
	i.refs = nil
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
		resIdx = slices.Index(i.res, rres)
		if resIdx < 0 {
			return
		}
	}

	// remove the resolver from the list
	i.res = append(i.res[:resIdx], i.res[resIdx+1:]...)
	// cancel the resolver
	rres.updateContextLocked(nil)
	// remove values associated with the resolver
	vals := rres.vals
	rres.vals = nil
	i.onValuesRemovedLocked(vals...)
	// call the resolver removed callbacks
	var cbs []func()
	for _, cb := range rres.rels {
		if !cb.released.Swap(true) && cb.cb != nil {
			cbs = append(cbs, cb.cb)
		}
	}
	rres.rels = nil
	i.callCallbacksLocked(cbs...)
	// check if the idle state changed
	i.handleIdleStateLocked()
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
			func(cb func()) {
				defer func() {
					if rerr := recover(); rerr != nil {
						_ = handlePanic(i.logger(), rerr)
					}
				}()
				cb()
			}(cb)
		}
		i.c.mtx.Lock()
		cbs = i.refCbs
		i.refCbs = nil
	}
	i.callingRefCbs = false
}

// handlePanic converts a panic error into a regular error
func handlePanic(le *logrus.Entry, panicErr any) error {
	debug.PrintStack()
	e, eOk := panicErr.(error)
	if !eOk {
		e = errors.Errorf("%v", panicErr)
	}
	le.WithError(e).Error("callback panic")
	return e
}

// logger returns the logger for this instance
// no locks required
func (i *directiveInstance) logger() *logrus.Entry {
	return i.c.le.WithField("directive", i.GetDirectiveIdent())
}

// _ is a type assertion
var _ directive.Instance = ((*directiveInstance)(nil))
