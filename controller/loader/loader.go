package loader

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// ControllerID is the controller identifier.
const ControllerID = "controllerbus/loader"

// Version is the controller version.
var Version = controller.MustParseVersion("0.0.1")

// Controller implements the loader controller.
// It responds to ExecController directives and attaches to a bus.
type Controller struct {
	// le is the logger.
	le *logrus.Entry
	// bus runs the loaded controllers.
	bus bus.Bus
}

// NewController builds a new loader controller given a bus.
func NewController(le *logrus.Entry, bus bus.Bus) (*Controller, error) {
	return &Controller{bus: bus, le: le}, nil
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() *controller.Info {
	return controller.NewInfo(
		ControllerID,
		Version,
		"controller loader",
	)
}

// Execute returns immediately; the loader works through HandleDirective.
func (c *Controller) Execute(ctx context.Context) error {
	return nil
}

// HandleDirective resolves ExecController directives. The context is
// canceled when the directive instance expires.
func (c *Controller) HandleDirective(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
	if d, ok := di.GetDirective().(ExecController); ok {
		return c.resolveExecController(ctx, d)
	}
	return nil, nil
}

// Close releases nothing; each loaded controller is released with its directive.
func (c *Controller) Close() error {
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*Controller)(nil))
