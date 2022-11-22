package controller

import (
	"context"

	"github.com/aperturerobotics/controllerbus/directive"
)

type attachedHandler struct {
	// Context is the attached handler context.
	Context context.Context
	// Cancel cancels context.
	Cancel context.CancelFunc
	// Handler is the handler.
	Handler directive.Handler
}

// newAttachedHandler builds a new attached handler.
func newAttachedHandler(ctx context.Context, hnd directive.Handler) *attachedHandler {
	h := &attachedHandler{Handler: hnd}
	h.Context, h.Cancel = context.WithCancel(ctx)
	return h
}
