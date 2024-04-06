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
// if useCtx is set, uses the refcount resolver context to start the refcount container.
// If buildValue is set, will be called with the values.
// if buildValue returns nil, nil, ignores the value.
type RefCountResolver[T comparable, R ComparableValue] struct {
	rc         *refcount.RefCount[T]
	useCtx     bool
	buildValue func(ctx context.Context, val T) (R, error)
}

// NewRefCountResolver constructs a new RefCountResolver.
//
// This is the simplified constructor, use NewRefCountResolverWithXfrm to transform the value.
func NewRefCountResolver[T comparable](rc *refcount.RefCount[T]) *RefCountResolver[T, T] {
	return NewRefCountResolverWithXfrm[T, T](rc, false, nil)
}

// NewRefCountResolverWithXfrm constructs a new RefCountResolver with a transform func.
//
// if useCtx is set, uses the refcount resolver context to start the refcount container.
func NewRefCountResolverWithXfrm[T comparable, R ComparableValue](
	rc *refcount.RefCount[T],
	useCtx bool,
	buildValue func(ctx context.Context, val T) (R, error),
) *RefCountResolver[T, R] {
	return &RefCountResolver[T, R]{
		rc:         rc,
		useCtx:     useCtx,
		buildValue: buildValue,
	}
}

// Resolve resolves the values, emitting them to the handler.
func (r *RefCountResolver[T, R]) Resolve(ctx context.Context, handler ResolverHandler) error {
	return r.rc.Access(ctx, r.useCtx, func(ctx context.Context, tval T) error {
		var val Value = tval
		if r.buildValue != nil {
			var err error
			val, err = r.buildValue(ctx, tval)
			if err != nil {
				return err
			}
		}
		if val != nil {
			_, _ = handler.AddValue(val)
		}
		handler.MarkIdle(true)
		<-ctx.Done()
		handler.ClearValues()
		return context.Canceled
	})
}

// _ is a type assertion
var _ Resolver = ((*RefCountResolver[int, Value])(nil))
