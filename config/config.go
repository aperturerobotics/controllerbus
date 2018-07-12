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
	// Example: bifrost/transport/udp/1
	GetConfigID() string
}
