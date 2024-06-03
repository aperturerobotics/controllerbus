package directive

import (
	"context"
	"io"

	"github.com/aperturerobotics/util/ccontainer"
)

// TransformValueFunc transforms a value.
// rval is the returned transformed value to emit to the handler (if any).
// rval is ignored if ok = false.
// err, if any, will unwind the resolver with the error.
type TransformValueFunc[T, R Value] func(ctx context.Context, val T) (rval R, ok bool, err error)

// WatchableTransformResolver wraps a Watchable into a resolver and transforms the values.
type WatchableTransformResolver[T comparable, R Value] struct {
	// ctr is the watchable
	ctr  ccontainer.Watchable[T]
	xfrm TransformValueFunc[T, R]
}

// NewWatchableTransformResolver constructs a new WatchableTransformResolver.
func NewWatchableTransformResolver[T comparable, R Value](ctr ccontainer.Watchable[T], xfrm TransformValueFunc[T, R]) *WatchableTransformResolver[T, R] {
	return &WatchableTransformResolver[T, R]{ctr: ctr, xfrm: xfrm}
}

// Resolve resolves the values, emitting them to the handler.
func (r *WatchableTransformResolver[T, R]) Resolve(ctx context.Context, handler ResolverHandler) error {
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
				result, ok, err := r.xfrm(ctx, value)
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
				var accepted bool
				valID, accepted = handler.AddValue(result)
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
var _ Resolver = ((*WatchableTransformResolver[bool, string])(nil))
