package bus

import (
	"github.com/aperturerobotics/controllerbus/directive"
)

// CallbackHandler is a reference handler that uses function callbacks.
type CallbackHandler = directive.CallbackHandler

// NewCallbackHandler wraps callback functions into a handler object.
func NewCallbackHandler(
	valCb func(directive.AttachedValue),
	removedCb func(directive.AttachedValue),
	disposeCb func(),
) directive.ReferenceHandler {
	return directive.NewCallbackHandler(valCb, removedCb, disposeCb)
}
