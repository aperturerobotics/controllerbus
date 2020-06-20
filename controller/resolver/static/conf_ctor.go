package static

import (
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
)

// ConfigCtor implements config.Constructor
type ConfigCtor struct {
	configID string
	f        controller.Factory
}

// NewConfigCtor builds a new configuration constructor.
func NewConfigCtor(configID string, f controller.Factory) config.Constructor {
	return &ConfigCtor{configID: configID, f: f}
}

// GetConfigID returns the unique string for this configuration type.
func (c *ConfigCtor) GetConfigID() string {
	return c.configID
}

// ConstructConfig constructs a new configuration object.
func (c *ConfigCtor) ConstructConfig() config.Config {
	return c.f.ConstructConfig()
}

// _ is a type assertion
var _ config.Constructor = ((*ConfigCtor)(nil))
