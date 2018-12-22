package controller_exec

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
)

// ExecuteController executes a controller and calls the callback with state.
func ExecuteController(
	ctx context.Context,
	cbus bus.Bus,
	conf config.Config,
	cb func(ControllerStatus),
) error {
	if cb == nil {
		cb = func(ControllerStatus) {}
	}
	dir := resolver.NewLoadControllerWithConfig(conf)

	cb(ControllerStatus_ControllerStatus_CONFIGURING)
	_, valRef, err := bus.ExecOneOff(ctx, cbus, dir, nil)
	if err != nil {
		return err
	}
	defer valRef.Release()

	cb(ControllerStatus_ControllerStatus_RUNNING)
	<-ctx.Done()
	return nil
}
