package configset_controller

import (
	"github.com/aperturerobotics/controllerbus/controller/configset"
	"sync"
)

// runningControllerRef implements the reference directive.
type runningControllerRef struct {
	id                  string
	relInternalCallback func()

	mtx sync.Mutex
	cbs []func(configset.State)
	rc  *runningController // may be nil

	relOnce sync.Once
}

func newRunningControllerRef(id string, rc *runningController) *runningControllerRef {
	return &runningControllerRef{id: id, rc: rc}
}

// GetConfigKey returns the configset key for this controller.
func (r *runningControllerRef) GetConfigKey() string {
	return r.id
}

// GetState returns the current state object.
func (r *runningControllerRef) GetState() configset.State {
	r.mtx.Lock()
	rc := r.rc
	r.mtx.Unlock()
	if rc == nil {
		return &runningControllerState{
			id: r.id,
		}
	}

	rc.mtx.Lock()
	st := rc.state
	rc.mtx.Unlock()
	return &st
}

// AddStateCb adds a callback that is called when the state changes.
// Should not block.
// Will be called with the initial state.
func (r *runningControllerRef) AddStateCb(cb func(configset.State)) {
	r.mtx.Lock()
	r.cbs = append(r.cbs, cb)
	r.mtx.Unlock()
	st := r.GetState()
	cb(st)
}

// GetRunningController gets the running controller.
func (r *runningControllerRef) GetRunningController() *runningController {
	r.mtx.Lock()
	rc := r.rc
	r.mtx.Unlock()
	return rc
}

// setRunningController updates the running controller.
// use ApplyReference
func (r *runningControllerRef) setRunningController(rc *runningController, st configset.State) bool {
	var updated bool
	r.mtx.Lock()
	if r.rc != rc {
		r.rc = rc
		updated = true
		for _, cb := range r.cbs { // r.rc.pushState(st)
			cb(st)
		}
	}
	r.mtx.Unlock()
	return updated
}

// pushState pushes an updated state
func (r *runningControllerRef) pushState(st configset.State) {
	r.mtx.Lock()
	for _, cb := range r.cbs {
		cb(st)
	}
	r.mtx.Unlock()
}

// Release releases the reference.
func (r *runningControllerRef) Release() {
	r.relOnce.Do(func() {
		if r.rc != nil {
			r.rc.DelReference(r)
		}
		if cb := r.relInternalCallback; cb != nil {
			go r.relInternalCallback()
		}
	})
}

// _ is a type assertion
var _ configset.Reference = ((*runningControllerRef)(nil))
