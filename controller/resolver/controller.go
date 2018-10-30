package resolver

import (
	"context"
	"strings"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// Controller implements the controller resolver controller.
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

// GetControllerID returns the controller ID.
func (c *Controller) GetControllerID() string {
	return strings.Join([]string{"controllerbus", "resolver", c.resolver.GetResolverID(), c.resolver.GetResolverVersion().String()}, "/")
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() controller.Info {
	return controller.NewInfo(
		c.GetControllerID(),
		c.resolver.GetResolverVersion(),
		"controller resolver "+c.resolver.GetResolverID()+"@"+c.resolver.GetResolverVersion().String(),
	)
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (c *Controller) HandleDirective(ctx context.Context, di directive.Instance) (directive.Resolver, error) {
	dir := di.GetDirective()
	if d, ok := dir.(LoadControllerWithConfig); ok {
		return c.resolveLoadControllerWithConfig(ctx, d)
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
