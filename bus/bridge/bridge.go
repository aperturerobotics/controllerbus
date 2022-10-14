package bus_bridge

import (
	"context"
	"sync"

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
	// filterFn is the filter function
	filterFn FilterFn

	// mtx guards below fields
	mtx sync.Mutex
	// seenDi is the list of seen ongoing directive instances
	seenDi map[directive.Directive]struct{}
}

// FilterFn filters directive instances for the bus bridge.
type FilterFn = func(di directive.Instance) (bool, error)

// NewBusBridge constructs a new bus bridge.
func NewBusBridge(target bus.Bus, filterFn FilterFn) *BusBridge {
	return &BusBridge{
		target:   target,
		filterFn: filterFn,
		seenDi:   make(map[directive.Directive]struct{}),
	}
}

// SetDirectiveBridgeTarget sets the target bus.
// must be called before HandleDirective and Execute (just after construct).
func (b *BusBridge) SetDirectiveBridgeTarget(target bus.Bus) {
	b.target = target
}

// GetControllerInfo returns information about the controller.
func (b *BusBridge) GetControllerInfo() *controller.Info {
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
func (b *BusBridge) HandleDirective(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
	if b.target == nil {
		return nil, nil
	}

	if b.filterFn != nil {
		process, err := b.filterFn(di)
		if err != nil || !process {
			return nil, err
		}
	}

	dir := di.GetDirective()
	b.mtx.Lock()
	_, seen := b.seenDi[dir]
	if !seen {
		b.seenDi[dir] = struct{}{}
	}
	b.mtx.Unlock()

	// avoid infinite loop
	if seen {
		return nil, nil
	}

	// add callback to remove when necessary
	// use separate goroutine to avoid mutex contention
	go di.AddDisposeCallback(func() {
		b.mtx.Lock()
		delete(b.seenDi, dir)
		b.mtx.Unlock()
	})

	return directive.Resolvers(NewBusBridgeResolver(b.target, dir)), nil
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (b *BusBridge) Close() error {
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*BusBridge)(nil))
