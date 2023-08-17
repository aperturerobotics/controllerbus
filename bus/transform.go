package bus

import (
	"github.com/aperturerobotics/controllerbus/directive"
)

// TransformHandler is a reference handler that transforms values to a resolver handler.
// xfrm is the transformation callback called with each AttachedValue.
// xfrm should return the Value to add to the ResolverHandler.
// if xfrm returns nil, false, the value is ignored.
// if xfrm is nil, acts identically to PassThruHandler.
type TransformHandler struct {
	handler   directive.ResolverHandler
	xfrm      func(val directive.AttachedValue) (directive.Value, bool)
	idMapping map[uint32]uint32
	disposeCb func()
}

// NewTransformHandler builds a new pass-through handler.
func NewTransformHandler(
	passthru directive.ResolverHandler,
	xfrm func(val directive.AttachedValue) (directive.Value, bool),
	disposeCb func(),
) directive.ReferenceHandler {
	return &TransformHandler{
		handler:   passthru,
		xfrm:      xfrm,
		idMapping: make(map[uint32]uint32),
		disposeCb: disposeCb,
	}
}

// HandleValueAdded is called when a value is added to the directive.
func (h *TransformHandler) HandleValueAdded(
	_ directive.Instance,
	v directive.AttachedValue,
) {
	if h.handler == nil {
		return
	}

	var val directive.Value
	if h.xfrm != nil {
		var valOk bool
		val, valOk = h.xfrm(v)
		if !valOk {
			return
		}
	} else {
		val = v.GetValue
	}

	id, accepted := h.handler.AddValue(val)
	if accepted {
		h.idMapping[v.GetValueID()] = id
	}
}

// HandleValueRemoved is called when a value is removed from the directive.
func (h *TransformHandler) HandleValueRemoved(
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
func (h *TransformHandler) HandleInstanceDisposed(di directive.Instance) {
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
var _ directive.ReferenceHandler = ((*TransformHandler)(nil))
