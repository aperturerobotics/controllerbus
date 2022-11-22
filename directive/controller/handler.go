package controller

import (
	"sync/atomic"

	"github.com/aperturerobotics/controllerbus/directive"
)

// handler contains a directive handler.
type handler struct {
	// rel indicates the handler has been removed
	rel atomic.Bool
	// h is the directive handler
	h directive.Handler
}

// newHandler constructs a new handler.
func newHandler(h directive.Handler) *handler {
	return &handler{h: h}
}
