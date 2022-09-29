package bus

import (
	"context"

	"github.com/aperturerobotics/controllerbus/directive"
)

// ExecWaitValue executes a directive and waits for a value matching the cb.
// If cb returns true, nil, returns the value.
// If checkCb is nil, returns first value.
func ExecWaitValue[T any](
	ctx context.Context,
	b Bus,
	dir directive.Directive,
	checkCb func(val T) (bool, error),
) (T, directive.Reference, error) {
	waitCtx, waitCtxCancel := context.WithCancel(ctx)
	defer waitCtxCancel()

	valCh := make(chan T, 1)
	pushVal := func(val T) {
		for {
			select {
			case valCh <- val:
				return
			default:
			}
			select {
			case <-valCh:
			default:
			}
		}
	}
	_, diRef, err := b.AddDirective(
		dir,
		NewCallbackHandler(func(av directive.AttachedValue) {
			val, valOk := av.GetValue().(T)
			if !valOk {
				return
			}
			pushVal(val)
		}, nil, waitCtxCancel),
	)
	if err != nil {
		var empty T
		return empty, nil, err
	}

	for {
		select {
		case <-waitCtx.Done():
			diRef.Release()
			var empty T
			return empty, nil, context.Canceled
		case val := <-valCh:
			if checkCb != nil {
				matched, err := checkCb(val)
				if err != nil {
					diRef.Release()
					var empty T
					return empty, nil, err
				}
				if !matched {
					continue
				}
			}
			return val, diRef, nil
		}
	}
}
