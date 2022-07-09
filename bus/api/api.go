package bus_api

import (
	"github.com/aperturerobotics/controllerbus/bus"
	srpc "github.com/aperturerobotics/starpc/srpc"
)

// API implements the rpc API.
type API struct {
	bus  bus.Bus
	conf *Config
}

// NewAPI constructs a new instance of the API.
func NewAPI(bus bus.Bus, conf *Config) *API {
	return &API{bus: bus, conf: conf}
}

// RegisterAsSRPCServer registers the API to the SRPC mux.
func (a *API) RegisterAsSRPCServer(mux srpc.Mux) error {
	return SRPCRegisterControllerBusService(mux, a)
}

// _ is a type assertion
var _ SRPCControllerBusServiceServer = ((*API)(nil))
