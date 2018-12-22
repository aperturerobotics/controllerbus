package static

import (
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
)

// configCtor implements config.Constructor
type configCtor struct {
	configID string
	f        controller.Factory
}

// newConfigCtor builds a new configuration constructor.
func newConfigCtor(configID string, f controller.Factory) config.Constructor {
	return &configCtor{configID: configID, f: f}
}

// GetConfigID returns the unique string for this configuration type.
func (c *configCtor) GetConfigID() string {
	return c.configID
}

// ConstructConfig constructs a new configuration object.
func (c *configCtor) ConstructConfig() config.Config {
	return c.f.ConstructConfig()
}

// _ is a type assertion
var _ config.Constructor = ((*configCtor)(nil))
