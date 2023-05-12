package resolver

import (
	"context"
	"errors"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// LoadFactoryByConfig loads a controller factory given a configuration object.
// This resolves the configuration object to a controller constructor, which may
// involve loading controller code. The directive should be held open until the
// factory is no longer needed, and the underlying resources have been
// referenced as needed by controller directives.
type LoadFactoryByConfig interface {
	// Directive indicates this is a directive.
	directive.Directive

	// LoadFactoryByConfig is the configuration to load a factory for.
	LoadFactoryByConfig() config.Config
}

// LoadFactoryByConfigValue is the value type for LoadFactoryByConfig.
type LoadFactoryByConfigValue = controller.Factory

// ExLoadFactoryByConfig executes the LoadFactoryByConfig directive.
// the directive should be released after the Factory is no longer needed.
func ExLoadFactoryByConfig(
	ctx context.Context,
	b bus.Bus,
	conf config.Config,
) (controller.Factory, directive.Reference, error) {
	av, _, avRef, err := bus.ExecOneOff(ctx, b, NewLoadFactoryByConfig(conf), nil, nil)
	if err != nil {
		return nil, nil, err
	}
	v, vOk := av.GetValue().(controller.Factory)
	if !vOk {
		avRef.Release()
		return nil, nil, errors.New("load factory by config returned unexpected type")
	}
	return v, avRef, nil
}

// loadFactoryByConfig is an LoadFactoryByConfig directive.
type loadFactoryByConfig struct {
	loadConfig config.Config
}

// NewLoadFactoryByConfig constructs a new LoadFactoryByConfig directive.
func NewLoadFactoryByConfig(
	loadConfig config.Config,
) LoadFactoryByConfig {
	return &loadFactoryByConfig{
		loadConfig: loadConfig,
	}
}

// LoadFactoryByConfig is the configuration to load.
func (d *loadFactoryByConfig) LoadFactoryByConfig() config.Config {
	return d.loadConfig
}

// GetValueOptions returns options relating to value handling.
func (d *loadFactoryByConfig) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{
		MaxValueCount:   1,
		MaxValueHardCap: true,
	}
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *loadFactoryByConfig) Validate() error {
	if d.loadConfig == nil {
		return errors.New("config cannot be nil")
	}

	return nil
}

// IsEquivalent checks if the other directive is equivalent.
// Ex: check if version range is inclusive of "other" version range.
// This will never be true, as we want unique config objects.
func (d *loadFactoryByConfig) IsEquivalent(other directive.Directive) bool {
	ot, ok := other.(LoadFactoryByConfig)
	if !ok {
		return false
	}

	dConf := d.loadConfig
	otConf := ot.LoadFactoryByConfig()
	return dConf.EqualsConfig(otConf)
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *loadFactoryByConfig) Superceeds(other directive.Directive) bool {
	return false
}

// GetName returns the directive's type name.
// This is not necessarily unique, and is primarily intended for display.
func (d *loadFactoryByConfig) GetName() string {
	return "LoadFactoryByConfig"
}

// GetDebugVals returns the directive arguments as k/v pairs.
// This is not necessarily unique, and is primarily intended for display.
func (d *loadFactoryByConfig) GetDebugVals() directive.DebugValues {
	vals := directive.NewDebugValues()
	conf := d.LoadFactoryByConfig()
	if conf != nil {
		vals["config-id"] = []string{conf.GetConfigID()}
	}
	return vals
}

// _ is a type assertion
var _ LoadFactoryByConfig = ((*loadFactoryByConfig)(nil))
