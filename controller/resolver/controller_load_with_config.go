package resolver

import (
	"context"
	"time"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/directive"
)

// loadWithConfigResolver implements directive.Resolver for loading a controller
// with a config.
type loadWithConfigResolver struct {
	// c is the controller
	c *Controller
	// bus is the controller bus
	bus bus.Bus
	// ctx is the directive context
	ctx context.Context
	// res is the factory resolver
	res controller.FactoryResolver
	// dir is the directive
	dir LoadControllerWithConfig
}

// resolveLoadControllerWithConfig executes the LoadControllerWithConfig directive.
func (c *Controller) resolveLoadControllerWithConfig(
	ctx context.Context,
	dir LoadControllerWithConfig,
) (directive.Resolver, error) {
	return &loadWithConfigResolver{c: c, ctx: ctx, res: c.resolver, dir: dir, bus: c.bus}, nil
}

// Resolve resolves the values.
// Any fatal error resolving the value is returned.
// When the context is canceled valCh will not be drained anymore.
func (r *loadWithConfigResolver) Resolve(ctx context.Context, vh directive.ResolverHandler) error {
	conf := r.dir.GetDesiredControllerConfig()
	var factoryRef directive.Reference
	factory, err := r.res.GetFactoryMatchingConfig(ctx, conf)
	if err != nil {
		return err
	}

	if factory == nil {
		// create a directive to lookup the factory.
		factory, factoryRef, err = ExLoadFactoryByConfig(ctx, r.bus, conf)
		if err != nil {
			return err
		}
		if factory == nil {
			factoryRef.Release()
			return nil
		}
	}

	factoryCtx := context.Background()
	factoryWithCtx, factoryWithCtxOk := factory.(controller.FactoryWithContext)
	if factoryWithCtxOk {
		factoryCtx = factoryWithCtx.GetFactoryContext()
	}

	valCtx, valCtxCancel := context.WithCancel(r.ctx)
	execDir := loader.NewExecControllerSingleton(factory, conf)

	// pass through all values.
	_, execRef, err := r.bus.AddDirective(execDir, bus.NewPassThruHandler(vh, valCtxCancel))
	if err != nil {
		_, _ = vh.AddValue(loader.NewExecControllerValue(
			time.Now(),
			time.Time{},
			nil,
			err,
		))
		// config has to be invalid for it to have failed here.
		// give up permanently
		// return err
		factoryRef.Release()
		return nil
	}

	// cancel the reference when ctx is canceled
	// note: it's correct to spawn a new goroutine here; indicate the resolver has exited.
	go func() {
		select {
		case <-factoryCtx.Done():
		case <-valCtx.Done():
		case <-r.c.subCtx.Done():
		}
		execRef.Release()
		factoryRef.Release()
		valCtxCancel()
	}()

	return nil
}

// _ is a type assertion
var _ directive.Resolver = ((*loadWithConfigResolver)(nil))
