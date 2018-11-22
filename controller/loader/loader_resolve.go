package loader

import (
	"context"
	"time"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
)

// resolver tracks a ExecController request
type resolver struct {
	ctx         context.Context
	directive   ExecController
	controller  *Controller
	execBackoff backoff.BackOff
	lastErr     error
}

// newResolver builds a new ExecController resolver.
func newResolver(ctx context.Context, directive ExecController, controller *Controller) *resolver {
	ebo := backoff.NewExponentialBackOff()
	ebo.InitialInterval = time.Millisecond * 500
	ebo.MaxInterval = time.Second * 30
	ebo.Multiplier = 2
	ebo.MaxElapsedTime = time.Minute
	return &resolver{
		ctx:         ctx,
		directive:   directive,
		controller:  controller,
		execBackoff: ebo,
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
	factory := c.directive.GetExecControllerFactory()
	le := c.controller.le.WithField("controller", factory.GetControllerID())
	bus := c.controller.bus
	config := c.directive.GetExecControllerConfig()

	ci, err := factory.Construct(config, controller.ConstructOpts{
		Logger: le,
	})
	if err != nil {
		return err
	}

	// execute the controller
	var execNextBo time.Duration
	if c.lastErr == nil {
		c.execBackoff.Reset()
	} else {
		execNextBo = c.execBackoff.NextBackOff()
		if execNextBo == -1 {
			return errors.Wrap(c.lastErr, "backoff timeout exceeded")
		}
	}

	if execNextBo != 0 {
		le.
			WithField("backoff-duration", execNextBo.String()).
			Debug("backing off before controller re-start")
		boTimer := time.NewTimer(execNextBo)
		defer boTimer.Stop()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-boTimer.C:
		}
	}

	go func() {
		// type assertion
		var civ ExecControllerValue = ci

		// emit the value
		vid, ok := vh.AddValue(civ)
		if !ok {
			// value rejected, drop the controller on the floor.
			ci.Close()
			return
		}

		le.Debug("starting controller")
		t1 := time.Now()
		err := bus.ExecuteController(c.ctx, ci)
		c.lastErr = err
		le := le.WithField("exec-time", time.Now().Sub(t1).String())
		if err != nil {
			le.WithError(err).Warn("controller exited with error")
			vh.RemoveValue(vid)
		} else {
			le.Debug("controller exited normally")
		}
	}()

	return nil
}

// _ is a type assertion
var _ directive.Resolver = ((*resolver)(nil))
