package loader

import (
	"context"
	"errors"
	"sync"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// WaitExecControllerRunning executes any directive which yields
// ExecControllerValue and waits for either a error or success state before
// returning. Disposed is called once if the directive is disposed or the
// returned controller leaves RUNNING, such as when it exits with an error.
func WaitExecControllerRunning(
	ctx context.Context,
	b bus.Bus,
	dir directive.Directive,
	disposeCb func(),
) (controller.Controller, directive.Instance, directive.Reference, error) {
	return waitExecControllerRunning(ctx, b, dir, disposeCb, false)
}

// WaitExecControllerRunningRetry is WaitExecControllerRunning for callers that
// need the controller rather than its first outcome. A failed execution is
// retried by the loader after backoff, so the wait continues until the
// controller runs or ctx ends. Disposed is called as in WaitExecControllerRunning.
func WaitExecControllerRunningRetry(
	ctx context.Context,
	b bus.Bus,
	dir directive.Directive,
	disposeCb func(),
) (controller.Controller, directive.Instance, directive.Reference, error) {
	return waitExecControllerRunning(ctx, b, dir, disposeCb, true)
}

// waitExecControllerRunning waits for the directive's controller to run. An
// execution error ends the wait unless retry is set.
func waitExecControllerRunning(
	ctx context.Context,
	b bus.Bus,
	dir directive.Directive,
	disposeCb func(),
	retry bool,
) (controller.Controller, directive.Instance, directive.Reference, error) {
	subCtx, subCtxCancel := context.WithCancel(ctx)
	defer subCtxCancel()
	var disposeOnce sync.Once
	dispose := func() {
		subCtxCancel()
		if disposeCb != nil {
			disposeOnce.Do(disposeCb)
		}
	}

	// runningID identifies the current running value; returned marks it handed to the caller.
	var mtx sync.Mutex
	var runningID uint32
	var returned bool
	execValueCh := make(chan directive.AttachedValue, 1)
	di, diRef, err := b.AddDirective(dir, bus.NewCallbackHandler(
		func(av directive.AttachedValue) {
			retVal, _ := av.GetValue().(ExecControllerValue)
			if retVal == nil {
				return
			}
			if retVal.GetController() != nil {
				mtx.Lock()
				runningID = av.GetValueID()
				mtx.Unlock()
			}
			select {
			case <-execValueCh:
			default:
			}
			select {
			case execValueCh <- av:
			default:
			}
		},
		func(av directive.AttachedValue) {
			mtx.Lock()
			left := runningID != 0 && av.GetValueID() == runningID
			if left {
				runningID = 0
			}
			left = left && returned
			mtx.Unlock()
			if left {
				dispose()
			}
		},
		dispose,
	))
	if err != nil {
		return nil, nil, nil, err
	}

	for {
		select {
		case <-subCtx.Done():
			diRef.Release()
			return nil, nil, nil, subCtx.Err()
		case av := <-execValueCh:
			val := av.GetValue().(ExecControllerValue)
			if err := val.GetError(); err != nil && !retry {
				diRef.Release()
				return nil, nil, nil, err
			}
			ctrl := val.GetController()
			if ctrl == nil {
				continue
			}

			// A value removed before it is returned is skipped for the next one.
			mtx.Lock()
			current := av.GetValueID() == runningID
			returned = current
			mtx.Unlock()
			if current {
				return ctrl, di, diRef, nil
			}
		}
	}
}

// WaitExecControllerRunningTyped executes any directive which yields
// ExecControllerValue and waits for either a error or success state before
// returning. Disposed is called if the state leaves RUNNING.
//
// If the controller is not of type T an error will be returned and the
// reference will be released immediately.
func WaitExecControllerRunningTyped[T controller.Controller](
	ctx context.Context,
	b bus.Bus,
	dir directive.Directive,
	disposeCb func(),
) (T, directive.Instance, directive.Reference, error) {
	return typedExecController[T](WaitExecControllerRunning(ctx, b, dir, disposeCb))
}

// WaitExecControllerRunningRetryTyped is WaitExecControllerRunningRetry with
// the controller type check of WaitExecControllerRunningTyped.
func WaitExecControllerRunningRetryTyped[T controller.Controller](
	ctx context.Context,
	b bus.Bus,
	dir directive.Directive,
	disposeCb func(),
) (T, directive.Instance, directive.Reference, error) {
	return typedExecController[T](WaitExecControllerRunningRetry(ctx, b, dir, disposeCb))
}

// typedExecController asserts the controller type of a wait result, releasing
// the reference on a mismatch.
func typedExecController[T controller.Controller](
	ctrl controller.Controller,
	di directive.Instance,
	diRef directive.Reference,
	err error,
) (T, directive.Instance, directive.Reference, error) {
	var empty T
	if err != nil {
		return empty, di, diRef, err
	}
	ctrlt, ok := ctrl.(T)
	if !ok {
		diRef.Release()
		return empty, di, diRef, errors.New("exec controller constructed unexpected controller type")
	}
	return ctrlt, di, diRef, nil
}
