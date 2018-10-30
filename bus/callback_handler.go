package bus

import (
	"github.com/aperturerobotics/controllerbus/directive"
)

// CallbackHandler is a reference handler that uses function callbacks.
type CallbackHandler struct {
	valCb     func(directive.Value)
	removedCb func(directive.Value)
	disposeCb func()
}

// NewCallbackHandler wraps callback functions into a handler object.
func NewCallbackHandler(
	valCb func(directive.Value),
	removedCb func(directive.Value),
	disposeCb func(),
) directive.ReferenceHandler {
	return &CallbackHandler{
		valCb:     valCb,
		removedCb: removedCb,
		disposeCb: disposeCb,
	}
}

// HandleValueAdded is called when a value is added to the directive.
func (h *CallbackHandler) HandleValueAdded(_ directive.Instance, v directive.Value) {
	if h.valCb != nil {
		h.valCb(v)
	}
}

// HandleValueRemoved is called when a value is removed from the directive.
func (h *CallbackHandler) HandleValueRemoved(_ directive.Instance, v directive.Value) {
	if h.removedCb != nil {
		h.removedCb(v)
	}
}

// HandleInstanceDisposed is called when a directive instance is disposed.
// This will occur if Close() is called on the directive instance.
func (h *CallbackHandler) HandleInstanceDisposed(directive.Instance) {
	if h.disposeCb != nil {
		h.disposeCb()
	}
}

// _ is a type assertion
var _ directive.ReferenceHandler = ((*CallbackHandler)(nil))
