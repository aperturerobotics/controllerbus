package directive

import "context"

// ValueResolver resolves a directive with a static list of values.
type ValueResolver[T any] struct {
	// vals is the static list of values
	vals []T
}

// NewValueResolver constructs a new ValueResolver.
func NewValueResolver[T any](vals []T) *ValueResolver[T] {
	return &ValueResolver[T]{vals: vals}
}

// Resolve resolves the values, emitting them to the handler.
func (r *ValueResolver[T]) Resolve(ctx context.Context, handler ResolverHandler) error {
	existingVals := handler.CountValues(false)
	if existingVals == len(r.vals) {
		return nil
	}
	if existingVals != 0 {
		handler.ClearValues()
	}
	for _, value := range r.vals {
		_, _ = handler.AddValue(value)
	}
	return nil
}

// _ is a type assertion
var _ Resolver = ((*ValueResolver[int])(nil))
