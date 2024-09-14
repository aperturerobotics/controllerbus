package directive

import (
	"context"
	"sync"
)

// AccessResolverFunc resolves a value similarly to a RefCount container.
//
// Accepts a released callback which can be called when the returned value is invalid.
// Returns a release function which will be called when the returned value is no longer needed.
type AccessResolverFunc[T comparable] func(ctx context.Context, released func()) (T, func(), error)

// AccessResolver resolves a directive with an access function.
type AccessResolver[T comparable] struct {
	getter AccessResolverFunc[T]
}

// NewAccessResolver constructs a new AccessResolver.
func NewAccessResolver[T comparable](getter AccessResolverFunc[T]) *AccessResolver[T] {
	return &AccessResolver[T]{getter: getter}
}

// Resolve resolves the values, emitting them to the handler.
func (r *AccessResolver[T]) Resolve(ctx context.Context, handler ResolverHandler) error {
	if r.getter == nil {
		return nil
	}

	for {
		handler.ClearValues()

		if err := ctx.Err(); err != nil {
			return context.Canceled
		}

		releasedCh := make(chan struct{})
		var releasedOnce sync.Once
		released := func() {
			releasedOnce.Do(func() {
				close(releasedCh)
			})
		}

		val, rel, err := r.getter(ctx, released)
		if err != nil {
			if rel != nil {
				rel()
			}
			return err
		}

		valID, valAccepted := handler.AddValue(val)
		if !valAccepted {
			if rel != nil {
				rel()
			}
			return nil
		}

		relCb := handler.AddValueRemovedCallback(valID, rel)
		select {
		case <-ctx.Done():
			relCb()
			if rel != nil {
				rel()
			}
			return context.Canceled
		case <-releasedCh:
			relCb()
			if rel != nil {
				rel()
			}
		}
	}
}

// _ is a type assertion
var _ Resolver = ((*AccessResolver[int])(nil))
