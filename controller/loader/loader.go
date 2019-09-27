package loader

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
)

// ControllerID is the controller identifier
const ControllerID = "controllerbus/loader/1"

// Version is the controller version
var Version = semver.MustParse("0.0.1")

// Controller implements the loader controller.
// It responds to ExecController directives and attaches to a bus.
type Controller struct {
	// le is the logger
	le *logrus.Entry
	// bus is the controller bus
	bus bus.Bus
}

// NewController builds a new loader controller given a bus.
func NewController(le *logrus.Entry, bus bus.Bus) (*Controller, error) {
	return &Controller{bus: bus, le: le}, nil
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() controller.Info {
	return controller.NewInfo(
		ControllerID,
		Version,
		"controller loader",
	)
}

// Execute executes the loader controller.
func (c *Controller) Execute(ctx context.Context) error {
	// No-op
	return nil
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
// The context passed is canceled when the directive instance expires.
func (c *Controller) HandleDirective(
	ctx context.Context,
	di directive.Instance,
) (directive.Resolver, error) {
	dir := di.GetDirective()
	if d, ok := dir.(ExecController); ok {
		return c.resolveExecController(ctx, d)
	}

	return nil, nil
}

// Close closes the controller.
func (c *Controller) Close() error {
	// TODO
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*Controller)(nil))
