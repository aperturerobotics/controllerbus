package resolver

import (
	"errors"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/directive"
)

// LoadControllerWithConfigSingleton is an LoadControllerWithConfig directive.
// Will override or yield to exiting directives for the controller.
type LoadControllerWithConfigSingleton struct {
	config config.Config
}

// NewLoadControllerWithConfigSingleton constructs a new LoadControllerWithConfig directive.
func NewLoadControllerWithConfigSingleton(
	config config.Config,
) LoadControllerWithConfig {
	return &LoadControllerWithConfigSingleton{
		config: config,
	}
}

// GetDesiredControllerConfig returns the factory desired to load.
func (d *LoadControllerWithConfigSingleton) GetDesiredControllerConfig() config.Config {
	return d.config
}

// GetValueOptions returns options relating to value handling.
func (d *LoadControllerWithConfigSingleton) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{
		MaxValueCount:   1,
		MaxValueHardCap: true,
	}
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *LoadControllerWithConfigSingleton) Validate() error {
	if d.config == nil {
		return errors.New("config cannot be nil")
	}

	return nil
}

// IsEquivalent checks if the other directive is equivalent.
// Ex: check if version range is inclusive of "other" version range.
func (d *LoadControllerWithConfigSingleton) IsEquivalent(other directive.Directive) bool {
	otherExec, ok := other.(LoadControllerWithConfig)
	if !ok {
		return false
	}

	f := d.GetDesiredControllerConfig()
	return f.EqualsConfig(otherExec.GetDesiredControllerConfig())
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *LoadControllerWithConfigSingleton) Superceeds(other directive.Directive) bool {
	return false
}

// _ is a type assertion
var _ directive.Directive = ((*LoadControllerWithConfigSingleton)(nil))

// _ is a type assertion
var _ LoadControllerWithConfig = ((*LoadControllerWithConfigSingleton)(nil))
