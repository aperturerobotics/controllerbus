package resolver

import (
	"errors"

	"github.com/aperturerobotics/controllerbus/config"
	loader "github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/directive"
)

// LoadControllerWithConfig is a directive indicating a controller should be
// loaded and executed (in one action) given a configuration.
type LoadControllerWithConfig interface {
	// Directive indicates this is a directive.
	directive.Directive

	// GetDesiredControllerConfig returns the desired controller config.
	GetDesiredControllerConfig() config.Config
}

// LoadControllerWithConfig is the value emitted to satisfy the
// LoadControllerWithConfig directive.
type LoadControllerWithConfigValue = loader.ExecControllerValue

// loadControllerWithConfig is an LoadControllerWithConfig directive.
// Will override or yield to exiting directives for the controller.
type loadControllerWithConfig struct {
	config    config.Config
	valueOpts directive.ValueOptions
}

// NewLoadControllerWithConfig constructs a new LoadControllerWithConfig directive.
func NewLoadControllerWithConfig(
	config config.Config,
) LoadControllerWithConfig {
	return &loadControllerWithConfig{
		config: config,
	}
}

// NewLoadControllerWithConfigAndValueOpts constructs a new LoadControllerWithConfig directive.
func NewLoadControllerWithConfigAndValueOpts(
	config config.Config,
	valueOpts directive.ValueOptions,
) LoadControllerWithConfig {
	return &loadControllerWithConfig{
		config:    config,
		valueOpts: valueOpts,
	}
}

// GetDesiredControllerConfig returns the factory desired to load.
func (d *loadControllerWithConfig) GetDesiredControllerConfig() config.Config {
	return d.config
}

// GetValueOptions returns options relating to value handling.
func (d *loadControllerWithConfig) GetValueOptions() directive.ValueOptions {
	return d.valueOpts
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *loadControllerWithConfig) Validate() error {
	if d.config == nil {
		return errors.New("config cannot be nil")
	}

	return nil
}

// IsEquivalent checks if the other directive is equivalent.
// Ex: check if version range is inclusive of "other" version range.
func (d *loadControllerWithConfig) IsEquivalent(other directive.Directive) bool {
	otherExec, ok := other.(LoadControllerWithConfig)
	if !ok {
		return false
	}

	f := d.GetDesiredControllerConfig()
	return f.EqualsConfig(otherExec.GetDesiredControllerConfig())
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *loadControllerWithConfig) Superceeds(other directive.Directive) bool {
	return false
}

// GetName returns the directive's type name.
// This is not necessarily unique, and is primarily intended for display.
func (d *loadControllerWithConfig) GetName() string {
	return "LoadControllerWithConfig"
}

// GetDebugVals returns the directive arguments as k/v pairs.
// This is not necessarily unique, and is primarily intended for display.
func (d *loadControllerWithConfig) GetDebugVals() directive.DebugValues {
	vals := directive.NewDebugValues()
	confID := d.GetDesiredControllerConfig().GetConfigID()
	vals["config-id"] = []string{confID}
	dbg, dbgOk := d.GetDesiredControllerConfig().(config.Debuggable)
	if dbgOk {
		dbgVals := dbg.GetDebugVals()
		if dbgVals != nil {
			for k, v := range dbg.GetDebugVals() {
				if _, ok := vals[k]; !ok {
					vals[k] = v
				}
			}
		}
	}
	return vals
}

// _ is a type assertion
var _ LoadControllerWithConfig = ((*loadControllerWithConfig)(nil))

// _ is a type assertion
var _ directive.Debuggable = ((*loadControllerWithConfig)(nil))
