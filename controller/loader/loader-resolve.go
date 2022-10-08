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

// newExecBackoff constructs the default exec backoff.
func newExecBackoff() backoff.BackOff {
	ebo := backoff.NewExponentialBackOff()
	ebo.InitialInterval = time.Millisecond * 100
	ebo.Multiplier = 1.8
	ebo.MaxInterval = time.Second * 2
	// ebo.MaxElapsedTime = time.Minute
	return ebo
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
	config := c.directive.GetExecControllerConfig()
	factory := c.directive.GetExecControllerFactory()

	le := c.controller.le.WithField("config", factory.GetConfigID())
	bus := c.controller.bus

	ci, ciErr := factory.Construct(config, controller.ConstructOpts{
		Logger: le,
	})
	if ciErr != nil {
		return ciErr
	}

	// execute the controller w/ retry backoff.
	execBackoff := newExecBackoff()
	var lastErr error
	var execNextBo time.Duration
	for {
		// if lastErr == nil: first run
		if lastErr != nil {
			execNextBo = execBackoff.NextBackOff()
			if execNextBo == backoff.Stop {
				return errors.Wrap(lastErr, "backoff timeout exceeded")
			}
		} else {
			execNextBo = 0
		}

		if execNextBo != 0 {
			le.
				WithField("backoff-duration", execNextBo.String()).
				Debug("backing off before controller re-start")
			boTimer := time.NewTimer(execNextBo)
			defer boTimer.Stop()

			// emit the value
			now := time.Now()
			vid, vidOk := vh.AddValue(NewExecControllerValue(
				now,
				now.Add(execNextBo),
				nil,
				lastErr,
			))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-boTimer.C:
				if vidOk {
					vh.RemoveValue(vid)
				}
			}
		}
		lastErr = nil

		// type assertion
		t1 := time.Now()

		// emit the value
		vid, vidOk := vh.AddValue(NewExecControllerValue(
			t1,
			time.Time{},
			ci,
			nil,
		))

		le.Debug("starting controller")
		execErr := bus.ExecuteController(c.ctx, ci)
		le := le.WithField("exec-time", time.Since(t1).String())
		if execErr != nil && execErr != context.Canceled {
			le.WithError(execErr).Warn("controller exited with error")
			lastErr = execErr
		} else {
			le.Debug("controller exited normally")
			execErr = nil
		}
		// note: if context canceled, loop once more to check.
		if execErr == nil {
			// controller Execute() is complete.
			return nil
		}
		// remove old value, will be replaced next loop.
		if vidOk {
			vh.RemoveValue(vid)
		}
	}
}

// _ is a type assertion
var _ directive.Resolver = ((*resolver)(nil))
