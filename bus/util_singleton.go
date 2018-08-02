package bus

import (
	"context"
	"github.com/aperturerobotics/controllerbus/directive"
)

// ExecSingleton executes a singleton directive.
func ExecSingleton(
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
) (directive.Value, error) {
	valCh := make(chan directive.Value, 1)
	_, ref, err := bus.AddDirective(
		dir,
		func(v directive.Value) {
			select {
			case valCh <- v:
			case <-ctx.Done():
			}
		},
	)
	if err != nil {
		return nil, err
	}
	defer ref.Release()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case n := <-valCh:
		return n, nil
	}
}
