package boilerplate_controller

import (
	"context"
	"errors"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/controllerbus/example/boilerplate"

	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
)

// Version is the version of the controller implementation.
var Version = semver.MustParse("0.0.1")

// ControllerID is the ID of the controller.
const ControllerID = "controllerbus/example/boilerplate/1"

// Controller implements the boilerplate example controller.
type Controller struct {
	// le is the log entry
	le *logrus.Entry
	// bus is the controller bus
	bus bus.Bus
	// conf is the configuration
	conf *Config
}

// NewController constructs a new boilerplate example controller.
func NewController(
	le *logrus.Entry,
	bus bus.Bus,
	conf *Config,
) (*Controller, error) {
	return &Controller{
		le:   le,
		bus:  bus,
		conf: conf,
	}, nil
}

// Factory constructs Boilerplate controllers.
type Factory = bus.BusFactory[*Config, *Controller]

// NewFactory builds a boilerplate factory.
func NewFactory(b bus.Bus) *Factory {
	return bus.NewBusFactory(b, ConfigID, Version, NewConfig, NewController)
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() *controller.Info {
	return controller.NewInfo(ControllerID, Version, "boilerplate example")
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
// The context passed is canceled when the directive instance expires.
func (c *Controller) HandleDirective(
	ctx context.Context,
	inst directive.Instance,
) ([]directive.Resolver, error) {
	dir := inst.GetDirective()
	switch d := dir.(type) {
	case boilerplate.Boilerplate:
		return c.resolveBoilerplate(ctx, inst, d)
	}

	return nil, nil
}

// Execute executes the given controller.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *Controller) Execute(ctx context.Context) error {
	if errStr := c.conf.GetFailWithErr(); errStr != "" {
		err := errors.New(errStr)
		c.le.
			WithError(err).
			Warn("boilerplate controller returning configured error")
		return err
	}

	c.le.Infof(
		"hello from boilerplate controller: %s",
		c.conf.GetExampleField(),
	)
	return nil
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (c *Controller) Close() error {
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*Controller)(nil))
