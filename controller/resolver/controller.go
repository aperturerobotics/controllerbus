package resolver

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// Controller implements the controller resolver contlroller.
// This controller responds to LoadControllerWithConfig directives.
type Controller struct {
	// bus is the controller bus
	bus bus.Bus
	// resolver is the factory resolver
	resolver controller.FactoryResolver
}

// NewController constructs a new controller with a resolver.
func NewController(bus bus.Bus, resolver controller.FactoryResolver) *Controller {
	return &Controller{resolver: resolver, bus: bus}
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (c *Controller) HandleDirective(di directive.Instance) (directive.Resolver, error) {
	dir := di.GetDirective()
	if d, ok := dir.(LoadControllerWithConfig); ok {
		return c.resolveLoadControllerWithConfig(d)
	}

	return nil, nil
}

// Execute executes the given controller.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *Controller) Execute(ctx context.Context) error {
	return nil
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (c *Controller) Close() error {
	return nil
}

var _ controller.Controller = ((*Controller)(nil))
