//go:build !js && !wasm
// +build !js,!wasm

package plugin_loader_filesystem

import (
	"github.com/aperturerobotics/controllerbus/config"
)

// ConfigID is the identifier for the config type.
const ConfigID = ControllerID

// GetConfigID returns the config identifier.
func (c *Config) GetConfigID() string {
	return ConfigID
}

// EqualsConfig checks equality between two configs.
func (c *Config) EqualsConfig(c2 config.Config) bool {
	return config.EqualsConfig[*Config](c, c2)
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	return nil
}

// _ is a type assertion
var _ config.Config = ((*Config)(nil))
