package hot_plugin

import (
	"context"
	"strings"

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
	id             string
	staticResolver *static.Resolver
}

// NewResolver constructs a new resolver with a plugin binary..
func NewResolver(
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
		Version.String(),
	}, "/")
	staticResolver := static.NewResolver(factories...)
	return &Resolver{
		id:             id,
		staticResolver: staticResolver,
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
	return r.staticResolver.GetFactoryMatchingConfig(ctx, c)
}

// PrePluginUnload is called just before the plugin is unloaded.
func (r *Resolver) PrePluginUnload() {
	// noop
	// TODO kill all attached controllers?
}

// _ is a type assertion
var _ HotResolver = ((*Resolver)(nil))
