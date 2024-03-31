package resolver

import (
	"context"
	"strings"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// Controller implements the controller resolver controller.
// This controller responds to LoadControllerWithConfig directives.
type Controller struct {
	// le is the default logger
	le *logrus.Entry
	// bus is the controller bus
	bus bus.Bus
	// resolver is the factory resolver
	resolver controller.FactoryResolver
	// subCtx is the sub-ctx used for signaling close to resolvers
	subCtx       context.Context
	subCtxCancel context.CancelFunc
}

// NewController constructs a new controller with a resolver.
func NewController(le *logrus.Entry, bus bus.Bus, resolver controller.FactoryResolver) *Controller {
	c := &Controller{le: le, resolver: resolver, bus: bus}
	c.subCtx, c.subCtxCancel = context.WithCancel(context.Background())
	return c
}

// GetControllerID returns the controller ID.
func (c *Controller) GetControllerID() string {
	return strings.Join([]string{
		"controllerbus",
		"resolver",
		c.resolver.GetResolverID(),
	}, "/")
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() *controller.Info {
	return controller.NewInfo(
		c.GetControllerID(),
		c.resolver.GetResolverVersion(),
		"controller resolver "+c.resolver.GetResolverID()+"@"+c.resolver.GetResolverVersion().String(),
	)
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any unexpected errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (c *Controller) HandleDirective(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
	dir := di.GetDirective()
	switch d := dir.(type) {
	case LoadControllerWithConfig:
		return c.resolveLoadControllerWithConfig(ctx, d)
	case LoadConfigConstructorByID:
		return c.resolveLoadConfigConstructorByID(ctx, d)
	case LoadFactoryByConfig:
		return c.resolveLoadFactoryByConfig(ctx, d)
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
	// When closing resolver controller, we need to cleanup all yielded resources.
	// At this point we shouldn't have any additional requests being fired.
	c.subCtxCancel()
	return nil
}

var _ controller.Controller = ((*Controller)(nil))
