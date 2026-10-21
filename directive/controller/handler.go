package controller

import (
	"context"
	"sync/atomic"

	bus_controller "github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// handler contains a directive handler.
type handler struct {
	// rel indicates the handler has been removed
	rel atomic.Bool
	// h is the directive handler
	h directive.Handler
	// controllerID is the ID of the controller h belongs to, if any.
	controllerID string
}

// newHandler constructs a new handler.
func newHandler(h directive.Handler) *handler {
	hnd := &handler{h: h}
	if ctrl, ok := h.(bus_controller.Controller); ok {
		hnd.controllerID = ctrl.GetControllerInfo().GetId()
	}
	return hnd
}

// handlerContext returns ctx with the handler's controller as the origin.
func (h *handler) handlerContext(ctx context.Context) context.Context {
	if h.controllerID == "" {
		return ctx
	}
	return directive.WithOriginController(ctx, h.controllerID)
}
