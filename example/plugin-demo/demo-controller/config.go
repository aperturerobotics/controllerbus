package hot_compile_demo_controller

import (
	"time"

	"github.com/aperturerobotics/controllerbus/config"
)

// ConfigID is the string used to identify this config object.
const ConfigID = ControllerID

// Validate validates the configuration.
// This is a cursory validation to see if the values "look correct."
func (c *Config) Validate() error {
	if _, err := c.ParseExitAfterDur(); err != nil {
		return err
	}
	return nil
}

// GetConfigID returns the unique string for this configuration type.
// This string is stored with the encoded config.
func (c *Config) GetConfigID() string {
	return ConfigID
}

// EqualsConfig checks if the other config is equal.
func (c *Config) EqualsConfig(oc config.Config) bool {
	return config.EqualsConfig[*Config](c, oc)
}

// ParseExitAfterDur parses the exit after duration if set.
func (c *Config) ParseExitAfterDur() (time.Duration, error) {
	var dur time.Duration
	var err error
	if exitAfterDur := c.GetExitAfterDur(); exitAfterDur != "" {
		dur, err = time.ParseDuration(exitAfterDur)
	}
	return dur, err
}

// _ is a type assertion
var _ config.Config = ((*Config)(nil))
