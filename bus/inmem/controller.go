package inmem

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
)

const releaseWarningInterval = 30 * time.Second

// attachedCtrl contains the lifecycle of one attached controller instance.
type attachedCtrl struct {
	ctrl controller.Controller
	rel  func()

	cancel      context.CancelFunc
	executeDone chan struct{}
	executeErr  error
	callback    func(error)

	finalizeOnce sync.Once
	finalErr     error
}

// newAttachedCtrl constructs an attached controller lifecycle.
func newAttachedCtrl(
	ctrl controller.Controller,
	rel func(),
	cancel context.CancelFunc,
	callback func(error),
) *attachedCtrl {
	return &attachedCtrl{
		ctrl:        ctrl,
		rel:         rel,
		cancel:      cancel,
		executeDone: make(chan struct{}),
		callback:    callback,
	}
}

// finishExecution records the result before publishing execution completion.
func (c *attachedCtrl) finishExecution(err error) {
	c.executeErr = err
	close(c.executeDone)
}

// finalize cancels, detaches, waits, closes, and reports exactly once.
func (c *attachedCtrl) finalize(b *Bus) error {
	c.finalizeOnce.Do(func() {
		c.cancel()
		b.detachController(c)

		ticker := time.NewTicker(releaseWarningInterval)
		defer ticker.Stop()
	waitForExecute:
		for {
			select {
			case <-c.executeDone:
				break waitForExecute
			case <-ticker.C:
				b.le.WithField("controller", c.ctrl).Warn("waiting for controller Execute to return")
			}
		}

		c.finalErr = joinControllerErrors(c.executeErr, c.ctrl.Close())
		if c.callback != nil {
			c.callback(c.finalErr)
		}
	})
	return c.finalErr
}

// joinControllerErrors preserves the execution error and identifies a close
// failure for errors.As.
func joinControllerErrors(executeErr, closeErr error) error {
	if closeErr == nil {
		return executeErr
	}
	wrappedCloseErr := &bus.ControllerCloseError{Err: closeErr}
	if executeErr == nil {
		return wrappedCloseErr
	}
	return errors.Join(executeErr, wrappedCloseErr)
}
