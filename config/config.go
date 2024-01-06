package config

import (
	"github.com/aperturerobotics/util/vtcompare"
	"google.golang.org/protobuf/proto"
)

// Config is an object specifying configuration for a component of the system.
type Config interface {
	proto.Message

	// Validate validates the configuration.
	// This is a cursory validation to see if the values "look correct."
	Validate() error
	// GetConfigID returns the unique string for this configuration type.
	// This string is stored with the encoded config.
	GetConfigID() string
	// EqualsConfig checks if the config is equal to another.
	EqualsConfig(other Config) bool

	// MarshalVT marshals the config object with vtprotobuf.
	MarshalVT() ([]byte, error)
	// UnmarshalVT unmarshals the config object with vtprotobuf.
	UnmarshalVT(data []byte) error
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
func EqualsConfig[T vtcompare.EqualVT[T]](c1, c2 Config) bool {
	t1, _ := c1.(T)
	t2, _ := c2.(T)
	return vtcompare.IsEqualVT[T](t1, t2)
}

// IsEqualVT compares two objects with VTEqual.
func IsEqualVT[T vtcompare.EqualVT[T]](c1, c2 T) bool {
	return vtcompare.IsEqualVT[T](c1, c2)
}

// MergeDebugVals merges multiple DebugValues into the first DebugValues passed.
func MergeDebugVals(to DebugValues, from ...DebugValues) {
	for _, f := range from {
		for k, v := range f {
			to[k] = append(to[k], v...)
		}
	}
}
