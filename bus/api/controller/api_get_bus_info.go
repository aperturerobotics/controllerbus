package controllerbus_grpc_controller

import (
	"context"

	api "github.com/aperturerobotics/controllerbus/bus/api"
	"github.com/aperturerobotics/controllerbus/directive"
)

// GetBusInfo requests information about the controller bus.
func (a *API) GetBusInfo(
	ctx context.Context,
	req *api.GetBusInfoRequest,
) (*api.GetBusInfoResponse, error) {
	/*
		if err := req.Validate(); err != nil {
			return nil, err
		}
	*/
	controllers := a.bus.GetControllers()
	directives := a.bus.GetDirectives()
	var resp api.GetBusInfoResponse
	for _, ctrl := range controllers {
		cinfo := ctrl.GetControllerInfo()
		resp.RunningControllers = append(resp.RunningControllers, &cinfo)
	}
	for _, dir := range directives {
		resp.RunningDirectives = append(resp.RunningDirectives, directive.NewDirectiveState(dir))
	}
	return &resp, nil
}
