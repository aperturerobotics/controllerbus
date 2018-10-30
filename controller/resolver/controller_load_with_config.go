package resolver

import (
	"context"
	"errors"

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
	return &loadWithConfigResolver{ctx: ctx, res: c.resolver, dir: dir, bus: c.bus}, nil
}

// Resolve resolves the values.
// Any fatal error resolving the value is returned.
// When the context is canceled valCh will not be drained anymore.
func (r *loadWithConfigResolver) Resolve(ctx context.Context, vh directive.ResolverHandler) error {
	conf := r.dir.GetDesiredControllerConfig()
	factory, err := r.res.GetFactoryMatchingConfig(ctx, conf)
	if err != nil {
		return err
	}

	if factory == nil {
		return nil
	}

	valCtx, valCtxCancel := context.WithCancel(r.ctx)
	execDir := loader.NewExecControllerSingleton(factory, conf)
	execVal, execRef, err := bus.ExecOneOff(ctx, r.bus, execDir, valCtxCancel)
	if err != nil {
		return err
	}

	vid, added := vh.AddValue(execVal)
	if !added {
		execRef.Release()
		return errors.New("exec controller value rejected, likely due to duplicate controller load")
	}

	// cancel the reference when ctx is canceled
	// TODO: handle early dispose of exec directive?
	go func() {
		<-valCtx.Done()
		// valCtxCancel()
		execRef.Release()
		_, _ = vh.RemoveValue(vid)
	}()

	return nil
}

// _ is a type assertion
var _ directive.Resolver = ((*loadWithConfigResolver)(nil))
