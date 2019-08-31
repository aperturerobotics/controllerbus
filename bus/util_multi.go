package bus

import (
	"context"
	"github.com/aperturerobotics/controllerbus/directive"
)

// ExecCollectValues collects one or more values while ctx is active and
// directive is not idle.
//
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
	defer di.AddIdleCallback(subCtxCancel)()

	var vals []directive.Value
	for {
		select {
		case <-subCtx.Done():
			return vals, ref, nil
		case n := <-valCh:
			vals = append(vals, n)
		}
	}
}
