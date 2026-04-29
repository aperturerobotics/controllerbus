package resolver

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// FactoryResolverWatcher is an optional side interface for FactoryResolver
// implementations that can wake pending lookups when their factory set
// changes. When implemented, LoadFactoryByConfig holds the directive open and
// re-checks after the wait channel fires, enabling partial ConfigSet apply
// where some factories register later. Implementations without this interface
// fall back to a single lookup with no re-check.
type FactoryResolverWatcher interface {
	// GetFactoryMatchingConfigWatch returns the factory matching the config
	// and a wait channel that closes when the resolver's factory set changes.
	// Returns (nil, waitCh, nil) when no factory currently matches.
	GetFactoryMatchingConfigWatch(
		ctx context.Context, conf config.Config,
	) (controller.Factory, <-chan struct{}, error)
}

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
	conf := r.dir.LoadFactoryByConfig()
	watcher, _ := r.res.(FactoryResolverWatcher)

	for {
		var (
			factory controller.Factory
			waitCh  <-chan struct{}
			err     error
		)
		if watcher != nil {
			factory, waitCh, err = watcher.GetFactoryMatchingConfigWatch(ctx, conf)
		} else {
			factory, err = r.res.GetFactoryMatchingConfig(ctx, conf)
		}
		if err != nil {
			return err
		}
		if factory != nil {
			id, accepted := vh.AddValue(factory)
			if accepted {
				<-ctx.Done()
				vh.RemoveValue(id)
			}
			return nil
		}
		if waitCh == nil {
			// No watcher; legacy behavior.
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case <-waitCh:
		}
	}
}

// _ is a type assertion
var _ directive.Resolver = ((*LoadFactoryByConfigResolver)(nil))
