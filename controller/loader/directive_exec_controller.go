package loader

import (
	"errors"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// ExecControllerSingleton is an ExecController directive.
// Will override or yield to exiting directives for the controller.
type ExecControllerSingleton struct {
	factory controller.Factory
	config  config.Config
}

// NewExecControllerSingleton constructs a new ExecController directive.
func NewExecControllerSingleton(
	factory controller.Factory,
	config config.Config,
) ExecController {
	return &ExecControllerSingleton{
		factory: factory,
		config:  config,
	}
}

// GetExecControllerFactory returns the factory desired to load.
func (d *ExecControllerSingleton) GetExecControllerFactory() controller.Factory {
	return d.factory
}

// GetExecControllerConfig returns the config to load the controller with.
func (d *ExecControllerSingleton) GetExecControllerConfig() config.Config {
	return d.config
}

// GetValueOptions returns options relating to value handling.
func (d *ExecControllerSingleton) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{
		// MaxValueCount:   1,
		// MaxValueHardCap: true,
	}
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *ExecControllerSingleton) Validate() error {
	if d.config == nil || d.factory == nil {
		return errors.New("config and factory cannot be nil")
	}

	if err := d.config.Validate(); err != nil {
		return err
	}

	return nil
}

// IsEquivalent checks if the other directive is equivalent.
// Ex: check if version range is inclusive of "other" version range.
func (d *ExecControllerSingleton) IsEquivalent(other directive.Directive) bool {
	otherExec, ok := other.(ExecController)
	if !ok {
		return false
	}

	f := d.GetExecControllerFactory()
	otherf := otherExec.GetExecControllerFactory()

	if otherf.GetConfigID() != f.GetConfigID() {
		return false
	}

	// This enforces that even with two factories, two controllers with the same
	// ID and configuration should not be started at the same time. The factory
	// with the greater version number "wins." If there are multiple sources for
	// code, the latest version is therefore taken.
	return d.GetExecControllerConfig().EqualsConfig(otherExec.GetExecControllerConfig())
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *ExecControllerSingleton) Superceeds(other directive.Directive) bool {
	otherExec, ok := other.(ExecController)
	if !ok {
		return false
	}

	f := d.GetExecControllerFactory()
	otherf := otherExec.GetExecControllerFactory()

	return f.GetVersion().GT(otherf.GetVersion())
}

// GetName returns the directive's type name.
// This is not necessarily unique, and is primarily intended for display.
func (d *ExecControllerSingleton) GetName() string {
	return "ExecController"
}

// GetDebugVals returns the directive arguments as k/v pairs.
// This is not necessarily unique, and is primarily intended for display.
func (d *ExecControllerSingleton) GetDebugVals() directive.DebugValues {
	vals := directive.NewDebugValues()
	vals["config-id"] = []string{d.GetExecControllerConfig().GetConfigID()}
	dbg, dbgOk := d.config.(config.Debuggable)
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

var (
	// _ is a type assertion
	_ directive.Directive = ((*ExecControllerSingleton)(nil))
	// _ is a type assertion
	_ directive.Debuggable = ((*ExecControllerSingleton)(nil))
	// _ is a type assertion
	_ ExecController = ((*ExecControllerSingleton)(nil))
)
