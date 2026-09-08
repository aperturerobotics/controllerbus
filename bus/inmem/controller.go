package inmem

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
)

// releaseWarningInterval bounds the interval between cleanup diagnostics.
const releaseWarningInterval = 30 * time.Second

// attachedCtrl contains the lifecycle of one attached controller instance.
type attachedCtrl struct {
	// ctrl owns execution and resource cleanup.
	ctrl controller.Controller
	// rel detaches directive handling.
	rel func()
	// detached excludes a closing controller from active lookup; guarded by Bus.mtx.
	detached bool

	// cancel stops execution.
	cancel context.CancelFunc
	// executeDone publishes executeErr after Execute returns.
	executeDone chan struct{}
	// executeErr is read only after executeDone closes.
	executeErr error
	// callback receives the completed lifecycle result.
	callback func(error)

	// finalizing selects the single finalizer.
	finalizing atomic.Bool
	// finalized publishes finalErr after all cleanup completes.
	finalized chan struct{}
	// finalErr is read only by the finalizer or after finalized closes.
	finalErr error
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
		finalized:   make(chan struct{}),
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
	if !c.finalizing.CompareAndSwap(false, true) {
		<-c.finalized
		return c.finalErr
	}
	defer close(c.finalized)
	defer b.forgetController(c)
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
