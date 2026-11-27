package controller

import (
	"context"
	"sync/atomic"

	bus_controller "github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// handler is a directive handler's registration with a Controller.
//
// Removal retires the registration, and retirement is terminal. The
// controller retires a registration with its mutex held, before it cancels the
// registration's resolvers, and attaches returned resolvers with the same mutex
// held only while the registration is live. A HandleDirective call that was
// pending at removal therefore attaches nothing.
type handler struct {
	// rel indicates the registration has been retired
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

// handleDirective calls the handler for di with the handler's controller as
// the origin of ctx. The caller must not hold the controller mutex.
func (h *handler) handleDirective(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
	return h.h.HandleDirective(h.handlerContext(ctx), di)
}

// retire marks the registration removed and cancels its resolvers in each of
// dirs. The caller holds the controller mutex and has already taken the
// registration out of the controller's handler set.
func (h *handler) retire(dirs []*directiveInstance) {
	h.rel.Store(true)
	for _, di := range dirs {
		di.removeHandlerLocked(h)
	}
}
