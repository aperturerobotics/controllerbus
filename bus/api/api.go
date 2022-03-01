package bus_api

import (
	"github.com/aperturerobotics/controllerbus/bus"
	"storj.io/drpc"
)

// API implements the GRPC API.
type API struct {
	bus  bus.Bus
	conf *Config
}

// NewAPI constructs a new instance of the API.
func NewAPI(bus bus.Bus, conf *Config) *API {
	return &API{bus: bus, conf: conf}
}

// RegisterAsDRPCServer registers the API to the DRPC mux.
func (a *API) RegisterAsDRPCServer(mux drpc.Mux) {
	DRPCRegisterControllerBusService(mux, a)
}

// _ is a type assertion
var _ DRPCControllerBusServiceServer = ((*API)(nil))
