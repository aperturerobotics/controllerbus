package bus

import (
	"github.com/aperturerobotics/controllerbus/directive"
)

// PassThruHandler is a reference handler that passes through to a resolver handler.
type PassThruHandler struct {
	handler   directive.ResolverHandler
	idMapping map[uint32]uint32
	disposeCb func()
}

// NewPassThruHandler builds a new pass-through handler.
func NewPassThruHandler(
	passthru directive.ResolverHandler,
	disposeCb func(),
) directive.ReferenceHandler {
	return &PassThruHandler{
		handler:   passthru,
		idMapping: make(map[uint32]uint32),
		disposeCb: disposeCb,
	}
}

// HandleValueAdded is called when a value is added to the directive.
func (h *PassThruHandler) HandleValueAdded(
	_ directive.Instance,
	v directive.AttachedValue,
) {
	if h.handler != nil {
		id, accepted := h.handler.AddValue(v.GetValue())
		if accepted {
			h.idMapping[v.GetValueID()] = id
		}
	}
}

// HandleValueRemoved is called when a value is removed from the directive.
func (h *PassThruHandler) HandleValueRemoved(
	_ directive.Instance,
	v directive.AttachedValue,
) {
	if h.handler != nil {
		mapping, ok := h.idMapping[v.GetValueID()]
		if ok {
			_, _ = h.handler.RemoveValue(mapping)
		}
	}
}

// HandleInstanceDisposed is called when a directive instance is disposed.
// This will occur if Close() is called on the directive instance.
func (h *PassThruHandler) HandleInstanceDisposed(di directive.Instance) {
	if h.handler != nil {
		for _, valID := range h.idMapping {
			_, _ = h.handler.RemoveValue(valID)
		}
	}
	if h.disposeCb != nil {
		h.disposeCb()
	}
}

// _ is a type assertion
var _ directive.ReferenceHandler = ((*PassThruHandler)(nil))
