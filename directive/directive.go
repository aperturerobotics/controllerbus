// Package directive declares the Directive type. The Directive is an
// instruction to all controllers attached to a bus indicating desired state.
// Directives are de-duplicated in the controller. Each directive is added with
// a listener, which receives events with links matching the directive.
package directive

import (
	"context"
	"github.com/golang/protobuf/proto"
)

// Directive implements a requested state (with a set of values).
type Directive interface {
	// Message requires Directive is a Protobuf message.
	// This allows easy transport / serialization.
	proto.Message

	// Validate validates the directive.
	// This is a cursory validation to see if the values "look correct."
	Validate() error

	// IsEquivalent checks if the other directive is equivalent.
	// Ex: check if version range is inclusive of "other" version range.
	IsEquivalent(other Directive) bool

	// Superceeds checks if the directive overrides another.
	// The other directive will be canceled if superceded.
	Superceeds() bool
}

// Controller manages directives.
type Controller interface {
	// AddDirective adds a directive to the controller.
	// This call de-duplicates equivalent directives.
	// Returns the instance, new reference, and any error.
	AddDirective(Directive) (Instance, Reference, error)
	// AddHandler adds a directive handler.
	// The handler will receive calls for all existing directives (initial set).
	AddHandler(Handler)
	// RemoveHandler removes a directive handler.
	RemoveHandler(Handler)
}

// Reference is a reference to a directive.
// This is used to expire directives with no references.
type Reference interface {
	// Resolve waits for directive values.
	// The final fatal error resolving the directive is returned.
	// If the context is canceled, the context error is returned.
	// Handler is called with every value, including values resolved before the call.
	Resolve(ctx context.Context, handler func(Value)) error
	// Release releases the reference.
	Release()
}

// Instance tracks a directive with reference counts and resolution state.
type Instance interface {
	// Directive is the underlying directive object.
	Directive

	// AddReference adds a reference to the directive.
	// Will return nil if the directive is already expired.
	AddReference() Reference
}

// Value satisfies a directive.
type Value interface{}

// Resolver resolves values for directives.
type Resolver interface {
	// Resolve resolves the values.
	// Any fatal error resolving the value is returned.
	// When the context is canceled valCh will not be drained anymore.
	Resolve(ctx context.Context, valCh chan<- Value) error
}

// Handler handles new reference instances.
type Handler interface {
	// HandleDirective asks if the handler can resolve the directive.
	// If it can, it returns a resolver. If not, returns nil.
	// Any exceptional errors are returned for logging.
	// It is safe to add a reference to the directive during this call.
	HandleDirective(Instance) (Resolver, error)
}
