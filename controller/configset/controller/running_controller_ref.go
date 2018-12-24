package configset_controller

import (
	"github.com/aperturerobotics/controllerbus/controller/configset"
	"sync"
)

// runningControllerRef implements the reference directive.
type runningControllerRef struct {
	rc *runningController

	mtx sync.Mutex
	cbs []func(configset.State)

	relOnce sync.Once
}

func newRunningControllerRef(rc *runningController) *runningControllerRef {
	return &runningControllerRef{rc: rc}
}

// GetConfigKey returns the configset key for this controller.
func (r *runningControllerRef) GetConfigKey() string {
	return r.rc.GetConfigKey()
}

// GetState returns the current state object.
func (r *runningControllerRef) GetState() configset.State {
	r.rc.mtx.Lock()
	st := r.rc.state
	r.rc.mtx.Unlock()
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
		r.rc.DelReference(r)
	})
}

// _ is a type assertion
var _ configset.Reference = ((*runningControllerRef)(nil))
