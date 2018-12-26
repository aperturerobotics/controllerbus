// Package directive declares the Directive type. The Directive is an
// instruction to all controllers attached to a bus indicating desired state.
// Directives are de-duplicated in the controller. Each directive is added with
// a listener, which receives events with links matching the directive.
package directive

import (
	"context"
	"time"
	// 	"github.com/golang/protobuf/proto"
)

// DebugValues maps string key to a list of values.
// It is used for debug visualizations.
type DebugValues map[string][]string

// NewDebugValues constructs a new DebugValues.
func NewDebugValues() DebugValues {
	return DebugValues{}
}

// ValueOptions are options related to value handling.
type ValueOptions struct {
	// MaxValueCount indicates a maximum number of values to retrieve.
	// The resolvers will be canceled when this many values are gathered.
	// If zero, accepts infinite values.
	MaxValueCount int

	// MaxValueHardCap indicates MaxValueCount is a hard cap. If it is not a
	// hard cap, any values found after resolvers are canceled is accepted. If
	// it is a hard cap, any values found after resolvers are canceled will be
	// rejected.
	MaxValueHardCap bool

	// UnrefDisposeDur is the duration to wait to dispose a directive after all
	// references have been released.
	UnrefDisposeDur time.Duration
}

// Directive implements a requested state (with a set of values).
type Directive interface {
	// Validate validates the directive.
	// This is a cursory validation to see if the values "look correct."
	Validate() error

	// GetValueOptions returns options relating to value handling.
	GetValueOptions() ValueOptions

	// IsEquivalent checks if the other directive is equivalent. If two
	// directives are equivalent, and the new directive does not superceed the
	// old, then the new directive will be merged (de-duplicated) into the old.
	IsEquivalent(other Directive) bool

	// Superceeds checks if the directive overrides another.
	// The other directive will be canceled if superceded.
	Superceeds(other Directive) bool

	// GetName returns the directive's type name.
	// This is not necessarily unique, and is primarily intended for display.
	GetName() string

	// GetDebugVals returns the directive arguments as key/value pairs.
	// This should be something like param1="test", param2="test".
	// This is not necessarily unique, and is primarily intended for display.
	GetDebugVals() DebugValues
}

// Controller manages directives.
type Controller interface {
	// AddDirective adds a directive to the controller.
	// This call de-duplicates equivalent directives.
	// cb receives values in order as they are emitted.
	// cb can be nil.
	// cb should not block.
	// Returns the instance, new reference, and any error.
	AddDirective(Directive, ReferenceHandler) (Instance, Reference, error)

	// AddHandler adds a directive handler.
	// The handler will receive calls for all existing directives (initial set).
	// If the handler returns an error for the initial set, will be returned.
	AddHandler(handler Handler) error

	// RemoveHandler removes a directive handler.
	RemoveHandler(Handler)

	// GetDirectives returns a list of all currently executing directives.
	GetDirectives() []Instance
}

// Reference is a reference to a directive.
// This is used to expire directive handles.
type Reference interface {
	// Release releases the reference.
	Release()
}

// ReferenceHandler handles values emitted by the directive instance.
type ReferenceHandler interface {
	// HandleValueAdded is called when a value is added to the directive.
	HandleValueAdded(Instance, AttachedValue)
	// HandleValueRemoved is called when a value is removed from the directive.
	HandleValueRemoved(Instance, AttachedValue)
	// HandleInstanceDisposed is called when a directive instance is disposed.
	// This will occur if Close() is called on the directive instance.
	HandleInstanceDisposed(Instance)
}

// Instance tracks a directive with reference counts and resolution state.
type Instance interface {
	// GetDirective returns the underlying directive object.
	GetDirective() Directive

	// AddReference adds a reference to the directive.
	// cb is called for each value.
	// cb calls should return immediately.
	// Will return nil if the directive is already expired.
	// If marked as a weak ref, the handler will not count towards the ref count.
	AddReference(cb ReferenceHandler, weakRef bool) Reference

	// AddDisposeCallback adds a callback that will be called when the instance
	// is disposed, either when Close() is called, or when the reference count
	// drops to zero. The callback may occur immediately if the instance is
	// already disposed, but will be made in a new goroutine.
	// Returns a callback release function.
	AddDisposeCallback(cb func()) func()

	// AddIdleCallback adds a callback that will be called when idle.
	// The callback is called exactly once.
	// Returns a callback release function.
	AddIdleCallback(cb func()) func()

	// Close cancels the directive instance.
	Close()
}

// Value satisfies a directive.
type Value interface{}

// AttachedValue is a value with some metadata.
type AttachedValue interface {
	// GetValueID returns the value ID.
	GetValueID() uint32
	// GetValue returns the value.
	GetValue() Value
}

// ResolverHandler handles values emitted by the resolver.
type ResolverHandler interface {
	// AddValue adds a value to the result, returning success and an ID. If
	// AddValue returns false, value was rejected. A rejected value should be
	// released immediately. If the value limit is reached, the value may not be
	// accepted. The value may be accepted, immediately before the resolver is
	// canceled (limit reached). It is always safe to call RemoveValue with the
	// ID at any time, even if the resolver is cancelled.
	AddValue(Value) (id uint32, accepted bool)
	// RemoveValue removes a value from the result, returning found.
	// It is safe to call this function even if the resolver is canceled.
	RemoveValue(id uint32) (val Value, found bool)
}

// Resolver resolves values for directives.
type Resolver interface {
	// Resolve resolves the values, emitting them to the handler.
	// The resolver may be canceled and restarted multiple times.
	// Any fatal error resolving the value is returned.
	// The resolver will not be retried after returning an error.
	// Values will be maintained from the previous call.
	Resolve(ctx context.Context, handler ResolverHandler) error
}

// Handler handles new reference instances.
type Handler interface {
	// HandleDirective asks if the handler can resolve the directive.
	// If it can, it returns a resolver. If not, returns nil.
	// Any exceptional errors are returned for logging.
	// It is safe to add a reference to the directive during this call.
	// The context passed is canceled when the directive instance expires.
	HandleDirective(context.Context, Instance) (Resolver, error)
}
