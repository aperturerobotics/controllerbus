package bus

import (
	"context"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
)

// BusController implements a base Controller with an attached Bus for directives.
// It has stubs for the functions implementing all required interfaces.
// Usually you should extend this type, overriding functions as needed.
type BusController[T config.Config] struct {
	le                *logrus.Entry
	bus               Bus
	conf              T
	controllerID      string
	controllerVersion semver.Version
	controllerDescrip string
}

// NewBusController constructs a new Controller with details.
func NewBusController[T config.Config](
	le *logrus.Entry,
	b Bus,
	conf T,
	controllerID string,
	controllerVersion semver.Version,
	controllerDescrip string,
) *BusController[T] {
	return &BusController[T]{
		le:                le,
		bus:               b,
		conf:              conf,
		controllerID:      controllerID,
		controllerVersion: controllerVersion,
		controllerDescrip: controllerDescrip,
	}
}

// NewBusControllerFactory builds a Factory for a type wrapping BusController.
func NewBusControllerFactory[T config.Config, C controller.Controller](
	b Bus,
	configID string,
	controllerID string,
	version semver.Version,
	controllerDescription string,
	newConfig func() T,
	newController func(base *BusController[T]) (C, error),
) *BusFactory[T, C] {
	return NewBusFactory(
		b,
		configID,
		version,
		newConfig,
		func(le *logrus.Entry, b Bus, conf T) (C, error) {
			base := NewBusController(le, b, conf, controllerID, version, controllerDescription)
			return newController(base)
		},
	)
}

// GetLogger returns the root logger.
func (c *BusController[T]) GetLogger() *logrus.Entry {
	return c.le
}

// GetConfig returns the config.
// Note: don't modify this object!
func (c *BusController[T]) GetConfig() T {
	return c.conf
}

// GetBus returns the bus.
func (c *BusController[T]) GetBus() Bus {
	return c.bus
}

// GetControllerInfo returns information about the controller.
func (c *BusController[T]) GetControllerInfo() *controller.Info {
	return controller.NewInfo(c.controllerID, c.controllerVersion, c.controllerDescrip)
}

// Execute executes the given controller.
func (c *BusController[T]) Execute(ctx context.Context) error {
	return nil
}

// HandleDirective asks if the handler can resolve the directive.
func (c *BusController[T]) HandleDirective(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
	return nil, nil
}

// Close releases any resources used by the controller.
func (c *BusController[T]) Close() error {
	return nil
}

// _ is a type assertion
var _ controller.Controller = (*BusController[*config.Placeholder])(nil)
