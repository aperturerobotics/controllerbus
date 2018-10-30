package loader

import (
	"context"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// resolver tracks a ExecController request
type resolver struct {
	ctx        context.Context
	directive  ExecController
	controller *Controller
}

// newResolver builds a new ExecController resolver.
func newResolver(ctx context.Context, directive ExecController, controller *Controller) *resolver {
	return &resolver{
		ctx:        ctx,
		directive:  directive,
		controller: controller,
	}
}

// resolveExecController handles a ExecController directive.
func (c *Controller) resolveExecController(
	ctx context.Context,
	dir ExecController,
) (directive.Resolver, error) {
	// Check if the ExecController is meant for / compatible with us.
	// In this case, we handle all ExecController requests.
	return newResolver(ctx, dir, c), nil
}

// Resolve resolves the values.
// Any fatal error resolving the value is returned.
// When the context is canceled valCh will not be drained anymore.
func (c *resolver) Resolve(ctx context.Context, vh directive.ResolverHandler) error {
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
	vid, ok := vh.AddValue(civ)
	if !ok {
		// value rejected, drop the controller on the floor.
		go ci.Close()
		return nil
	}

	// execute the controller
	go func() {
		err := bus.ExecuteController(c.ctx, ci)
		if err != nil {
			le.WithError(err).Warn("controller exited with error")
		}
		vh.RemoveValue(vid)
	}()

	return nil
}

// _ is a type assertion
var _ directive.Resolver = ((*resolver)(nil))
