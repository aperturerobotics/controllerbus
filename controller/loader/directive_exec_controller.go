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
		MaxValueCount:   1,
		MaxValueHardCap: true,
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

	// This is a singleton directive, so they are equiv if their IDs are the same.
	return otherf.GetControllerID() == f.GetControllerID()
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

// _ is a type assertion
var _ directive.Directive = ((*ExecControllerSingleton)(nil))

// _ is a type assertion
var _ ExecController = ((*ExecControllerSingleton)(nil))
