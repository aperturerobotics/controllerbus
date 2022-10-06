package bus

import (
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// BusFactory implements Factory with an attached Bus for directives.
type BusFactory[T config.Config, C controller.Controller] struct {
	b             Bus
	configID      string
	version       semver.Version
	newConfig     func() T
	newController func(le *logrus.Entry, b Bus, conf T) (C, error)
}

// NewFactory constructs a new Factory with details.
func NewFactory[T config.Config, C controller.Controller](
	b Bus,
	configID string,
	version semver.Version,
	newConfig func() T,
	newController func(le *logrus.Entry, b Bus, conf T) (C, error),
) *BusFactory[T, C] {
	return &BusFactory[T, C]{
		b:             b,
		configID:      configID,
		version:       version,
		newConfig:     newConfig,
		newController: newController,
	}
}

// GetConfigID returns the unique config ID for the controller.
func (f *BusFactory[T, C]) GetConfigID() string {
	return f.configID
}

// ConstructConfig constructs an instance of the controller configuration.
func (f *BusFactory[T, C]) ConstructConfig() config.Config {
	return f.newConfig()
}

// Construct constructs the associated controller given configuration.
func (f *BusFactory[T, C]) Construct(cc config.Config, opts controller.ConstructOpts) (controller.Controller, error) {
	conf, ok := cc.(T)
	if !ok {
		expected := f.configID
		got := cc.GetConfigID()
		if expected == got {
			return nil, errors.New("config has correct config id but incorrect go type")
		}
		return nil, errors.Errorf("config has unexpected id: expected %s got %s", expected, got)
	}

	return f.newController(opts.Logger, f.b, conf)
}

// GetVersion returns the version of this controller.
func (f *BusFactory[T, C]) GetVersion() semver.Version {
	return f.version
}
