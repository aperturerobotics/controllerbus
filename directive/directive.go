// Package directive declares the Directive type. The Directive is an
// instruction to all controllers attached to a bus indicating desired state.
// Directives are de-duplicated in the controller. Each directive is added with
// a listener, which receives events with links matching the directive.
package directive

import (
	"context"
	"sort"
	"time"
)

// DebugValues maps string key to a list of values.
// It is used for debug visualizations.
type DebugValues map[string][]string

// NewDebugValues constructs a new DebugValues.
func NewDebugValues() DebugValues {
	return DebugValues{}
}

// NewProtoDebugValues constructs a new ProtoDebugValue set.
func NewProtoDebugValues(dv DebugValues) []*ProtoDebugValue {
	res := make([]*ProtoDebugValue, 0, len(dv))
	for k, v := range dv {
		res = append(res, &ProtoDebugValue{
			Key:    k,
			Values: v,
		})
	}
	sort.Slice(res, func(i int, j int) bool {
		return res[i].GetKey() < res[j].GetKey()
	})
	return res
}

// NewDirectiveInfo constructs a new DirectiveInfo from a directive.
func NewDirectiveInfo(dir Directive) *DirectiveInfo {
	var debugVals []*ProtoDebugValue
	debugDir, debugDirOk := dir.(Debuggable)
	if debugDirOk {
		debugVals = NewProtoDebugValues(debugDir.GetDebugVals())
	}

	return &DirectiveInfo{
		Name:      dir.GetName(),
		DebugVals: debugVals,
	}
}

// NewDirectiveState constructs a new DirectiveState from a running directive.
func NewDirectiveState(di Instance) *DirectiveState {
	return &DirectiveState{
		Info: NewDirectiveInfo(di.GetDirective()),
		// TODO state
	}
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

	// UnrefDisposeEmptyImmediate indicates we should immediately dispose a
	// directive that has become unreferenced if there are no associated Values
	// with the directive (it is unresolved) regardless of UnrefDisposeDur.
	UnrefDisposeEmptyImmediate bool
}

// Directive implements a requested state (with a set of values).
type Directive interface {
	// Validate validates the directive.
	// This is a cursory validation to see if the values "look correct."
	Validate() error

	// GetValueOptions returns options relating to value handling.
	GetValueOptions() ValueOptions

	// GetName returns the directives type name (i.e. DoSomething).
	// This is not intended to be unique and is primarily used for display.
	GetName() string
}

// DirectiveWithEquiv contains a check to see if it is equivalent to another directive.
type DirectiveWithEquiv interface {
	Directive

	// IsEquivalent checks if the other directive is equivalent. If two
	// directives are equivalent, and the new directive does not superceed the
	// old, then the new directive will be merged (de-duplicated) into the old.
	IsEquivalent(other Directive) bool
}

// DirectiveWithSuperceeds contains a check to see if the directive superceeds another.
type DirectiveWithSuperceeds interface {
	DirectiveWithEquiv

	// Superceeds checks if the directive overrides another.
	// The other directive will be canceled if superceded.
	Superceeds(other Directive) bool
}

// Debuggable indicates the directive implements the DebugVals interface.
type Debuggable interface {
	// GetDebugVals returns the directive arguments as key/value pairs.
	// This should be something like param1="test", param2="test".
	// This is not necessarily unique, and is primarily intended for display.
	GetDebugVals() DebugValues
}

// Networked is a directive which can be serialized and uniquely identified
// across IPC domains.
type Networked interface {
	// Directive indicates this is a directive.
	Directive
	// GetNetworkedCodec returns the encoder / decoder for this directive.
	// The same encoder/decoder should also be compatible with the results.
	GetNetworkedCodec() NetworkedCodec
}

// NetworkedCodec is the encoder/decoder for a networked directive.
type NetworkedCodec interface {
	// Marshal encodes the networked directive.
	Marshal(Networked) ([]byte, error)
	// Unmarshal decodes the data to the networked directive.
	// The type must match the expected type for the codec.
	Unmarshal([]byte, Networked) error
}

// DirectiveAdder can add a directive to a bus.
type DirectiveAdder interface {
	// AddDirective adds a directive to the controller.
	// This call de-duplicates equivalent directives.
	//
	// cb receives values in order as they are emitted.
	// cb can be nil.
	// cb should not block.
	//
	// Returns the instance, new reference, and any error.
	AddDirective(Directive, ReferenceHandler) (Instance, Reference, error)
}

// HandlerAdder can add a handler to a bus.
type HandlerAdder interface {
	// AddHandler adds a directive handler.
	// The handler will receive calls for all existing directives (initial set).
	// An error is returned only if adding the handler failed.
	// Returns a function to remove the handler.
	// The release function must be non-nil if err is nil, and nil if err != nil.
	AddHandler(handler Handler) (func(), error)
}

// DirectiveLister can list directives.
type DirectiveLister interface {
	// GetDirectives returns a list of all currently executing directives.
	GetDirectives() []Instance
}

// Controller manages running directives and handlers.
type Controller interface {
	// DirectiveLister has GetDirectives.
	DirectiveLister

	// DirectiveAdder has AddDirective.
	DirectiveAdder

	// HandlerAdder has AddHandler.
	HandlerAdder
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
	// Should not block.
	// Avoid calling directive functions in this routine.
	HandleValueAdded(Instance, AttachedValue)
	// HandleValueRemoved is called when a value is removed from the directive.
	// Should not block.
	// Avoid calling directive functions in this routine.
	HandleValueRemoved(Instance, AttachedValue)
	// HandleInstanceDisposed is called when a directive instance is disposed.
	// This will occur if Close() is called on the directive instance.
	// Avoid calling directive functions in this routine.
	HandleInstanceDisposed(Instance)
}

