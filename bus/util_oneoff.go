package bus

import (
	"context"
	"github.com/aperturerobotics/controllerbus/directive"
)

type oneoffHandler struct {
	valCb     func(directive.Value)
	disposeCb func()
}

// HandleValueAdded is called when a value is added to the directive.
func (h *oneoffHandler) HandleValueAdded(_ directive.Instance, v directive.Value) {
	h.valCb(v)
}

// HandleValueRemoved is called when a value is removed from the directive.
func (h *oneoffHandler) HandleValueRemoved(directive.Instance, directive.Value) {
	// noop
}

// HandleInstanceDisposed is called when a directive instance is disposed.
// This will occur if Close() is called on the directive instance.
func (h *oneoffHandler) HandleInstanceDisposed(directive.Instance) {
	if h.disposeCb != nil {
		h.disposeCb()
	}
}

// _ is a type assertion
var _ directive.ReferenceHandler = ((*oneoffHandler)(nil))

// ExecOneOff executes a one-off directive.
// Returns nil if the directive is canceled for some reason during the execution.
func ExecOneOff(
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
	valDisposeCallback func(),
) (directive.Value, directive.Reference, error) {
	valCh := make(chan directive.Value, 1)
	_, ref, err := bus.AddDirective(
		dir,
		&oneoffHandler{
			disposeCb: valDisposeCallback,
			valCb: func(v directive.Value) {
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
