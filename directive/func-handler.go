package directive

import "context"

// FuncHandler implements Handler with a function.
type FuncHandler struct {
	fn HandlerFunc
}

// NewFuncHandler constructs a new FuncHandler.
func NewFuncHandler(fn HandlerFunc) *FuncHandler {
	return &FuncHandler{fn: fn}
}

// HandleDirective implements Handler.
func (h *FuncHandler) HandleDirective(ctx context.Context, di Instance) ([]Resolver, error) {
	if h.fn == nil {
		return nil, nil
	}
	return h.fn(ctx, di)
}

// _ is a type assertion
var _ Handler = ((*FuncHandler)(nil))
