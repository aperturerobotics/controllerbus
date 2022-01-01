package configset

import (
	"github.com/aperturerobotics/controllerbus/directive"
)

// LookupConfigSet looks up the status of ConfigSet controllers.
// Value type: LookupConfigSetValue
type LookupConfigSet interface {
	// Directive indicates LookupConfigSet is a directive.
	directive.Directive

	// GetLookupConfigSetControllerKeys returns the controller ids to lookup.
	// These are the keys in the configset map.
	GetLookupConfigSetControllerKeys() []string
}

// LookupConfigSetValue is the result type for LookupConfigSet.
// The value is removed and replaced when any values change.
type LookupConfigSetValue = State

// lookupConfigSet implements LookupConfigSet
type lookupConfigSet struct {
	controllerKeys []string
}

// NewLookupConfigSet constructs a new LookupConfigSet directive.
func NewLookupConfigSet(controllerKeys []string) LookupConfigSet {
	return &lookupConfigSet{
		controllerKeys: controllerKeys,
	}
}

// GetLookupConfigSetControllerKeys returns the configset id to lookup.
func (d *lookupConfigSet) GetLookupConfigSetControllerKeys() []string {
	return d.controllerKeys
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *lookupConfigSet) Validate() error {
	return nil
}

// GetValueOptions returns options relating to value handling.
func (d *lookupConfigSet) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{}
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (d *lookupConfigSet) IsEquivalent(other directive.Directive) bool {
	od, ok := other.(LookupConfigSet)
	if !ok {
		return false
	}

	dv := d.GetLookupConfigSetControllerKeys()
	ov := od.GetLookupConfigSetControllerKeys()
	if len(dv) != len(ov) {
		return false
	}
	for i, v1 := range dv {
		if ov[i] != v1 {
			return false
		}
	}
	return true
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *lookupConfigSet) Superceeds(other directive.Directive) bool {
	return false
}

// GetName returns the directive's type name.
// This is not necessarily unique, and is primarily intended for display.
func (d *lookupConfigSet) GetName() string {
	return "LookupConfigSet"
}

// GetDebugString returns the directive arguments stringified.
// This should be something like param1="test", param2="test".
// This is not necessarily unique, and is primarily intended for display.
func (d *lookupConfigSet) GetDebugVals() directive.DebugValues {
	vals := directive.DebugValues{}
	vals["controller-keys"] = d.GetLookupConfigSetControllerKeys()
	return vals
}

// _ is a type assertion
var _ LookupConfigSet = ((*lookupConfigSet)(nil))
