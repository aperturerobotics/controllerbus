package bus

import (
	"context"
	"sync"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/broadcast"
)

// ExecCollectValues collects one or more values while ctx is active and
// directive is not idle.
//
// Returns vals, ref, nil if the directive is idle.
// Returns err if any resolver returns an error.
// If err != nil, ref == nil.
func ExecCollectValues(
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
	valDisposeCallback func(),
) ([]directive.Value, directive.Reference, error) {
	// mtx, bcast guard these variables
	var mtx sync.Mutex
	var bcast broadcast.Broadcast

	var vals []directive.Value
	var resErr error
	var idle bool

	di, ref, err := bus.AddDirective(
		dir,
		NewCallbackHandler(
			func(v directive.AttachedValue) {
				mtx.Lock()
				vals = append(vals, v.GetValue())
				bcast.Broadcast()
				mtx.Unlock()
			},
			nil,
			func() {
				mtx.Lock()
				if !idle {
					idle = true
					bcast.Broadcast()
				}
				mtx.Unlock()
				if valDisposeCallback != nil {
					go valDisposeCallback()
				}
			},
		),
	)
	if err != nil {
		if ref != nil {
			ref.Release()
		}
		return nil, nil, err
	}

	defer di.AddIdleCallback(func(errs []error) {
		mtx.Lock()
		defer mtx.Unlock()
		if resErr != nil || idle {
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
			mtx.Unlock()
			ref.Release()
			return vals, nil, resErr
		}
		if idle {
			mtx.Unlock()
			return vals, ref, nil
		}
		waitCh := bcast.GetWaitCh()
		mtx.Unlock()

		select {
		case <-ctx.Done():
			ref.Release()
			return nil, nil, context.Canceled
		case <-waitCh:
		}
	}
}
