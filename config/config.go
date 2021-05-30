package config

import (
	"github.com/golang/protobuf/proto"
)

// Config is an object specifying configuration for a component of the system.
type Config interface {
	proto.Message

	// Validate validates the configuration.
	// This is a cursory validation to see if the values "look correct."
	Validate() error
	// GetConfigID returns the unique string for this configuration type.
	// This string is stored with the encoded config.
	// Example: controllerbus/example/boilerplate/1
	GetConfigID() string
	// EqualsConfig checks if the config is equal to another.
	EqualsConfig(other Config) bool
}

// Constructor constructs configuration objects.
type Constructor interface {
	// GetConfigID returns the unique string for this configuration type.
	// Example: controllerbus/example/boilerplate/1
	GetConfigID() string
	// ConstructConfig constructs a new configuration object.
	ConstructConfig() Config
}

// DebugValues maps string key to a list of values.
// It is used for debug visualizations.
type DebugValues map[string][]string

// Debuggable is a configuration with debug values.
//
// These debug values are sometimes combined with Exec or Load directives.
type Debuggable interface {
	// GetDebugVals returns the directive arguments as key/value pairs.
	// This should be something like param1="test", param2="test".
	// This is not necessarily unique, and is primarily intended for display.
	GetDebugVals() DebugValues
}
