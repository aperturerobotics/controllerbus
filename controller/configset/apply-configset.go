package configset

import (
	"sort"
	"strconv"
	"strings"

	"github.com/aperturerobotics/controllerbus/directive"
)

// ApplyConfigSet is a directive to apply a controllerbus config set.
// Value type: ApplyConfigSetValue
type ApplyConfigSet interface {
	// Directive indicates ApplyConfigSet is a directive.
	directive.Directive

	// GetApplyConfigSet returns the configset to apply.
	GetApplyConfigSet() ConfigSet
}

// ApplyConfigSetValue is the result type for ApplyConfigSet.
// The value is removed and replaced when any values change.
type ApplyConfigSetValue = State

// applyConfigSet implements ApplyConfigSet
type applyConfigSet struct {
	configSet ConfigSet
}

// NewApplyConfigSet constructs a new ApplyConfigSet directive.
func NewApplyConfigSet(configSet ConfigSet) ApplyConfigSet {
	return &applyConfigSet{
		configSet: configSet,
	}
}

// GetApplyConfigSet returns the configset to apply.
func (d *applyConfigSet) GetApplyConfigSet() ConfigSet {
	return d.configSet
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *applyConfigSet) Validate() error {
	return nil
}

// GetValueOptions returns options relating to value handling.
func (d *applyConfigSet) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{}
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (d *applyConfigSet) IsEquivalent(other directive.Directive) bool {
	od, ok := other.(ApplyConfigSet)
	if !ok {
		return false
	}

	return d.GetApplyConfigSet().Equal(od.GetApplyConfigSet())
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *applyConfigSet) Superceeds(other directive.Directive) bool {
	return false
}

// GetName returns the directive's type name.
// This is not necessarily unique, and is primarily intended for display.
func (d *applyConfigSet) GetName() string {
	return "ApplyConfigSet"
}

// GetDebugString returns the directive arguments stringified.
// This should be something like param1="test", param2="test".
// This is not necessarily unique, and is primarily intended for display.
func (d *applyConfigSet) GetDebugVals() directive.DebugValues {
	vals := directive.DebugValues{}
	confIDs := make([]string, len(d.GetApplyConfigSet()))
	if len(confIDs) != 0 {
		i := 0
		for k, v := range d.GetApplyConfigSet() {
			confIDs[i] = strings.Join([]string{
				k,
				"@",
				strconv.FormatUint(v.GetRev(), 10),
			}, "")
			i++
		}

		sort.Strings(confIDs)
		vals["controller-keys"] = confIDs
	}
	return vals
}

// _ is a type assertion
var _ ApplyConfigSet = ((*applyConfigSet)(nil))
