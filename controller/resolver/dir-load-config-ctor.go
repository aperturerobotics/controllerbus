package resolver

import (
	"errors"
	"time"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/directive"
)

// LoadConfigConstructorByID loads a controller configuration object constructor
// given a configuration ID. This resolves the configuration ID to a
// configuration object, which may involve loading controller code. The
// directive should be held open until the constructor is no longer needed, and
// the underlying resources have been referenced as needed by controller directives.
//
// Value type: config.Constructor
type LoadConfigConstructorByID interface {
	// Directive indicates this is a directive.
	directive.Directive

	// LoadConfigConstructorByIDConfigID is the configuration ID to use.
	LoadConfigConstructorByIDConfigID() string
}

// LoadConfigConstructorByIDValue is the value type for LoadConfigConstructorByID.
type LoadConfigConstructorByIDValue = config.Constructor

// loadConfigConstructorByID is an LoadConfigConstructorByID directive.
type loadConfigConstructorByID struct {
	configID string
}

// NewLoadConfigConstructorByID constructs a new LoadConfigConstructorByID directive.
func NewLoadConfigConstructorByID(
	configID string,
) LoadConfigConstructorByID {
	return &loadConfigConstructorByID{
		configID: configID,
	}
}

// LoadConfigConstructorByIDConfigID returns the desired config ID to match.
func (d *loadConfigConstructorByID) LoadConfigConstructorByIDConfigID() string {
	return d.configID
}

// GetValueOptions returns options relating to value handling.
func (d *loadConfigConstructorByID) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{
		MaxValueCount:   1,
		MaxValueHardCap: true,

		UnrefDisposeDur:            time.Millisecond * 500,
		UnrefDisposeEmptyImmediate: true,
	}
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *loadConfigConstructorByID) Validate() error {
	if d.configID == "" {
		return errors.New("config ID cannot be empty")
	}

	return nil
}

// IsEquivalent checks if the other directive is equivalent.
// Ex: check if version range is inclusive of "other" version range.
// This will never be true, as we want unique config objects.
func (d *loadConfigConstructorByID) IsEquivalent(other directive.Directive) bool {
	ot, ok := other.(LoadConfigConstructorByID)
	if !ok {
		return false
	}

	return ot.LoadConfigConstructorByIDConfigID() == d.LoadConfigConstructorByIDConfigID()
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *loadConfigConstructorByID) Superceeds(other directive.Directive) bool {
	return false
}

// GetName returns the directive's type name.
// This is not necessarily unique, and is primarily intended for display.
func (d *loadConfigConstructorByID) GetName() string {
	return "LoadConfigConstructorByID"
}

// GetDebugVals returns the directive arguments as k/v pairs.
// This is not necessarily unique, and is primarily intended for display.
func (d *loadConfigConstructorByID) GetDebugVals() directive.DebugValues {
	vals := directive.NewDebugValues()
	confID := d.LoadConfigConstructorByIDConfigID()
	vals["config-id"] = []string{confID}
	return vals
}

// _ is a type assertion
var _ LoadConfigConstructorByID = ((*loadConfigConstructorByID)(nil))
