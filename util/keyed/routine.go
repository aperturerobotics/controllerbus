package keyed

import (
	"context"
)

// runningRoutine tracks a running routine
type runningRoutine struct {
	// k is the keyed instance
	k *Keyed

	// fields guarded by k.mtx
	// ctx is the context
	ctx context.Context
	// ctxCancel cancels the context
	// if nil, not running
	ctxCancel context.CancelFunc
	// routine is the routine callback
	routine Routine
	// err is the error if any
	err error
	// success indicates the routine succeeded
	success bool
}

// newRunningRoutine constructs a new runningRoutine
func newRunningRoutine(k *Keyed, routine Routine) *runningRoutine {
	return &runningRoutine{
		k:       k,
		routine: routine,
	}
}

// start starts or restarts the routine.
// expects k.mtx to be locked by caller
func (r *runningRoutine) start(ctx context.Context) {
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
func (r *runningRoutine) execute(ctx context.Context, cancel context.CancelFunc) {
	err := r.routine(ctx)
	cancel()

	r.k.mtx.Lock()
	if r.ctx == ctx {
		r.err = err
		r.success = err == nil
		r.ctxCancel = nil
		r.ctx = nil
	}
	r.k.mtx.Unlock()
}
