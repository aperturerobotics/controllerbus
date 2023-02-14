package plugin

import (
	"context"
	"strings"
	"sync"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/resolver/static"
	"github.com/blang/semver"
)

// Version is the resolver version.
var Version = semver.MustParse("0.0.1")

// Resolver implements the controller resolver using a list of built-in
// controller implementations.
//
// Does not require semantic versioning for binary versions.
type Resolver struct {
	ctx                 context.Context
	id                  string
	pluginBinaryID      string
	pluginBinaryVersion string
	staticResolver      *static.Resolver
	bus                 bus.Bus

	mtx       sync.Mutex
	preUnload []func()
}

// NewResolver constructs a new resolver with a plugin binary.
func NewResolver(
	ctx context.Context,
	bus bus.Bus,
	pluginBinaryID string,
	pluginBinaryVersion string,
	factories ...controller.Factory,
) *Resolver {
	id := strings.Join([]string{
		"controllerbus",
		"hot",
		"plugin",
		pluginBinaryID,
		pluginBinaryVersion,
		"static-resolver",
	}, "/")
	staticResolver := static.NewResolver(factories...)
	return &Resolver{
		ctx:                 ctx,
		bus:                 bus,
		id:                  id,
		staticResolver:      staticResolver,
		pluginBinaryID:      pluginBinaryID,
		pluginBinaryVersion: pluginBinaryVersion,
	}
}

// GetResolverID returns the resolver identifier.
func (r *Resolver) GetResolverID() string {
	return r.id
}

// GetResolverVersion returns the resolver version.
func (r *Resolver) GetResolverVersion() semver.Version {
	return Version
}

// GetConfigCtorByID returns a config constructor matching the ID.
// If none found, return nil, nil
func (r *Resolver) GetConfigCtorByID(
	ctx context.Context, id string,
) (config.Constructor, error) {
	return r.staticResolver.GetConfigCtorByID(ctx, id)
}

// GetFactoryMatchingConfig returns the factory that matches the config.
// If no factory is found, return nil.
// If an unexpected error occurs, return it.
func (r *Resolver) GetFactoryMatchingConfig(
	ctx context.Context, c config.Config,
) (controller.Factory, error) {
	tfac, err := r.staticResolver.GetFactoryMatchingConfig(ctx, c)
	if err != nil || tfac == nil {
		return tfac, err
	}

	// wrap factories with pre-unload hook
	sf := NewStaticPluginFactory(r.ctx, tfac, r.pluginBinaryID, r.bus)
	r.mtx.Lock()
	r.preUnload = append(r.preUnload, sf.Close)
	r.mtx.Unlock()
	return sf, nil
}

// PrePluginUnload is called just before the plugin is unloaded.
func (r *Resolver) PrePluginUnload() {
	r.mtx.Lock()
	for _, f := range r.preUnload {
		f()
	}
	r.preUnload = nil
	r.mtx.Unlock()
}

// _ is a type assertion
var _ PluginResolver = ((*Resolver)(nil))
