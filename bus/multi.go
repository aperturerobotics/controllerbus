package bus

import (
	"context"
	"slices"
	"sync/atomic"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/broadcast"
)

// ExecCollectValues collects one or more values while ctx is active and
// directive is not idle.
//
// Returns vals, ref, nil if the directive is idle.
// Returns err if any resolver returns an error.
// valDisposeCb is called if any of the values are no longer valid.
// valDisposeCb might be called multiple times.
// If err != nil, ref == nil.
// If waitOne=true, waits for at least one value before returning.
func ExecCollectValues[T directive.Value](
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
	waitOne bool,
	valDisposeCb func(),
) ([]T, directive.Instance, directive.Reference, error) {
	var disposed atomic.Bool
	valDisposeCallback := func() {
		if !disposed.Swap(true) {
			if valDisposeCb != nil {
				valDisposeCb()
			}
		}
	}

	// bcast guards these variables
	var bcast broadcast.Broadcast

	var vals []T
	var valIDs []uint32
	var resErr error
	var idle bool
	var returned bool

	di, ref, err := bus.AddDirective(
		dir,
		NewCallbackHandler(
			func(v directive.AttachedValue) {
				val, valOk := v.GetValue().(T)
				if !valOk {
					return
				}
				bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
					if !returned {
						// we can append without re-allocating.
						vals = append(vals, val)
						valIDs = append(valIDs, v.GetValueID())
						broadcast()
					}
				})
			},
			func(v directive.AttachedValue) {
				_, valOk := v.GetValue().(T)
				if !valOk {
					return
				}
				var cb func()
				bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
					id := v.GetValueID()
					for i, valID := range valIDs {
						if valID == id {
							if returned {
								cb = valDisposeCallback
							} else {
								// removing requires re-allocating.
								valIDs = slices.Clone(valIDs)
								valIDs[i] = valIDs[len(valIDs)-1]
								valIDs = valIDs[:len(valIDs)-1]
								vals = slices.Clone(vals)
								vals[i] = vals[len(vals)-1]
								var empty T
								vals[len(vals)-1] = empty
								vals = vals[:len(vals)-1]
								broadcast()
							}
							break
						}
					}
				})
				if cb != nil {
					cb()
				}
			},
			func() {
				var cb func()
				bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
					if returned {
						cb = valDisposeCallback
					} else if resErr == nil {
						resErr = directive.ErrDirectiveDisposed
						broadcast()
					}
				})
				if cb != nil {
					cb()
				}
			},
		),
	)
	if err != nil {
		if ref != nil {
			ref.Release()
		}
		return nil, nil, nil, err
	}

	defer di.AddIdleCallback(func(isIdle bool, errs []error) {
		bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
			// if the resolver already returned an error, do nothing.
			if resErr != nil {
				return
			}
			if idle != isIdle {
				idle = isIdle
				broadcast()
			}
			for _, err := range errs {
				if err != nil {
					resErr = err
					broadcast()
					break
				}
			}
		})
	})()

	for {
		var currVals []T
		var currResErr error
		var currIdle bool
		var currRef directive.Reference
		var waitCh <-chan struct{}
		var returnVals bool
		bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
			currVals, currResErr, currIdle, currRef = vals, resErr, idle, ref
			waitCh = getWaitCh()
			returnVals = currResErr == nil && currIdle && (!waitOne || len(currVals) != 0)
			returned = currResErr != nil || returnVals
		})
		if currResErr != nil {
			currRef.Release()
			return currVals, di, nil, currResErr
		}
		if returnVals {
			return currVals, di, currRef, nil
		}

		select {
		case <-ctx.Done():
			currRef.Release()
			return nil, nil, nil, context.Canceled
		case <-waitCh:
		}
	}
}

