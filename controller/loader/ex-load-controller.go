package loader

import (
	"context"
	"errors"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// WaitExecControllerRunning executes any directive which yields
// ExecControllerValue and waits for either a error or success state before
// returning. Disposed is called if the state leaves RUNNING.
func WaitExecControllerRunning(
	ctx context.Context,
	b bus.Bus,
	dir directive.Directive,
	disposeCb func(),
) (controller.Controller, directive.Instance, directive.Reference, error) {
	execValueCh, di, diRef, err := bus.ExecOneOffWatchCh[ExecControllerValue](b, dir)
	if err != nil {
		return nil, nil, nil, err
	}

	dispose := func() {
		diRef.Release()
		if disposeCb != nil {
			disposeCb()
		}
	}
	defer dispose()

	for {
		select {
		case <-ctx.Done():
			return nil, nil, nil, context.Canceled
		case val := <-execValueCh:
			if val == nil {
				continue
			}
			if err := val.GetValue().GetError(); err != nil {
				return nil, nil, nil, err
			}
			if ctrl := val.GetValue().GetController(); ctrl != nil {
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
	var empty T
	ctrl, di, diRef, err := WaitExecControllerRunning(ctx, b, dir, disposeCb)
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
