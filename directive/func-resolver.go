package directive

import "context"

// FuncResolver resolves a directive with a function.
type FuncResolver struct {
	fn func(ctx context.Context, handler ResolverHandler) error
}

// NewFuncResolver constructs a new FuncResolver.
func NewFuncResolver(fn func(ctx context.Context, handler ResolverHandler) error) *FuncResolver {
	return &FuncResolver{fn: fn}
}

// Resolve resolves the values, emitting them to the handler.
func (r *FuncResolver) Resolve(ctx context.Context, handler ResolverHandler) error {
	if r.fn == nil {
		return nil
	}
	return r.fn(ctx, handler)
}

// _ is a type assertion
var _ Resolver = ((*FuncResolver)(nil))
