package bus

import (
	"context"

	"github.com/aperturerobotics/controllerbus/directive"
)

// ExecWaitValue executes a directive and waits for a value matching the cb.
//
// valDisposeCallback can be nil, will be called if the value is disposed.
// If cb returns true, nil, returns the value.
// If checkCb is nil, returns first value.
//
// idleCb is called when idle with the list of resolver errors.
// idleCb should return (wait, error): if wait=true, continues to wait.
// if idleCb is nil: continues to wait when the directive becomes idle
// errs is the list of errors from the resolvers (if any)
func ExecWaitValue[T directive.Value](
	ctx context.Context,
	b Bus,
	dir directive.Directive,
	idleCb func(errs []error) (bool, error),
	valDisposeCallback func(),
	checkCb func(val T) (bool, error),
) (T, directive.Instance, directive.Reference, error) {
	av, avDi, avRef, err := ExecOneOffWithFilter(ctx, b, dir, idleCb, valDisposeCallback, func(val directive.AttachedValue) (bool, error) {
		v, vOk := val.GetValue().(T)
		if !vOk {
			return false, nil
		}
		if checkCb == nil {
			return true, nil
		}
		return checkCb(v)
	})
	if err != nil || av == nil {
		if avRef != nil {
			avRef.Release()
		}
		var empty T
		return empty, nil, nil, err
	}

	// note: type is already asserted above
	return av.GetValue().(T), avDi, avRef, nil
}
