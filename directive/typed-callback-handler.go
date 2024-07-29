package directive

// TypedCallbackHandler is a generic reference handler that uses function callbacks.
//
// Values with unknown type are optionally passed to the optional unknownTypeHandler.
type TypedCallbackHandler[T ComparableValue] struct {
	valCb     func(TypedAttachedValue[T])
	removedCb func(TypedAttachedValue[T])
	disposeCb func()

	unknownTypeHandler ReferenceHandler
}

// NewTypedCallbackHandler wraps callback functions into a handler object.
//
// unknownTypeHandler can be nil
func NewTypedCallbackHandler[T ComparableValue](
	valCb func(TypedAttachedValue[T]),
	removedCb func(TypedAttachedValue[T]),
	disposeCb func(),
	unknownTypeHandler ReferenceHandler,
) ReferenceHandler {
	return &TypedCallbackHandler[T]{
		valCb:              valCb,
		removedCb:          removedCb,
		disposeCb:          disposeCb,
		unknownTypeHandler: unknownTypeHandler,
	}
}

// HandleValueAdded is called when a value is added to the directive.
func (h *TypedCallbackHandler[T]) HandleValueAdded(
	di Instance,
	v AttachedValue,
) {
	val, ok := v.GetValue().(T)
	if ok {
		if h.valCb != nil {
			h.valCb(NewTypedAttachedValue(v.GetValueID(), val))
		}
	} else if h.unknownTypeHandler != nil {
		h.unknownTypeHandler.HandleValueAdded(di, v)
	}
}

// HandleValueRemoved is called when a value is removed from the directive.
func (h *TypedCallbackHandler[T]) HandleValueRemoved(
	di Instance,
	v AttachedValue,
) {
	val, ok := v.GetValue().(T)
	if ok {
		if h.removedCb != nil {
			h.removedCb(NewTypedAttachedValue(v.GetValueID(), val))
		}
	} else if h.unknownTypeHandler != nil {
		h.unknownTypeHandler.HandleValueRemoved(di, v)
	}
}

// HandleInstanceDisposed is called when a directive instance is disposed.
// This will occur if Close() is called on the directive instance.
func (h *TypedCallbackHandler[T]) HandleInstanceDisposed(i Instance) {
	if h.disposeCb != nil {
		h.disposeCb()
	}
	if h.unknownTypeHandler != nil {
		h.unknownTypeHandler.HandleInstanceDisposed(i)
	}
}

// _ is a type assertion
var _ ReferenceHandler = ((*TypedCallbackHandler[int])(nil))
