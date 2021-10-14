package bus_bridge

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/blang/semver"
)

// ControllerID is the controller identifier.
const ControllerID = "controllerbus/bus/bridge/1"

// Version is the API version.
var Version = semver.MustParse("0.0.1")

// BusBridge forwards directives to a bus.
type BusBridge struct {
	// target is the target bus
	target bus.Bus
}

// NewBusBridge constructs a new bus bridge.
func NewBusBridge(target bus.Bus) *BusBridge {
	return &BusBridge{
		target: target,
	}
}

// GetControllerInfo returns information about the controller.
func (b *BusBridge) GetControllerInfo() controller.Info {
	return controller.NewInfo(ControllerID, Version, "forwards directives to bus")
}

// Execute executes the given controller.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (b *BusBridge) Execute(ctx context.Context) error {
	return nil
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
// The context passed is canceled when the directive instance expires.
func (b *BusBridge) HandleDirective(ctx context.Context, di directive.Instance) (directive.Resolver, error) {
	if b.target == nil {
		return nil, nil
	}

	return NewBusBridgeResolver(b.target, di.GetDirective()), nil
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (b *BusBridge) Close() error {
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*BusBridge)(nil))
