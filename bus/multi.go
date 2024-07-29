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
