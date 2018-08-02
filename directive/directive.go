// Package directive declares the Directive type. The Directive is an
// instruction to all controllers attached to a bus indicating desired state.
// Directives are de-duplicated in the controller. Each directive is added with
// a listener, which receives events with links matching the directive.
package directive

import (
	"context"
	// 	"github.com/golang/protobuf/proto"
)

// Directive implements a requested state (with a set of values).
type Directive interface {
	// Validate validates the directive.
	// This is a cursory validation to see if the values "look correct."
	Validate() error
	// IsEquivalent checks if the other directive is equivalent. If two
	// directives are equivalent, and the new directive does not superceed the
	// old, then the new directive will be merged (de-duplicated) into the old.
	IsEquivalent(other Directive) bool
	// Superceeds checks if the directive overrides another.
	// The other directive will be canceled if superceded.
	Superceeds(other Directive) bool
}

// Controller manages directives.
type Controller interface {
	// AddDirective adds a directive to the controller.
	// This call de-duplicates equivalent directives.
	// cb receives values in order as they are emitted.
	// cb should not block.
	// Returns the instance, new reference, and any error.
	AddDirective(Directive, func(Value)) (Instance, Reference, error)
	// AddHandler adds a directive handler.
	// The handler will receive calls for all existing directives (initial set).
	// If the handler returns an error for the initial set, will be returned.
	AddHandler(Handler) error
	// RemoveHandler removes a directive handler.
	RemoveHandler(Handler)
}

// Reference is a reference to a directive.
// This is used to expire directive handles.
type Reference interface {
	// Release releases the reference.
	Release()
}

// Instance tracks a directive with reference counts and resolution state.
type Instance interface {
	// GetDirective returns the underlying directive object.
	GetDirective() Directive

	// AddReference adds a reference to the directive.
	// cb is called for each value in order.
	// cb should return immediately.
	// Will return nil if the directive is already expired.
	AddReference(cb func(Value)) Reference

	// Close cancels the directive instance.
	Close()
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
