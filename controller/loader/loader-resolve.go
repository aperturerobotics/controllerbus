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

// execStableDuration is how long a controller must run before its failure
// starts a fresh backoff.
const execStableDuration = time.Second * 10

// newExecBackoff constructs the default exec backoff. It never stops: a
// controller that fails for a long time, such as on a full disk, starts again
// once the cause clears.
func newExecBackoff() backoff.BackOff {
	// Retry quickly at first, then at most every two seconds.
	ebo := backoff.NewExponentialBackOff()
	ebo.InitialInterval = time.Millisecond * 100
	ebo.Multiplier = 1.8
	ebo.MaxInterval = time.Second * 2
	ebo.MaxElapsedTime = 0
	return ebo
}

// resolveExecController handles every ExecController directive.
func (c *Controller) resolveExecController(ctx context.Context, dir ExecController) ([]directive.Resolver, error) {
	return directive.R(newResolver(ctx, dir, c), nil)
}

// Resolve constructs the controller and runs it on the bus.
// It retries with backoff after a failure. The backoff counts only the
// failures since the controller last ran for execStableDuration, so a
// directive's backoff that stops gives up on a controller that cannot stay
// up, not on one that ran for hours before failing.
// The value carrying the controller is emitted after the bus attaches it.
// A directive added after the value is observed is offered to the controller.
// The controller stays attached until ctx is canceled or it is removed from the bus.
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
			execDur := time.Since(t1)
			if execDur >= execStableDuration {
				execBackoff.Reset()
			}
			le.
				WithField("exec-dur", execDur.String()).
				WithError(err).
				Warn("controller exited with error")
			lastErr = err
		}
	}
}

// _ is a type assertion
var _ directive.Resolver = ((*resolver)(nil))
