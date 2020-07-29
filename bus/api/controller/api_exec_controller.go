package controllerbus_grpc_controller

import (
	"errors"

	ce "github.com/aperturerobotics/controllerbus/controller/exec"
	api "github.com/aperturerobotics/controllerbus/bus/api"
)

// ErrExecControllerDisabled is returned if exec controller isn't enabled.
var ErrExecControllerDisabled = errors.New("exec controller is disabled on this api")

// ExecController executes a controller configuration on the bus.
func (a *API) ExecController(
	req *ce.ExecControllerRequest,
	server api.ControllerBusService_ExecControllerServer,
) error {
	if !a.enableExecController {
		return ErrExecControllerDisabled
	}
	return req.Execute(server.Context(), a.bus, true, server.Send)
}
