package bus

import (
	"context"
	"sync"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/broadcast"
)

// ExecIdleCallback is an idle callback for ExecOneOffWithFilter.
//
// idleCb is called when idle with the list of resolver errors.
// idleCb should return (wait, error): if wait=true, continues to wait.
// if idleCb is nil: continues to wait when the directive becomes idle
// errs is the list of errors from the resolvers (if any)
type ExecIdleCallback func(errs []error) (bool, error)

// WaitWhenIdle returns an ExecIdleCallback that waits when the directive becomes idle.
//
// if ignoreErrors is false, returns any non-nil resolver error that occurs.
func WaitWhenIdle(ignoreErrors bool) ExecIdleCallback {
	return func(errs []error) (bool, error) {
		if !ignoreErrors {
			for _, err := range errs {
				if err != nil {
					return false, err
				}
			}
		}
		return true, nil
	}
}

// ReturnWhenIdle returns an ExecIdleCallback that returns when the directive becomes idle.
func ReturnWhenIdle() ExecIdleCallback {
	return func(errs []error) (bool, error) {
		for _, err := range errs {
			if err != nil {
				return false, err
			}
		}
		return false, nil
	}
}

// ReturnIfIdle returns an ExecIdleCallback that returns when the directive becomes idle if true.
func ReturnIfIdle(returnIfIdle bool) ExecIdleCallback {
	if returnIfIdle {
		return ReturnWhenIdle()
	}
	return WaitWhenIdle(false)
}

// ExecOneOff executes a one-off directive.
//
// idleCb is called when idle with the list of resolver errors.
// idleCb should return (wait, error): if wait=true, continues to wait.
// if idleCb is nil: continues to wait when the directive becomes idle
// errs is the list of errors from the resolvers (if any)
//
// If any resolvers return an error, returns that error.
// valDisposeCb is called if the value is no longer valid.
// valDisposeCb might be called multiple times.
// If err != nil, ref == nil.
func ExecOneOff(
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
	idleCb ExecIdleCallback,
	valDisposeCallback func(),
) (directive.AttachedValue, directive.Instance, directive.Reference, error) {
	return ExecOneOffWithFilter(ctx, bus, dir, idleCb, valDisposeCallback, nil)
}

// ExecOneOffWithFilter executes a one-off directive with a filter cb.
//
// Waits until the callback returns true before returning a value.
// valDisposeCb is called if the value is no longer valid.
// valDisposeCb might be called multiple times.
//
// idleCb is called when idle with the list of resolver errors.
// idleCb should return (wait, error): if wait=true, continues to wait.
// if idleCb is nil: continues to wait when the directive becomes idle
// errs is the list of errors from the resolvers (if any)
//
// If err != nil, ref == nil.
func ExecOneOffWithFilter(
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
	idleCb ExecIdleCallback,
	valDisposeCallback func(),
	filterCb func(val directive.AttachedValue) (bool, error),
) (directive.AttachedValue, directive.Instance, directive.Reference, error) {
	// mtx, bcast guard these variables
	var mtx sync.Mutex
	var bcast broadcast.Broadcast

	var val directive.AttachedValue
	var resErr error
	var idle, wait bool

	if idleCb == nil {
		idleCb = WaitWhenIdle(false)
	}

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
			func(v directive.AttachedValue) {
				if valDisposeCallback == nil {
					return
				}
				mtx.Lock()
				valueRemoved := val != nil && val.GetValueID() == v.GetValueID()
				mtx.Unlock()
				if valueRemoved {
					valDisposeCallback()
				}
			},
			func() {
				mtx.Lock()
				if resErr != nil && !idle {
					resErr = ErrDirectiveDisposed
					bcast.Broadcast()
				}
				mtx.Unlock()
				if valDisposeCallback != nil {
					valDisposeCallback()
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

	defer di.AddIdleCallback(func(errs []error) {
		mtx.Lock()
		defer mtx.Unlock()
		if resErr != nil {
			return
		}
		// idle
		idle = true
		wait, resErr = idleCb(errs)
		bcast.Broadcast()
	})()

	for {
		mtx.Lock()
		if val != nil {
			val, ref := val, ref // copy
			mtx.Unlock()
			return val, di, ref, nil
		}
		if resErr != nil || (idle && !wait) {
			err, ref := resErr, ref
			mtx.Unlock()
			ref.Release()
			return nil, nil, nil, err
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
