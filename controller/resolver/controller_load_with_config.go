package resolver

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/directive"
)

// loadWithConfigResolver implements directive.Resolver for loading a controller
// with a config.
type loadWithConfigResolver struct {
	// bus is the controller bus
	bus bus.Bus
	// res is the factory resolver
	res controller.FactoryResolver
	// dir is the directive
	dir LoadControllerWithConfig
}

// resolveLoadControllerWithConfig executes the LoadControllerWithConfig directive.
func (c *Controller) resolveLoadControllerWithConfig(
	dir LoadControllerWithConfig,
) (directive.Resolver, error) {
	return &loadWithConfigResolver{res: c.resolver, dir: dir, bus: c.bus}, nil
}

// Resolve resolves the values.
// Any fatal error resolving the value is returned.
// When the context is canceled valCh will not be drained anymore.
func (r *loadWithConfigResolver) Resolve(ctx context.Context, valCh chan<- directive.Value) error {
	conf := r.dir.GetDesiredControllerConfig()
	factory, err := r.res.GetFactoryMatchingConfig(ctx, conf)
	if err != nil {
		return err
	}

	if factory == nil {
		return nil
	}

	execDir := loader.NewExecControllerSingleton(factory, conf)
	_, execRef, err := r.bus.AddDirective(execDir, func(val directive.Value) {
		select {
		case <-ctx.Done():
		case valCh <- val:
		}
	})
	if err != nil {
		return err
	}

	// cancel the reference when ctx is canceled
	<-ctx.Done()
	execRef.Release()
	return ctx.Err()
}

// _ is a type assertion
var _ directive.Resolver = ((*loadWithConfigResolver)(nil))
