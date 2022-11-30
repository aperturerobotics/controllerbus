package core

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/bus/inmem"
	"github.com/aperturerobotics/controllerbus/controller"
	configset_controller "github.com/aperturerobotics/controllerbus/controller/configset/controller"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/controller/resolver/static"
	cdc "github.com/aperturerobotics/controllerbus/directive/controller"
	"github.com/sirupsen/logrus"
)

// FactoryResolverCtor constructs a Factory resolver.
type FactoryResolverCtor func(b bus.Bus, sr *static.Resolver) (controller.FactoryResolver, error)

// CoreBusConfig configures NewCoreBus.
type CoreBusConfig struct {
	// FactoryResolver overrides the static resolver.
	FactoryResolverCtor FactoryResolverCtor
	// BuiltInFactories is the list of built in controller factories.
	BuiltInFactories []controller.Factory
}

// Option is a core config option.
type Option func(c *CoreBusConfig) error

// WithControllerFactories adds built-in factories.
func WithControllerFactories(factories ...controller.Factory) Option {
	return func(c *CoreBusConfig) error {
		c.BuiltInFactories = append(c.BuiltInFactories, factories...)
		return nil
	}
}

// WithFactoryResolverCtor sets the Factory resolver constructor.
func WithFactoryResolverCtor(ctor FactoryResolverCtor) Option {
	return func(c *CoreBusConfig) error {
		c.FactoryResolverCtor = ctor
		return nil
	}
}

// NewCoreBus constructs a standard in-memory bus stack.
func NewCoreBus(
	ctx context.Context,
	le *logrus.Entry,
	opts ...Option,
) (bus.Bus, *static.Resolver, error) {
	dc := cdc.NewController(ctx, le)
	b := inmem.NewBus(dc)

	// Process options
	conf := &CoreBusConfig{}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(conf); err != nil {
			return nil, nil, err
		}
	}

	// Loader controller constructs and executes controllers
	cl, err := loader.NewController(le, b)
	if err != nil {
		return nil, nil, err
	}

	// Execute the loader controller.
	_, err = b.AddController(ctx, cl, nil)
	if err != nil {
		return nil, nil, err
	}

	// If there are any built in factories append them.
	sr := static.NewResolver(conf.BuiltInFactories...)
	sr.AddFactory(configset_controller.NewFactory(b))

	// fres is the factory resolver controller
	var fres controller.FactoryResolver = sr

	// Add a custom factory resolver wrapper if configured.
	if fctor := conf.FactoryResolverCtor; fctor != nil {
		fres, err = fctor(b, sr)
		if err != nil {
			return nil, nil, err
		}
	}

	// execute the factory resolver
	_, err = b.AddController(ctx, resolver.NewController(le, b, fres), nil)
	if err != nil {
		return nil, nil, err
	}

	return b, sr, nil
}
