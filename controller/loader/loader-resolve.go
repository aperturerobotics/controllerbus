package loader

import (
	"context"
	"time"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	backoff "github.com/aperturerobotics/util/backoff/cbackoff"
	"github.com/pkg/errors"
)

// resolver runs the controller an ExecController directive requests.
type resolver struct {
	// ctx is the lifetime of the attached controller.
	ctx context.Context
	// dir is the directive being resolved.
	dir ExecController
	// controller is the loader that owns the resolver.
	controller *Controller
}

// newResolver builds a new ExecController resolver.
func newResolver(ctx context.Context, dir ExecController, controller *Controller) *resolver {
	return &resolver{
		ctx:        ctx,
		dir:        dir,
		controller: controller,
	}
}

// newExecBackoff constructs the default exec backoff.
func newExecBackoff() backoff.BackOff {
	// Retry quickly at first, then at most every two seconds.
	ebo := backoff.NewExponentialBackOff()
	ebo.InitialInterval = time.Millisecond * 100
	ebo.Multiplier = 1.8
	ebo.MaxInterval = time.Second * 2
	return ebo
}

// resolveExecController handles every ExecController directive.
func (c *Controller) resolveExecController(ctx context.Context, dir ExecController) ([]directive.Resolver, error) {
	return directive.R(newResolver(ctx, dir, c), nil)
}

// Resolve constructs the controller and runs it on the bus, retrying with
// backoff after a failure. The value carrying the controller is emitted only
// after the bus attaches it, so a directive added after the value is observed
// is offered to the controller. The controller stays attached until ctx is
// canceled or the controller is removed from the bus.
func (c *resolver) Resolve(ctx context.Context, vh directive.ResolverHandler) error {
	// Read the factory and config to construct.
	config := c.dir.GetExecControllerConfig()
	factory := c.dir.GetExecControllerFactory()
	le := c.controller.le.WithField("config", factory.GetConfigID())
	bus := c.controller.bus

	// Use the directive's retry backoff, or the default.
	var execBackoff backoff.BackOff
	if buildBackoff := c.dir.GetExecControllerRetryBackoff(); buildBackoff != nil {
		execBackoff = buildBackoff()
	}
	if execBackoff == nil {
		execBackoff = newExecBackoff()
	}

	// Run the controller until ctx is canceled, constructing a new instance
	// after each failure because the bus closes a controller that fails.
	var lastErr error
	for {
		_ = vh.ClearValues()

		// Back off after a failure, publishing the error and the retry time.
		if lastErr != nil {
			delay := execBackoff.NextBackOff()
			if delay == backoff.Stop {
				return errors.Wrap(lastErr, "backoff timeout exceeded")
			}
			le.
				WithField("backoff-duration", delay.String()).
				Debug("backing off before controller re-start")
			now := time.Now()
			vid, vidOk := vh.AddValue(NewExecControllerValue(now, now.Add(delay), nil, lastErr))
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
			if vidOk {
				vh.RemoveValue(vid)
			}
		}

		// Construct the controller.
		t1 := time.Now()
		ci, err := factory.Construct(ctx, config, controller.ConstructOpts{Logger: le})
		if err != nil {
			lastErr = err
			continue
		}
		if ci == nil {
			err := errors.New("controller construct returned nil")
			le.Warn(err.Error())
			return err
		}

		// Attach the controller, then publish it. The bus closes the
		// controller when attaching fails or Execute returns an error.
		le.Debug("starting controller")
		exited := make(chan error, 1)
		release, err := bus.AddController(c.ctx, ci, func(err error) {
			exited <- err
		})
		if err != nil {
			lastErr = err
			continue
		}
		_, _ = vh.AddValue(NewExecControllerValue(t1, time.Time{}, ci, nil))

		// Hold the controller until ctx is canceled or Execute fails.
		select {
		case <-ctx.Done():
			_ = vh.ClearValues()
			release()
			return context.Canceled
		case err := <-exited:
			// A nil result means the controller was removed from the bus.
			if err == nil {
				_ = vh.ClearValues()
				return nil
			}
			le.
				WithField("exec-dur", time.Since(t1).String()).
				WithError(err).
				Warn("controller exited with error")
			lastErr = err
		}
	}
}

// _ is a type assertion
var _ directive.Resolver = ((*resolver)(nil))
