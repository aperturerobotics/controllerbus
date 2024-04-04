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
	di         directive.Instance
	dir        ExecController
	controller *Controller
}

// newResolver builds a new ExecController resolver.
func newResolver(ctx context.Context, di directive.Instance, dir ExecController, controller *Controller) *resolver {
	return &resolver{
		ctx:        ctx,
		di:         di,
		dir:        dir,
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
	di directive.Instance,
	dir ExecController,
) ([]directive.Resolver, error) {
	// Check if the ExecController is meant for / compatible with us.
	// In this case, we handle all ExecController requests.
	return directive.R(newResolver(ctx, di, dir, c), nil)
}

// Resolve resolves the values.
// Any fatal error resolving the value is returned.
// When the context is canceled valCh will not be drained anymore.
func (c *resolver) Resolve(ctx context.Context, vh directive.ResolverHandler) error {
	// Construct and attach the new controller to the bus.
	config := c.dir.GetExecControllerConfig()
	factory := c.dir.GetExecControllerFactory()

	var execBackoff backoff.BackOff
	if buildBackoff := c.dir.GetExecControllerRetryBackoff(); buildBackoff != nil {
		execBackoff = buildBackoff()
	}
	if execBackoff == nil {
		execBackoff = newExecBackoff()
	}

	configID := factory.GetConfigID()
	le := c.controller.le.WithField("config", configID)
	bus := c.controller.bus

	// execute the controller w/ retry backoff.
	var lastErr error
	var execNextBo time.Duration
	var ci controller.Controller
	closeCi := func() {
		if ci == nil {
			return
		}
		err := ci.Close()
		if err != nil && err != context.Canceled {
			le.WithError(err).Warn("controller close returned an error")
		}
	}

	for {
		// Clear any old values
		_ = vh.ClearValues()

		// if lastErr == nil: first run
		if lastErr != nil {
			execNextBo = execBackoff.NextBackOff()
			if execNextBo == backoff.Stop {
				closeCi()
				return errors.Wrap(lastErr, "backoff timeout exceeded")
			}
		} else {
			execNextBo = 0
		}

		// if we need to wait for a backoff
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
				closeCi()
				return ctx.Err()
			case <-boTimer.C:
				if vidOk {
					vh.RemoveValue(vid)
				}
			}
		}

		// construct controller (once)
		t1 := time.Now()
		if ci == nil {
			ci, lastErr = factory.Construct(
				ctx,
				config,
				controller.ConstructOpts{Logger: le},
			)
			if lastErr != nil {
				ci = nil
				continue
			}
			if ci == nil {
				err := errors.New("controller construct returned nil")
				le.Warn(err.Error())
				return err
			}
		}

		// emit the value
		vid, vidOk := vh.AddValue(NewExecControllerValue(
			t1,
			time.Time{},
			ci,
			nil,
		))

		// run execute
		le.Debug("starting controller")
		execErr := bus.ExecuteController(c.ctx, ci)

		le := le.WithField("exec-time", time.Since(t1).String())
		ctxCanceled := ctx.Err() != nil
		if execErr != nil && (!ctxCanceled || execErr != context.Canceled) {
			le.WithError(execErr).Warn("controller exited with error")
			lastErr = execErr
		} else {
			// controller was canceled or returned nil error
			le.Debug("controller exited normally")
		}

		// context was canceled, return now.
		if ctxCanceled {
			_ = vh.ClearValues()
			bus.RemoveController(ci)
			closeCi()
			return context.Canceled
		}

		// an error occurred, try again.
		if execErr != nil {
			continue
		}

		// controller Execute() is complete.
		// note: we need to take care to RemoveController later
		if vidOk {
			vh.AddValueRemovedCallback(vid, func() {
				bus.RemoveController(ci)
				closeCi()
			})
		} else {
			vh.AddResolverRemovedCallback(func() {
				bus.RemoveController(ci)
				closeCi()
			})
		}
		return nil
	}
}

// _ is a type assertion
var _ directive.Resolver = ((*resolver)(nil))
