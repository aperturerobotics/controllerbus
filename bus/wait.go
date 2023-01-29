package bus

import (
	"context"

	"github.com/aperturerobotics/controllerbus/directive"
)

// ExecWaitValue executes a directive and waits for a value matching the cb.
//
// valDisposeCallback can be nil, will be called if the value is disposed.
// If returnIfIdle is set, returns nil, nil, nil if idle.
// If cb returns true, nil, returns the value.
// If checkCb is nil, returns first value.
func ExecWaitValue[T directive.Value](
	ctx context.Context,
	b Bus,
	dir directive.Directive,
	returnIfIdle bool,
	valDisposeCallback func(),
	checkCb func(val T) (bool, error),
) (T, directive.Instance, directive.Reference, error) {
	av, avDi, avRef, err := ExecOneOffWithFilter(ctx, b, dir, returnIfIdle, valDisposeCallback, func(val directive.AttachedValue) (bool, error) {
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
