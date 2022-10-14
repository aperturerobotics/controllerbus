package resolver

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// LoadFactoryByConfigResolver implements directive.Resolver for loading a
// controller factory with a config object.
type LoadFactoryByConfigResolver struct {
	// bus is the controller bus
	bus bus.Bus
	// ctx is the directive context
	ctx context.Context
	// res is the factory resolver
	res controller.FactoryResolver
	// dir is the directive
	dir LoadFactoryByConfig
}

// resolveLoadFactoryByConfig executes the LoadFactoryByConfig directive.
func (c *Controller) resolveLoadFactoryByConfig(
	ctx context.Context,
	dir LoadFactoryByConfig,
) ([]directive.Resolver, error) {
	return directive.Resolvers(&LoadFactoryByConfigResolver{
		ctx: ctx,
		res: c.resolver,
		dir: dir,
		bus: c.bus,
	}), nil
}

// Resolve resolves the values.
// Any fatal error resolving the value is returned.
func (r *LoadFactoryByConfigResolver) Resolve(
	ctx context.Context,
	vh directive.ResolverHandler,
) error {
	config := r.dir.LoadFactoryByConfig()
	factory, err := r.res.GetFactoryMatchingConfig(ctx, config)
	if err != nil {
		return err
	}
	if factory == nil {
		return nil
	}

	id, accepted := vh.AddValue(factory)
	if accepted {
		<-ctx.Done()
		vh.RemoveValue(id)
	}
	return nil
}

// _ is a type assertion
var _ directive.Resolver = ((*LoadFactoryByConfigResolver)(nil))
