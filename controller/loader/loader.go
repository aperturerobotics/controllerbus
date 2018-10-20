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
	<-ctx.Done()
	return nil
}

// HandleDirective asks if the handler can resolve the directive.
func (c *Controller) HandleDirective(
	di directive.Instance,
) (directive.Resolver, error) {
	dir := di.GetDirective()
	if d, ok := dir.(ExecController); ok {
		return c.resolveExecController(d)
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
