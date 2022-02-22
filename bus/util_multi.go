package bus

import (
	"context"

	"github.com/aperturerobotics/controllerbus/directive"
)

// ExecCollectValues collects one or more values while ctx is active and
// directive is not idle.
//
// Returns err if the directive is idle and a resolver returned an error.
// Does not return an error if ctx was canceled.
// Does not release ref if err == nil.
func ExecCollectValues(
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
	valDisposeCallback func(),
) ([]directive.Value, directive.Reference, error) {
	subCtx, subCtxCancel := context.WithCancel(ctx)
	defer subCtxCancel()

	valCh := make(chan directive.Value, 1)
	di, ref, err := bus.AddDirective(
		dir,
		&CallbackHandler{
			disposeCb: func() {
				subCtxCancel()
				if valDisposeCallback != nil {
					valDisposeCallback()
				}
			},
			valCb: func(v directive.AttachedValue) {
				select {
				case <-subCtx.Done():
				case valCh <- v.GetValue():
				}
			},
		},
	)
	if err != nil {
		if ref != nil {
			ref.Release()
		}
		return nil, nil, err
	}

	errCh := make(chan error, 2)
	defer di.AddIdleCallback(func(errs []error) {
		if len(errs) != 0 {
			for _, err := range errs {
				if err == nil {
					continue
				}

				select {
				case errCh <- err:
				default:
					return
				}
			}
		} else {
			// idle
			select {
			case <-subCtx.Done():
			case valCh <- nil:
			}
		}
	})()

	var vals []directive.Value
	for {
		select {
		case <-ctx.Done():
			if ref != nil {
				ref.Release()
			}
			return vals, nil, context.Canceled
		case <-subCtx.Done():
			return vals, ref, nil
		case err := <-errCh:
			if ref != nil {
				ref.Release()
			}
			return nil, nil, err
		case n := <-valCh:
			if n == nil {
				return vals, ref, nil
			}
			vals = append(vals, n)
		}
	}
}
