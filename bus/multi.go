package bus

import (
	"context"
	"sync"
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

	// mtx, bcast guard these variables
	var mtx sync.Mutex
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
				mtx.Lock()
				idle = false
				if !returned {
					vals = append(vals, val)
					valIDs = append(valIDs, v.GetValueID())
					bcast.Broadcast()
				}
				mtx.Unlock()
			},
			func(v directive.AttachedValue) {
				_, valOk := v.GetValue().(T)
				if !valOk {
					return
				}
				mtx.Lock()
				idle = false
				id := v.GetValueID()
				for i, valID := range valIDs {
					if valID == id {
						if returned {
							defer valDisposeCallback()
						} else {
							valIDs[i] = valIDs[len(valIDs)-1]
							valIDs = valIDs[:len(valIDs)-1]
							vals[i] = vals[len(vals)-1]
							var empty T
							vals[len(vals)-1] = empty
							vals = vals[:len(vals)-1]
						}
						break
					}
				}
				mtx.Unlock()
			},
			func() {
				mtx.Lock()
				if returned {
					defer valDisposeCallback()
				} else if resErr == nil {
					resErr = ErrDirectiveDisposed
					bcast.Broadcast()
				}
				mtx.Unlock()
			},
		),
	)
	if err != nil {
		if ref != nil {
			ref.Release()
		}
		return nil, nil, nil, err
	}

	defer di.AddIdleCallback(func(errs []error) {
		mtx.Lock()
		defer mtx.Unlock()
		if resErr != nil {
			return
		}
		for _, err := range errs {
			if err != nil {
				resErr = err
				break
			}
		}
		// idle
		idle = true
		bcast.Broadcast()
	})()

	for {
		mtx.Lock()
		if resErr != nil {
			vals, resErr := vals, resErr // copy
			returned = true
			mtx.Unlock()
			ref.Release()
			return vals, di, nil, resErr
		}
		if idle && (!waitOne || len(vals) != 0) {
			vals, ref := vals, ref // copy
			returned = true
			mtx.Unlock()
			return vals, di, ref, nil
		}
		waitCh := bcast.GetWaitCh()
		mtx.Unlock()

		select {
		case <-ctx.Done():
			ref.Release()
			return nil, nil, nil, context.Canceled
		case <-waitCh:
		}
	}
}
