package boilerplate_controller

import (
	"errors"

	"github.com/aperturerobotics/controllerbus/config"
)

// ConfigID is the string used to identify this config object.
const ConfigID = ControllerID

// NewConfig constructs the config object.
func NewConfig() *Config {
	return &Config{}
}

// Validate validates the configuration.
// This is a cursory validation to see if the values "look correct."
func (c *Config) Validate() error {
	if c.GetExampleField() == "" {
		return errors.New("example field must be set")
	}
	return nil
}

// GetConfigID returns the unique string for this configuration type.
// This string is stored with the encoded config.
func (c *Config) GetConfigID() string {
	return ConfigID
}

// EqualsConfig checks if the other config is equal.
func (c *Config) EqualsConfig(other config.Config) bool {
	return config.EqualsConfig[*Config](c, other)
}

// _ is a type assertion
var _ config.Config = ((*Config)(nil))
