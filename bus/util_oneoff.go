package bus

import (
	"context"
	"sync"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/broadcast"
)

// ExecOneOff executes a one-off directive.
//
// If returnIfIdle is set, returns nil, nil, nil if idle.
// If any resolvers return an error, returns that error.
// If err != nil, ref == nil.
func ExecOneOff(
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
	returnIfIdle bool,
	valDisposeCallback func(),
) (directive.AttachedValue, directive.Reference, error) {
	return ExecOneOffWithFilter(ctx, bus, dir, returnIfIdle, valDisposeCallback, nil)
}

// ExecOneOffWithFilter executes a one-off directive with a filter cb.
//
// Waits until the callback returns true before returning a value.
// If returnIfIdle is set, returns nil, nil, nil if idle.
// If any resolvers return an error, returns that error.
// If err != nil, ref == nil.
func ExecOneOffWithFilter(
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
	returnIfIdle bool,
	valDisposeCallback func(),
	filterCb func(val directive.AttachedValue) (bool, error),
) (directive.AttachedValue, directive.Reference, error) {
	// mtx, bcast guard these variables
	var mtx sync.Mutex
	var bcast broadcast.Broadcast

	var val directive.AttachedValue
	var resErr error
	var idle bool

	di, ref, err := bus.AddDirective(
		dir,
		NewCallbackHandler(
			func(v directive.AttachedValue) {
				mtx.Lock()
				if val == nil && resErr == nil {
					ok := filterCb == nil
					if !ok {
						ok, resErr = filterCb(v)
						if resErr != nil {
							bcast.Broadcast()
						}
					}
					if ok && resErr == nil {
						val = v
						bcast.Broadcast()
					}
				}
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
		if resErr == nil {
			idle = true
		}
		bcast.Broadcast()
	})()

	for {
		mtx.Lock()
		if val != nil {
			mtx.Unlock()
			return val, ref, nil
		}
		if resErr != nil || (idle && returnIfIdle) {
			mtx.Unlock()
			ref.Release()
			return nil, nil, resErr
		}
		mtx.Unlock()

		select {
		case <-ctx.Done():
			ref.Release()
			return nil, nil, context.Canceled
		case <-bcast.GetWaitCh():
		}
	}
}
