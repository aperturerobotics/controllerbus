package directive

import "context"

// AccessResolverFunc resolves a value similarly to a RefCount container.
//
// Accepts a released callback which can be called when the returned value is invalid.
// Returns a release function called once when the returned value is no longer needed.
// The resolver owns that call, including when value removal races cancellation.
type AccessResolverFunc[T comparable] func(ctx context.Context, released func()) (T, func(), error)

// AccessResolver resolves a directive with an access function.
type AccessResolver[T comparable] struct {
	// getter acquires the value and its release function.
	getter AccessResolverFunc[T]
}

// NewAccessResolver constructs a new AccessResolver.
func NewAccessResolver[T comparable](getter AccessResolverFunc[T]) *AccessResolver[T] {
	return &AccessResolver[T]{getter: getter}
}

// Resolve resolves the values, emitting them to the handler.
func (r *AccessResolver[T]) Resolve(ctx context.Context, handler ResolverHandler) error {
	// An empty accessor publishes nothing.
	if r.getter == nil {
		return nil
	}

	// Acquire each replacement only after retiring the previous value.
	for {
		// Retire the old publication before observing cancellation.
		handler.ClearValues()
		if ctx.Err() != nil {
			return context.Canceled
		}

		// Coalesce invalidations without blocking the component publishing them.
		invalidated := make(chan struct{}, 1)
		invalidate := func() {
			select {
			case invalidated <- struct{}{}:
			default:
			}
		}

		// Acquire the current value and retire empty or failed acquisitions.
		val, rel, err := r.getter(ctx, invalidate)
		var empty T
		if err != nil || val == empty {
			if rel != nil {
				rel()
			}
			return err
		}

		// Publish the acquisition, releasing it if the handler refuses it.
		valID, accepted := handler.AddValue(val)
		if !accepted {
			if rel != nil {
				rel()
			}
			return nil
		}

		// Removal only notifies this owner, even if the callback runs late.
		removeCallback := handler.AddValueRemovedCallback(valID, invalidate)
		select {
		case <-ctx.Done():
		case <-invalidated:
		}

		// Stop callback delivery before retiring this acquisition.
		removeCallback()
		if rel != nil {
			rel()
		}

		// End the owner when its directive is canceled.
		if ctx.Err() != nil {
			return context.Canceled
		}
	}
}

// _ is a type assertion
var _ Resolver = (*AccessResolver[int])(nil)
