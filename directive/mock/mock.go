package directive_mock

import (
	"github.com/aperturerobotics/controllerbus/directive"
)

// MockDirective is a mock directive.
type MockDirective struct {
	ValidateErr error
	ValueOpts   directive.ValueOptions
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (m *MockDirective) Validate() error {
	return m.ValidateErr
}

func (m *MockDirective) GetValueOptions() directive.ValueOptions {
	return m.ValueOpts
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (m *MockDirective) IsEquivalent(other directive.Directive) bool {
	return false
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (m *MockDirective) Superceeds(other directive.Directive) bool {
	return false
}

// _ is a type assertion
var _ directive.Directive = ((*MockDirective)(nil))
