package controller_exec

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/directive"
)

// ExecuteController executes a controller and calls the callback with state.
//
// Note: any ERROR state will return immediately after calling cb(). Some
// terminal error states may exit without calling cb().
func ExecuteController(
	ctx context.Context,
	cbus bus.Bus,
	conf config.Config,
	cb func(ControllerStatus),
) error {
	if cb == nil {
		cb = func(ControllerStatus) {}
	}

	subCtx, subCtxCancel := context.WithCancel(ctx)
	defer subCtxCancel()

	addedCh := make(chan resolver.LoadControllerWithConfigValue, 1)
	removedCh := make(chan resolver.LoadControllerWithConfigValue, 1)

	_, dirRef, err := cbus.AddDirective(
		resolver.NewLoadControllerWithConfig(conf),
		bus.NewCallbackHandler(
			// value added
			func(val directive.AttachedValue) {
				csVal, csValOk := val.GetValue().(resolver.LoadControllerWithConfigValue)
				if !csValOk || csVal == nil {
					return
				}
				select {
				case <-subCtx.Done():
					return
				case addedCh <- csVal:
				}
			},
			// value removed
			func(val directive.AttachedValue) {
				csVal, csValOk := val.GetValue().(resolver.LoadControllerWithConfigValue)
				if !csValOk || csVal == nil {
					return
				}
				select {
				case <-subCtx.Done():
					return
				case removedCh <- csVal:
				}
			},
			subCtxCancel,
		),
	)
	if err != nil {
		return err
	}
	defer dirRef.Release()

	var prevState ControllerStatus
	callCb := func(nextState ControllerStatus) {
		if nextState == prevState || cb == nil {
			return
		}
		prevState = nextState
		cb(nextState)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-subCtx.Done():
			return subCtx.Err()
		case csv := <-addedCh:
			ctrl := csv.GetController()
			csvErr := csv.GetError()
			if csvErr != nil {
				callCb(ControllerStatus_ControllerStatus_ERROR)
				return csvErr
			}
			if ctrl != nil {
				callCb(ControllerStatus_ControllerStatus_RUNNING)
			} else {
				callCb(ControllerStatus_ControllerStatus_CONFIGURING)
			}
		case csv := <-removedCh:
			// removed == value is no longer applicable.
			if csv.GetController() != nil && prevState == ControllerStatus_ControllerStatus_RUNNING {
				callCb(ControllerStatus_ControllerStatus_CONFIGURING)
			}
		}
	}
}
