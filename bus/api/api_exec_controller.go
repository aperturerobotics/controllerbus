package bus_api

import (
	"errors"

	ce "github.com/aperturerobotics/controllerbus/controller/exec"
)

// ErrExecControllerDisabled is returned if exec controller isn't enabled.
var ErrExecControllerDisabled = errors.New("exec controller is disabled on this api")

// ExecController executes a controller configuration on the bus.
func (a *API) ExecController(
	req *ce.ExecControllerRequest,
	server DRPCControllerBusService_ExecControllerStream,
) error {
	if !a.conf.GetEnableExecController() {
		return ErrExecControllerDisabled
	}
	return req.Execute(server.Context(), a.bus, true, server.Send)
}
