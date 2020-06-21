package resolver

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// LoadConfigConstructorByIDResolver implements directive.Resolver for loading a
// controller configuration constructor with a config ID.
type LoadConfigConstructorByIDResolver struct {
	// bus is the controller bus
	bus bus.Bus
	// ctx is the directive context
	ctx context.Context
	// res is the factory resolver
	res controller.FactoryResolver
	// dir is the directive
	dir LoadConfigConstructorByID
}

// resolveLoadConfigConstructorByID executes the LoadConfigConstructorByID directive.
func (c *Controller) resolveLoadConfigConstructorByID(
	ctx context.Context,
	dir LoadConfigConstructorByID,
) (directive.Resolver, error) {
	return &LoadConfigConstructorByIDResolver{
		ctx: ctx,
		res: c.resolver,
		dir: dir,
		bus: c.bus,
	}, nil
}

// Resolve resolves the values.
// Any fatal error resolving the value is returned.
func (r *LoadConfigConstructorByIDResolver) Resolve(
	ctx context.Context,
	vh directive.ResolverHandler,
) error {
	configID := r.dir.LoadConfigConstructorByIDConfigID()
	ctor, err := r.res.GetConfigCtorByID(ctx, configID)
	if err != nil {
		return err
	}

	if ctor == nil {
		return nil
	}

	id, accepted := vh.AddValue(ctor)
	if accepted {
		<-ctx.Done()
		vh.RemoveValue(id)
	}
	return nil
}

// _ is a type assertion
var _ directive.Resolver = ((*LoadConfigConstructorByIDResolver)(nil))
