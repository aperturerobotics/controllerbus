package directive

import (
	"context"
	"io"

	"github.com/aperturerobotics/util/ccontainer"
)

// WatchableResolver wraps a Watchable into a resolver.
type WatchableResolver[T comparable] struct {
	// ctr is the watchable
	ctr ccontainer.Watchable[T]
}

// NewWatchableResolver constructs a new retry resolver.
func NewWatchableResolver[T comparable](ctr ccontainer.Watchable[T]) *WatchableResolver[T] {
	return &WatchableResolver[T]{ctr: ctr}
}

// Resolve resolves the values, emitting them to the handler.
func (r *WatchableResolver[T]) Resolve(ctx context.Context, handler ResolverHandler) error {
	var valID uint32
	_ = handler.ClearValues()
	var empty T
	err := ccontainer.WatchChanges(
		ctx,
		empty,
		r.ctr,
		func(value T) error {
			if valID != 0 {
				_, _ = handler.RemoveValue(valID)
				valID = 0
			}
			if value != empty {
				var accepted bool
				valID, accepted = handler.AddValue(value)
				if !accepted {
					return io.EOF
				}
			}
			return nil
		},
		nil,
	)
	if err == io.EOF {
		return nil
	}
	return err
}

// _ is a type assertion
var _ Resolver = ((*WatchableResolver[bool])(nil))
