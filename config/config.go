package config

import (
	"encoding/json"

	protobuf_go_lite "github.com/aperturerobotics/protobuf-go-lite"
)

// Config is an object specifying configuration for a component of the system.
type Config interface {
	// Validate validates the configuration.
	// This is a cursory validation to see if the values "look correct."
	Validate() error
	// GetConfigID returns the unique string for this configuration type.
	// This string is stored with the encoded config.
	GetConfigID() string
	// EqualsConfig checks if the config is equal to another.
	EqualsConfig(other Config) bool

	// Marshaler is the json marshaler type.
	json.Marshaler
	// Unmarshaler is the json unmarshaler type.
	json.Unmarshaler

	// Message indicates this is a protobuf_go_lite message.
	protobuf_go_lite.Message
}

// Constructor constructs configuration objects.
type Constructor interface {
	// GetConfigID returns the unique string for this configuration type.
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

// EqualsConfig implements EqualsConfig assuming both Config are the same type and have VTEqual.
func EqualsConfig[T protobuf_go_lite.EqualVT[T]](t1 T, c2 Config) bool {
	t2, ok := c2.(T)
	if !ok {
		return false
	}
	return protobuf_go_lite.IsEqualVT(t1, t2)
}

// IsEqualVT compares two objects with VTEqual.
func IsEqualVT[T protobuf_go_lite.EqualVT[T]](c1, c2 T) bool {
	return protobuf_go_lite.IsEqualVT[T](c1, c2)
}

// MergeDebugVals merges multiple DebugValues into the first DebugValues passed.
func MergeDebugVals(to DebugValues, from ...DebugValues) {
	for _, f := range from {
		for k, v := range f {
			to[k] = append(to[k], v...)
		}
	}
}
