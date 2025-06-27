package bus_api

import (
	"context"

	"github.com/aperturerobotics/controllerbus/directive"
)

// GetBusInfo requests information about the controller bus.
func (a *API) GetBusInfo(
	ctx context.Context,
	req *GetBusInfoRequest,
) (*GetBusInfoResponse, error) {
	/*
		if err := req.Validate(); err != nil {
			return nil, err
		}
	*/
	controllers := a.bus.GetControllers()
	directives := a.bus.GetDirectives()
	var resp GetBusInfoResponse
	for _, ctrl := range controllers {
		cinfo := ctrl.GetControllerInfo()
		resp.RunningControllers = append(resp.RunningControllers, cinfo)
	}
	for _, dir := range directives {
		resp.RunningDirectives = append(resp.RunningDirectives, directive.NewDirectiveState(dir))
	}

	return &resp, nil
}
