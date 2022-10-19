package keyed

import (
	"context"
	"time"
)

// runningRoutine tracks a running routine
type runningRoutine[T comparable] struct {
	// k is the keyed instance
	k *Keyed[T]
	// key is the key for this routine
	key string

	// fields guarded by k.mtx
	// ctx is the context
	ctx context.Context
	// ctxCancel cancels the context
	// if nil, not running
	ctxCancel context.CancelFunc
	// routine is the routine callback
	routine Routine
	// data is the associated routine data
	data T
	// err is the error if any
	err error
	// success indicates the routine succeeded
	success bool
	// deferRemove is set if we are waiting to remove this.
	deferRemove *time.Timer
}

// newRunningRoutine constructs a new runningRoutine
func newRunningRoutine[T comparable](k *Keyed[T], key string, routine Routine, data T) *runningRoutine[T] {
	return &runningRoutine[T]{
		k:       k,
		key:     key,
		routine: routine,
		data:    data,
	}
}

// start starts or restarts the routine.
// expects k.mtx to be locked by caller
func (r *runningRoutine[T]) start(ctx context.Context) {
	if r.success || r.routine == nil {
		return
	}
	if r.ctx != nil {
		select {
		case <-r.ctx.Done():
		default:
			// routine is still running
			return
		}
	}
	if r.err != nil {
		r.err = nil
	}
	if r.ctxCancel != nil {
		r.ctxCancel()
		r.ctxCancel = nil
		r.ctx = nil
	}
	r.ctx, r.ctxCancel = context.WithCancel(ctx)
	go r.execute(r.ctx, r.ctxCancel)
}

// execute executes the routine.
func (r *runningRoutine[T]) execute(ctx context.Context, cancel context.CancelFunc) {
	err := r.routine(ctx)
	cancel()

	r.k.mtx.Lock()
	if r.ctx == ctx {
		r.err = err
		r.success = err == nil
		r.ctxCancel = nil
		r.ctx = nil
		for i := len(r.k.exitedCbs) - 1; i >= 0; i-- {
			// run after unlocking mtx
			defer (r.k.exitedCbs[i])(r.key, r.routine, r.data, r.err)
		}
	}
	r.k.mtx.Unlock()
}

// remove is called when the routine is removed / canceled.
// expects r.k.mtx to be locked
func (r *runningRoutine[T]) remove() {
	if r.deferRemove != nil {
		return
	}
	removeNow := func() {
		if r.ctxCancel != nil {
			r.ctxCancel()
		}
		delete(r.k.routines, r.key)
	}
	if r.k.releaseDelay == 0 {
		removeNow()
		return
	}

	timerCb := func() {
		r.k.mtx.Lock()
		if r.k.routines[r.key] == r && r.deferRemove != nil {
			_ = r.deferRemove.Stop()
			r.deferRemove = nil
			removeNow()
		}
		r.k.mtx.Unlock()
	}
	r.deferRemove = time.AfterFunc(r.k.releaseDelay, timerCb)
}