// ExecCollectValuesWatch collects values and calls a callback whenever the slice changes.
//
// The callback is called with a snapshot of the current values. Its behavior is
// determined by the waitIdle flag and the directive's state.
//
// Rules for Emitting:
//  1. Initial State: Before the first emit, if waitIdle is true, all adds and
//     removals will be buffered. The first emit occurs only when the directive
//     becomes idle. This first emit will happen even if the collected value set
//     is empty.
//  2. If waitIdle is false, any change (add, remove, error) triggers an immediate emit.
//  3. After the First Emit: Once the initial collection has been emitted, any
//     subsequent change (add, remove, error) will trigger an immediate emit,
//     regardless of the waitIdle flag.
//
// The callback runs in a separate goroutine. If the callback returns an error,
// it will not be called again, and the error will be passed to errorCb.
//
// errorCb can be nil.
//
// Returns the directive instance and reference for cleanup. If err != nil, ref == nil.
func ExecCollectValuesWatch[T directive.Value](
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
	waitIdle bool,
	callback func(resErr error, vals []T) error,
	errorCb func(err error),
) (directive.Instance, directive.Reference, error) {
	// bcast guards these variables
	var bcast broadcast.Broadcast

	var vals []T
	var valIDs []uint32
	var resErr error
	var idle bool
	var emittedOnce bool // Set to true after the first emit and never reset.
	var pendingEmit bool

	// scheduleEmitIfReady checks if an emit should be triggered now or deferred.
	scheduleEmitIfReady := func(broadcast func()) {
		// Emit immediately if we are not waiting for an initial idle, OR if the
		// initial emit has already happened.
		if !waitIdle || emittedOnce {
			pendingEmit = true
			broadcast()
		}
	}

	di, ref, err := bus.AddDirective(
		dir,
		NewCallbackHandler(
			func(v directive.AttachedValue) { // Add handler
				val, valOk := v.GetValue().(T)
				if !valOk {
					return
				}
				bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
					vals = append(vals, val)
					valIDs = append(valIDs, v.GetValueID())
					scheduleEmitIfReady(broadcast)
				})
			},
			func(v directive.AttachedValue) { // Remove handler
				_, valOk := v.GetValue().(T)
				if !valOk {
					return
				}
				bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
					id := v.GetValueID()
					for i, valID := range valIDs {
						if valID == id {
							// In-place removal for efficiency.
							valIDs[i] = valIDs[len(valIDs)-1]
							valIDs = valIDs[:len(valIDs)-1]
							vals[i] = vals[len(vals)-1]
							var empty T
							vals[len(vals)-1] = empty
							vals = vals[:len(vals)-1]
							scheduleEmitIfReady(broadcast)
							break
						}
					}
				})
			},
			func() { // Dispose handler
				bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
					if resErr == nil {
						resErr = directive.ErrDirectiveDisposed
						pendingEmit = true
						broadcast()
					}
				})
			},
		),
	)
	if err != nil {
		if ref != nil {
			ref.Release()
		}
		return nil, nil, err
	}

	defer di.AddIdleCallback(func(isIdle bool, errs []error) {
		bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
			if resErr != nil {
				return
			}

			wasIdle := idle
			idle = isIdle

			// Errors always trigger an emit.
			for _, err := range errs {
				if err != nil {
					resErr = err
					pendingEmit = true
					broadcast()
					return
				}
			}

			// If we just became idle and haven't emitted the initial set yet,
			// trigger the emit now. This is the core of the waitIdle logic.
			if !wasIdle && idle && !emittedOnce {
				pendingEmit = true
				broadcast()
			}
		})
	})()

	// Goroutine to handle callbacks
	go func() {
		for {
			var currVals []T
			var currErr error
			var shouldEmit bool
			var waitCh <-chan struct{}

			bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
				currErr = resErr
				shouldEmit = pendingEmit
				pendingEmit = false
				waitCh = getWaitCh()

				if shouldEmit {
					// Mark that the initial emit has happened.
					// This is the one and only time this flag is set.
					emittedOnce = true
					if currErr == nil {
						currVals = slices.Clone(vals)
					}
				}
			})

			if shouldEmit {
				if err := callback(currErr, currVals); err != nil {
					if errorCb != nil {
						errorCb(err)
					}
					return // Stop processing on callback error.
				}
			}

			select {
			case <-ctx.Done():
				return
			case <-waitCh:
				// Await next broadcast
			}
		}
	}()

	return di, ref, nil
}
