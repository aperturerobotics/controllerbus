package boilerplate_controller

import (
	"github.com/aperturerobotics/controllerbus/bus"
)

// Factory constructs Boilerplate controllers.
type Factory = bus.BusFactory[*Config, *Controller]

// NewFactory builds a boilerplate factory.
func NewFactory(b bus.Bus) *Factory {
	return bus.NewFactory(b, ConfigID, Version, NewConfig, NewController)
}
