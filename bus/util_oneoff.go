package bus

import (
	"context"
	"github.com/aperturerobotics/controllerbus/directive"
)

// ExecOneOff executes a one-off directive.
// If returnIfIdle is set, returns nil, nil, nil if idle.
func ExecOneOff(
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
	returnIfIdle bool,
	valDisposeCallback func(),
) (directive.AttachedValue, directive.Reference, error) {
	valCh := make(chan directive.AttachedValue, 1)
	di, ref, err := bus.AddDirective(
		dir,
		&CallbackHandler{
			disposeCb: valDisposeCallback,
			valCb: func(v directive.AttachedValue) {
				select {
				case valCh <- v:
				default:
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

	errCh := make(chan error, 1)
	defer di.AddIdleCallback(func(errs []error) {
		var err error
		if len(errs) != 0 {
			for _, rerr := range errs {
				if rerr != nil {
					err = rerr
					break
				}
			}
		}
		if !returnIfIdle && err == nil {
			return
		}
		select {
		case errCh <- err:
		default:
		}
	})()

	select {
	case <-ctx.Done():
		if ref != nil {
			ref.Release()
		}
		return nil, nil, context.Canceled
	case n := <-valCh:
		return n, ref, nil
	case err := <-errCh:
		if ref != nil {
			ref.Release()
		}
		return nil, nil, err
	}
}
