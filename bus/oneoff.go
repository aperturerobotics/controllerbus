package bus

import (
	"context"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/broadcast"
)

// ExecIdleCallback is an idle callback for ExecOneOffWithFilter.
//
// idleCb is called with the idle state and list of resolver errors.
// idleCb should return (wait, error): if wait=true, continues to wait.
// if idleCb is nil: continues to wait when the directive becomes idle
// errs is the list of errors from the resolvers (if any)
type ExecIdleCallback func(isIdle bool, errs []error) (bool, error)

// WaitWhenIdle returns an ExecIdleCallback that waits when the directive becomes idle.
//
// if ignoreErrors is false, returns any non-nil resolver error that occurs.
func WaitWhenIdle(ignoreErrors bool) ExecIdleCallback {
	return func(isIdle bool, errs []error) (bool, error) {
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
	return func(isIdle bool, errs []error) (bool, error) {
		if !isIdle {
			return true, nil
		}
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
// if idleCb is nil: if the directive becomes idle & any resolvers failed, returns the resolver error.
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
	var bcast broadcast.Broadcast

	var val directive.AttachedValue
	var resErr error
	var idle, wait bool

	if idleCb == nil {
		// register a callback to catch any resolver errors
		idleCb = WaitWhenIdle(false)
	}

	di, ref, err := bus.AddDirective(
		dir,
		NewCallbackHandler(
			func(v directive.AttachedValue) {
				bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
					if val == nil && resErr == nil {
						ok := filterCb == nil
						if !ok {
							ok, resErr = filterCb(v)
							if resErr != nil {
								broadcast()
							}
						}
						if ok && resErr == nil {
							val = v
							broadcast()
						}
					}
				})
			},
			func(v directive.AttachedValue) {
				if valDisposeCallback == nil {
					return
				}
				var valueRemoved bool
				bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
					valueRemoved = val != nil && val.GetValueID() == v.GetValueID()
				})
				if valueRemoved {
					valDisposeCallback()
				}
			},
			func() {
				bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
					if resErr == nil {
						resErr = ErrDirectiveDisposed
						broadcast()
					}
				})
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

	defer di.AddIdleCallback(func(isIdle bool, errs []error) {
		bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
			// idle or errors list changed
			idle = isIdle
			wait, resErr = idleCb(idle, errs)
			broadcast()
		})
	})()

	for {
		var currVal directive.AttachedValue
		var currResErr error
		var currIdle, currWait bool
		var waitCh <-chan struct{}
		bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
			waitCh = getWaitCh()
			currVal, currResErr, currIdle, currWait = val, resErr, idle, wait
		})
		if currVal != nil {
			return currVal, di, ref, nil
		}
		if currResErr != nil || (currIdle && !currWait) {
			ref.Release()
			return nil, nil, nil, currResErr
		}

		select {
		case <-ctx.Done():
			ref.Release()
			return nil, nil, nil, context.Canceled
		case <-waitCh:
		}
	}
}

// ExecOneOffTyped executes a one-off directive with a value type.
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
func ExecOneOffTyped[T comparable](
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
	idleCb ExecIdleCallback,
	valDisposeCallback func(),
) (directive.TypedAttachedValue[T], directive.Instance, directive.Reference, error) {
	return ExecOneOffWithFilterTyped[T](ctx, bus, dir, idleCb, valDisposeCallback, nil)
}

// ExecOneOffWithFilterTyped executes a one-off directive with a filter cb.
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
func ExecOneOffWithFilterTyped[T comparable](
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
	idleCb ExecIdleCallback,
	valDisposeCallback func(),
	filterCb func(val directive.TypedAttachedValue[T]) (bool, error),
) (directive.TypedAttachedValue[T], directive.Instance, directive.Reference, error) {
	av, di, ref, err := ExecOneOffWithFilter(
		ctx,
		bus,
		dir,
		idleCb,
		valDisposeCallback,
		func(val directive.AttachedValue) (bool, error) {
			tval, ok := val.GetValue().(T)
			if !ok {
				return false, nil
			}
			if filterCb != nil {
				return filterCb(directive.NewTypedAttachedValue(val.GetValueID(), tval))
			}
			return true, nil
		},
	)
	if err != nil || ref == nil {
		return nil, di, ref, err
	}

	// we type asserted in filterCb above.
	avVal := av.GetValue().(T)
	return directive.NewTypedAttachedValue(av.GetValueID(), avVal), di, ref, nil
}

// ExecOneOffWithXfrm executes a one-off directive with a transformation filter cb.
//
// Waits until the callback returns true and nil err before returning.
// valDisposeCb is called if the value is no longer valid.
// valDisposeCb might be called multiple times.
//
// idleCb is called when idle with the list of resolver errors.
// idleCb should return (wait, error): if wait=true, continues to wait.
// if idleCb is nil: continues to wait when the directive becomes idle
// errs is the list of errors from the resolvers (if any)
//
// If err != nil, ref == nil.
func ExecOneOffWithXfrm(
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
	idleCb ExecIdleCallback,
	valDisposeCallback func(),
	xfrmCb func(val directive.AttachedValue) (directive.Value, bool, error),
) (directive.Value, directive.AttachedValue, directive.Instance, directive.Reference, error) {
	var xfrm directive.Value
	av, di, ref, err := ExecOneOffWithFilter(
		ctx,
		bus,
		dir,
		idleCb,
		valDisposeCallback,
		func(val directive.AttachedValue) (bool, error) {
			var ok bool
			xval, ok, err := xfrmCb(val)
			if !ok || err != nil {
				return false, err
			}
			xfrm = xval
			return true, nil
		},
	)
	return xfrm, av, di, ref, err
}

// ExecOneOffWithXfrmTyped executes a one-off directive with a transformation filter cb.
//
// Waits until the callback returns true and nil err before returning.
// valDisposeCb is called if the value is no longer valid.
// valDisposeCb might be called multiple times.
//
// idleCb is called when idle with the list of resolver errors.
// idleCb should return (wait, error): if wait=true, continues to wait.
// if idleCb is nil: continues to wait when the directive becomes idle
// errs is the list of errors from the resolvers (if any)
//
// If err != nil, ref == nil.
func ExecOneOffWithXfrmTyped[T, R directive.ComparableValue](
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
	idleCb ExecIdleCallback,
	valDisposeCallback func(),
	xfrmCb func(val directive.TypedAttachedValue[T]) (R, bool, error),
) (R, directive.TypedAttachedValue[T], directive.Instance, directive.Reference, error) {
	var xfrm R
	av, di, ref, err := ExecOneOffWithFilterTyped(
		ctx,
		bus,
		dir,
		idleCb,
		valDisposeCallback,
		func(val directive.TypedAttachedValue[T]) (bool, error) {
			var ok bool
			xval, ok, err := xfrmCb(val)
			if !ok || err != nil {
				return false, err
			}
			xfrm = xval
			return true, nil
		},
	)
	if err != nil || av == nil {
		// ensure xfrm is empty too
		var empty R
		xfrm = empty
	}
	return xfrm, av, di, ref, err
}
