package main

import (
	"errors"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/golang/protobuf/proto"
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
// Example: controllerbus/example/boilerplate/1
func (c *ToyControllerConfig) GetConfigID() string {
	return ControllerID
}

// EqualsConfig checks if the config is equal to another.
func (c *ToyControllerConfig) EqualsConfig(other config.Config) bool {
	return proto.Equal(c, other)
}
