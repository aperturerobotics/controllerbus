package loader

import (
	"context"

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
	subCtx, subCtxCancel := context.WithCancel(ctx)
	dispose := func() {
		subCtxCancel()
		if disposeCb != nil {
			disposeCb()
		}
	}
	defer subCtxCancel()

	execValueCh := make(chan ExecControllerValue, 1)
	di, diRef, err := b.AddDirective(dir, bus.NewCallbackHandler(
		func(av directive.AttachedValue) {
			retVal, _ := av.GetValue().(ExecControllerValue)
			if retVal != nil {
				select {
				case <-execValueCh:
				default:
				}
				select {
				case execValueCh <- retVal:
				default:
				}
			}
		}, nil, dispose,
	))
	if err != nil {
		return nil, nil, nil, err
	}

	for {
		select {
		case <-subCtx.Done():
			diRef.Release()
			return nil, nil, nil, subCtx.Err()
		case val := <-execValueCh:
			if err := val.GetError(); err != nil {
				diRef.Release()
				return nil, nil, nil, err
			}
			if ctrl := val.GetController(); ctrl != nil {
				return ctrl, di, diRef, nil
			}
		}
	}
}
