package main

import (
	"errors"
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
// Example: bifrost/transport/udp/1
func (c *ToyControllerConfig) GetConfigID() string {
	return "toy-controller/config/1"
}
