package loader

import (
	"context"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// resolver tracks a ExecController request
type resolver struct {
	directive  ExecController
	controller *Controller
}

// newResolver builds a new ExecController resolver.
func newResolver(directive ExecController, controller *Controller) *resolver {
	return &resolver{
		directive:  directive,
		controller: controller,
	}
}

// resolveExecController handles a ExecController directive.
func (c *Controller) resolveExecController(
	dir ExecController,
) (directive.Resolver, error) {
	// Check if the ExecController is meant for / compatible with us.
	// In this case, we handle all ExecController requests.
	return newResolver(dir, c), nil
}

// Resolve resolves the values.
// Any fatal error resolving the value is returned.
// When the context is canceled valCh will not be drained anymore.
func (c *resolver) Resolve(ctx context.Context, valCh chan<- directive.Value) error {
	// Construct and attach the new controller to the bus.
	le := c.controller.le
	bus := c.controller.bus
	config := c.directive.GetExecControllerConfig()
	factory := c.directive.GetExecControllerFactory()

	ci, err := factory.Construct(config, controller.ConstructOpts{
		Logger: le.WithField("controller", factory.GetControllerID()),
	})
	if err != nil {
		return err
	}

	// type assertion
	var civ ExecControllerValue = ci

	// emit the value
	select {
	case <-ctx.Done():
		return ctx.Err()
	case valCh <- civ:
	}

	// execute the controller
	return bus.ExecuteController(ctx, ci)
}

// _ is a type assertion
var _ directive.Resolver = ((*resolver)(nil))
