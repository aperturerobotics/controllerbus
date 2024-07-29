package directive

import "github.com/sirupsen/logrus"

// LogHandler is a reference handler that logs when values are added and removed.
//
// Optionally wraps a base ReferenceHandler,
type LogHandler struct {
	// le is the logger
	// may be nil
	le *logrus.Entry
	// base is the base ReferenceHandler
	// may be nil
	base ReferenceHandler
}

// NewLogHandler constructs a logger ReferenceHandler.
//
// If set, base will be called after logging the values.
// base can be nil
// le can be nil
func NewLogHandler(
	le *logrus.Entry,
	base ReferenceHandler,
) ReferenceHandler {
	return &LogHandler{
		le:   le,
		base: base,
	}
}

// HandleValueAdded is called when a value is added to the directive.
func (h *LogHandler) HandleValueAdded(
	di Instance,
	v AttachedValue,
) {
	if h.le != nil {
		h.le.Debugf("HandleValueAdded: valueId(%v)", v.GetValueID())
	}
	if h.base != nil {
		h.base.HandleValueAdded(di, v)
	}
}

// HandleValueRemoved is called when a value is removed from the directive.
func (h *LogHandler) HandleValueRemoved(
	di Instance,
	v AttachedValue,
) {
	if h.le != nil {
		h.le.Debugf("HandleValueRemoved: valueId(%v)", v.GetValueID())
	}
	if h.base != nil {
		h.base.HandleValueRemoved(di, v)
	}
}

// HandleInstanceDisposed is called when a directive instance is disposed.
// This will occur if Close() is called on the directive instance.
func (h *LogHandler) HandleInstanceDisposed(di Instance) {
	if h.le != nil {
		h.le.Debug("HandleInstanceDisposed")
	}
	if h.base != nil {
		h.base.HandleInstanceDisposed(di)
	}
}

// _ is a type assertion
var _ ReferenceHandler = ((*LogHandler)(nil))
