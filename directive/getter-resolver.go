package directive

import "context"

// GetterResolver resolves a directive with a getter function.
type GetterResolver[T any] struct {
	getter func(ctx context.Context) (T, error)
}

// NewGetterResolver constructs a new GetterResolver.
func NewGetterResolver[T any](getter func(ctx context.Context) (T, error)) *GetterResolver[T] {
	return &GetterResolver[T]{getter: getter}
}

// Resolve resolves the values, emitting them to the handler.
func (r *GetterResolver[T]) Resolve(ctx context.Context, handler ResolverHandler) error {
	handler.ClearValues()
	if r.getter == nil {
		return nil
	}
	val, err := r.getter(ctx)
	if err != nil {
		return err
	}
	handler.AddValue(val)
	return nil
}

// _ is a type assertion
var _ Resolver = ((*GetterResolver[int])(nil))