// IdleCallback is called when the directive becomes idle.
// Errs is the list of non-nil resolver errors.
type IdleCallback func(errs []error)

// Instance tracks a directive with reference counts and resolution state.
type Instance interface {
	// GetContext returns a context that is canceled when Instance is released.
	GetContext() context.Context

	// GetDirective returns the underlying directive object.
	GetDirective() Directive

	// GetDirectiveIdent returns a human-readable string identifying the directive.
	//
	// Ex: DoSomething or DoSomething<param=foo>
	GetDirectiveIdent() string

	// GetResolverErrors returns a snapshot of any errors returned by resolvers.
	GetResolverErrors() []error

	// AddReference adds a reference to the directive.
	// cb is called for each value.
	// cb calls should return immediately.
	// the release callback is called immediately if already released
	// If marked as a weak ref, the handler will not count towards the ref count.
	// will never return nil
	AddReference(cb ReferenceHandler, weakRef bool) Reference

	// AddDisposeCallback adds a callback that will be called when the instance
	// is disposed, either when Close() is called, or when the reference count
	// drops to zero. The callback may occur immediately if the instance is
	// already disposed, but will be made in a new goroutine.
	// Returns a callback release function.
	AddDisposeCallback(cb func()) func()

	// AddIdleCallback adds a callback that will be called when the directive becomes idle.
	// May be called multiple times if the directive is restarted.
	// Returns a callback release function.
	AddIdleCallback(cb IdleCallback) func()

	// CloseIfUnreferenced cancels the directive instance if there are no refs.
	//
	// This bypasses the unref dispose timer.
	// If inclWeakRefs=true, keeps the instance if there are any weak refs.
	// Returns if the directive instance was closed.
	CloseIfUnreferenced(inclWeakRefs bool) bool

	// Close cancels the directive instance and removes the directive.
	Close()
}

// Value satisfies a directive.
type Value interface{}

// ComparableValue is a type constraint for a comparable Value.
type ComparableValue interface {
	Value
	comparable
}

// AttachedValue is a value with some metadata.
type AttachedValue interface {
	// GetValueID returns the value ID.
	GetValueID() uint32
	// GetValue returns the value.
	GetValue() Value
}

// NewAttachedValue constructs a new typed attached value.
func NewAttachedValue(vid uint32, val Value) AttachedValue {
	return &attachedValue{
		vid: vid,
		val: val,
	}
}

// attachedValue implements TypedAttachedValue
type attachedValue struct {
	vid uint32
	val Value
}

// GetValueID returns the value ID.
func (t *attachedValue) GetValueID() uint32 {
	return t.vid
}

// GetValue returns the value.
func (t *attachedValue) GetValue() Value {
	return t.val
}

// _ is a type assertion
var _ AttachedValue = &attachedValue{}

// TypedAttachedValue is a typed value with some metadata.
type TypedAttachedValue[T ComparableValue] interface {
	// GetValueID returns the value ID.
	GetValueID() uint32
	// GetValue returns the value.
	GetValue() T
}

// NewTypedAttachedValue constructs a new typed attached value.
func NewTypedAttachedValue[T ComparableValue](vid uint32, val T) TypedAttachedValue[T] {
	return &typedAttachedValue[T]{
		vid: vid,
		val: val,
	}
}

// typedAttachedValue implements TypedAttachedValue
type typedAttachedValue[T ComparableValue] struct {
	vid uint32
	val T
}

// GetValueID returns the value ID.
func (t *typedAttachedValue[T]) GetValueID() uint32 {
	return t.vid
}

// GetValue returns the value.
func (t *typedAttachedValue[T]) GetValue() T {
	return t.val
}

// _ is a type assertion
var _ TypedAttachedValue[int] = &typedAttachedValue[int]{}

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
	// MarkIdle marks the resolver as idle.
	// If the resolver returns nil or an error, it's also marked as idle.
	MarkIdle()
	// CountValues returns the number of values that were set.
	// if allResolvers=false, returns the number set by this ResolverHandler.
	// if allResolvers=true, returns the number set by all resolvers.
	CountValues(allResolvers bool) int
	// ClearValues removes any values that were set by this ResolverHandler.
	// Returns list of value IDs that were removed.
	ClearValues() []uint32
	// AddValueRemovedCallback adds a callback that will be called when the
	// given value id is disposed or removed.
	//
	// The callback will be called if the value is removed for any reason,
	// including if the parent resolver, handler, or directive are removed.
	//
	// The callback might be called immediately if the value was already removed.
	//
	// Returns a release function to clear the callback early.
	AddValueRemovedCallback(id uint32, cb func()) func()
	// AddResolverRemovedCallback adds a callback that will be called when the
	// directive resolver is removed.
	//
	// The callback will be called if the resolver is removed for any reason,
	// including if the parent resolver, handler, or directive are removed.
	//
	// The callback might be called immediately if the resolver was already removed.
	//
	// Returns a release function to clear the callback early.
	AddResolverRemovedCallback(cb func()) func()
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
	// If it can, it returns resolver(s). If not, returns nil.
	// It is safe to add a reference to the directive during this call.
	// The passed context is canceled when the directive instance expires.
	// NOTE: the passed context is not canceled when the handler is removed.
	HandleDirective(context.Context, Instance) ([]Resolver, error)
}
