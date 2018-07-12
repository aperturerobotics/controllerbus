package loader

import (
	"context"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// resolver tracks a LoadController request
type resolver struct {
	directive  LoadController
	controller *Controller
}

// newResolver builds a new LoadController resolver.
func newResolver(directive LoadController, controller *Controller) *resolver {
	return &resolver{
		directive:  directive,
		controller: controller,
	}
}

// resolveLoadController handles a LoadController directive.
func (c *Controller) resolveLoadController(
	dir LoadController,
) (directive.Resolver, error) {
	// Check if the LoadController is meant for / compatible with us.
	// In this case, we handle all LoadController requests.

	return newResolver(dir, c), nil
}

// Resolve resolves the values.
// Any fatal error resolving the value is returned.
// When the context is canceled valCh will not be drained anymore.
func (c *resolver) Resolve(ctx context.Context, valCh chan<- directive.Value) error {
	// Construct and attach the new controller to the bus.
	le := c.controller.le
	bus := c.controller.bus
	config := c.directive.GetLoadControllerConfig()
	factory := c.directive.GetLoadControllerFactory()

	ci, err := factory.Construct(config, controller.ConstructOpts{
		Logger: le.WithField("controller", factory.GetControllerID()),
	})
	if err != nil {
		return err
	}

	// type assertion
	var _ LoadControllerValue = ci

	// emit the value
	select {
	case <-ctx.Done():
		return ctx.Err()
	case valCh <- ci:
	}

	// execute the controller
	return bus.ExecuteController(ci)
}

// _ is a type assertion
var _ directive.Resolver = ((*resolver)(nil))
