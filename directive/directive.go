// Package directive declares the Directive type. The Directive is an
// instruction to a node controller of desired connectivity. Directives are
// de-duplicated in the controller. Each directive is added with a listener,
// which receives events with links matching the directive.
package directive

import (
	"github.com/golang/protobuf/proto"
)

// Handler handles new directives.
type Handler interface {
	// AddDirective handles a new directive.
	AddDirective(Directive)
}

// Directive implements a request for connectivity.
type Directive interface {
	// Message requires Directive is a Protobuf message.
	// This allows easy transport / serialization.
	proto.Message

	// Validate validates the directive.
	// This is a cursory validation to see if the values "look correct."
	Validate() error

	// IsEquivalent checks if the other directive is equivalent.
	IsEquivalent(other Directive) bool
}
