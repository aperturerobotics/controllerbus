package main

import (
	"errors"

	"github.com/aperturerobotics/controllerbus/config"
)

// Validate validates the configuration.
// This is a cursory validation to see if the values "look correct."
func (c *ToyControllerConfig) Validate() error {
	if c.GetName() == "" {
		return errors.New("name cannot be empty")
	}

	return nil
}

// GetConfigID returns the unique string for this configuration type.
// This string is stored with the encoded config.
func (c *ToyControllerConfig) GetConfigID() string {
	return ControllerID
}

// EqualsConfig checks if the config is equal to another.
func (c *ToyControllerConfig) EqualsConfig(other config.Config) bool {
	return config.EqualsConfig[*ToyControllerConfig](c, other)
}
