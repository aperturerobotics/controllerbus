package loader

import (
	"errors"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/cenkalti/backoff"
)

// execController implements the ExecController directive.
// Will override or yield to exiting directives for the controller.
type execController struct {
	factory      controller.Factory
	config       config.Config
	retryBackoff func() backoff.BackOff
	valueOpts    directive.ValueOptions
}

// NewExecController constructs a new ExecController directive.
func NewExecController(
	factory controller.Factory,
	config config.Config,
) ExecController {
	return &execController{
		factory: factory,
		config:  config,
	}
}

// NewExecControllerWithOpts constructs a new ExecController directive.
func NewExecControllerWithOpts(
	factory controller.Factory,
	config config.Config,
	retryBackoff func() backoff.BackOff,
	valueOpts directive.ValueOptions,
) ExecController {
	return &execController{
		factory:      factory,
		config:       config,
		retryBackoff: retryBackoff,
		valueOpts:    valueOpts,
	}
}

// GetExecControllerFactory returns the factory desired to load.
func (d *execController) GetExecControllerFactory() controller.Factory {
	return d.factory
}

// GetExecControllerConfig returns the config to load the controller with.
func (d *execController) GetExecControllerConfig() config.Config {
	return d.config
}

// GetExecControllerRetryBackoff returns the backoff to use for retries.
// If empty / nil, uses the default.
func (d *execController) GetExecControllerRetryBackoff() func() backoff.BackOff {
	return d.retryBackoff
}

// GetValueOptions returns options relating to value handling.
func (d *execController) GetValueOptions() directive.ValueOptions {
	return d.valueOpts
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *execController) Validate() error {
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
func (d *execController) IsEquivalent(other directive.Directive) bool {
	otherExec, ok := other.(ExecController)
	if !ok {
		return false
	}

	f := d.GetExecControllerFactory()
	otherf := otherExec.GetExecControllerFactory()

	if otherf.GetConfigID() != f.GetConfigID() {
		return false
	}

	// If the two configurations are identical, de-duplicate the directive.
	return d.GetExecControllerConfig().EqualsConfig(otherExec.GetExecControllerConfig())
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *execController) Superceeds(other directive.Directive) bool {
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
func (d *execController) GetName() string {
	return "ExecController"
}

// GetDebugVals returns the directive arguments as k/v pairs.
// This is not necessarily unique, and is primarily intended for display.
func (d *execController) GetDebugVals() directive.DebugValues {
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
	_ directive.Directive = ((*execController)(nil))
	// _ is a type assertion
	_ directive.Debuggable = ((*execController)(nil))
	// _ is a type assertion
	_ ExecController = ((*execController)(nil))
)
