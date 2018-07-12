package loader

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// Controller implements the loader controller.
// It responds to LoadController directives and attaches to a bus.
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

// Execute executes the loader controller.
func (c *Controller) Execute(ctx context.Context) error {
	// TODO
	return nil
}

// HandleDirective asks if the handler can resolve the directive.
func (c *Controller) HandleDirective(
	dir directive.Instance,
) (directive.Resolver, error) {
	switch d := dir.(type) {
	case LoadController:
		return c.resolveLoadController(d)
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
