package directive_mock

import (
	"github.com/aperturerobotics/controllerbus/directive"
)

// MockDirective is a mock directive.
type MockDirective struct {
	ValidateErr error
	ValueOpts   directive.ValueOptions
}

// GetName returns the directive name.
func (m *MockDirective) GetName() string {
	return "Mock"
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (m *MockDirective) Validate() error {
	return m.ValidateErr
}

func (m *MockDirective) GetValueOptions() directive.ValueOptions {
	return m.ValueOpts
}

// GetDebugVals returns the directive arguments as key/value pairs.
// This should be something like param1="test", param2="test".
// This is not necessarily unique, and is primarily intended for display.
func (m *MockDirective) GetDebugVals() directive.DebugValues {
	return directive.DebugValues{}
}

// _ is a type assertion
var _ directive.Directive = ((*MockDirective)(nil))
