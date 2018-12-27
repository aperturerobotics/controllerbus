package bus

import (
	"context"
	"github.com/aperturerobotics/controllerbus/directive"
)

// ExecOneOff executes a one-off directive.
// Returns nil if the directive is canceled for some reason during the execution.
func ExecOneOff(
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
	valDisposeCallback func(),
) (directive.AttachedValue, directive.Reference, error) {
	valCh := make(chan directive.AttachedValue, 1)
	_, ref, err := bus.AddDirective(
		dir,
		&CallbackHandler{
			disposeCb: valDisposeCallback,
			valCb: func(v directive.AttachedValue) {
				select {
				case valCh <- v:
				case <-ctx.Done():
				}
			},
		},
	)
	if err != nil {
		ref.Release()
		return nil, nil, err
	}

	select {
	case <-ctx.Done():
		ref.Release()
		return nil, nil, ctx.Err()
	case n := <-valCh:
		return n, ref, nil
	}
}
