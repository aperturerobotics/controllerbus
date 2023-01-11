package directive

import (
	"context"

	"github.com/aperturerobotics/util/refcount"
)

// RefCountResolver resolves a directive with a RefCount container.
//
// Adds a reference, waits for a value, and adds it to the handler.
// Removes the value and waits for a new one when the value is released.
// Calls MarkIdle when the value has been added.
// buildValue can be nil
// If buildValue is set, will be called with the values.
// if buildValue returns nil, nil, ignores the value.
type RefCountResolver[T comparable] struct {
	rc         *refcount.RefCount[T]
	buildValue func(val T) (Value, error)
}

// NewRefCountResolver constructs a new RefCountResolver.
func NewRefCountResolver[T comparable](
	rc *refcount.RefCount[T],
	buildValue func(val T) (Value, error),
) *RefCountResolver[T] {
	return &RefCountResolver[T]{rc: rc, buildValue: buildValue}
}

// Resolve resolves the values, emitting them to the handler.
func (r *RefCountResolver[T]) Resolve(ctx context.Context, handler ResolverHandler) error {
	return r.rc.Access(ctx, func(ctx context.Context, tval T) error {
		var val Value = tval
		if r.buildValue != nil {
			var err error
			val, err = r.buildValue(tval)
			if err != nil {
				return err
			}
		}
		if val != nil {
			_, _ = handler.AddValue(val)
		}
		handler.MarkIdle()
		<-ctx.Done()
		handler.ClearValues()
		return context.Canceled
	})
}

// _ is a type assertion
var _ Resolver = ((*RefCountResolver[int])(nil))
