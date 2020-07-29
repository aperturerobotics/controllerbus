package controllerbus_grpc_controller

import (
	"github.com/aperturerobotics/controllerbus/bus"
	api "github.com/aperturerobotics/controllerbus/bus/api"
	"google.golang.org/grpc"
)

// API implements the GRPC API.
type API struct {
	bus                  bus.Bus
	enableExecController bool
}

// NewAPI constructs a new instance of the API.
func NewAPI(bus bus.Bus, enableExecController bool) (*API, error) {
	return &API{bus: bus, enableExecController: enableExecController}, nil
}

// RegisterAsGRPCServer registers the API to the GRPC instance.
func (a *API) RegisterAsGRPCServer(grpcServer *grpc.Server) {
	api.RegisterControllerBusServiceServer(grpcServer, a)
}

// _ is a type assertion
var _ api.ControllerBusServiceServer = ((*API)(nil))
