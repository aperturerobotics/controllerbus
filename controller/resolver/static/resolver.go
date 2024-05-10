package static

import (
	"context"
	"slices"
	"strings"
	"sync"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/blang/semver"
	"golang.org/x/exp/maps"
)

// ResolverID is the resolver identifier.
const ResolverID = "static"

// Version is the resolver version.
var Version = semver.MustParse("0.0.1")

// Resolver implements the controller resolver using a list of built-in
// controller implementations.
type Resolver struct {
	factoryMtx sync.Mutex
	factories  map[string]controller.Factory
}

// NewResolver constructs a new resolver.
func NewResolver(factories ...controller.Factory) *Resolver {
	r := &Resolver{
		factories: make(map[string]controller.Factory),
	}

	for _, f := range factories {
		r.AddFactory(f)
	}

	return r
}

// GetResolverID returns the resolver identifier.
func (r *Resolver) GetResolverID() string {
	return ResolverID
}

// GetResolverVersion returns the resolver version.
func (r *Resolver) GetResolverVersion() semver.Version {
	return Version
}

// AddFactory adds a factory to the resolver.
func (r *Resolver) AddFactory(factory controller.Factory) {
	configID := factory.GetConfigID()
	version := factory.GetVersion()

	r.factoryMtx.Lock()
	defer r.factoryMtx.Unlock()

	existing, ok := r.factories[configID]
	if ok {
		existingVer := existing.GetVersion()
		if existingVer.GTE(version) {
			return
		}
	}

	r.factories[configID] = factory
}

// GetFactories returns the factories associated w/ the resolver.
func (r *Resolver) GetFactories() []controller.Factory {
	vals := maps.Values(r.factories)
	slices.SortFunc(vals, func(a, b controller.Factory) int {
		return strings.Compare(a.GetConfigID(), b.GetConfigID())
	})
	return vals
}

// GetConfigCtorByID returns a config constructor matching the ID.
// If none found, return nil, nil
func (r *Resolver) GetConfigCtorByID(
	ctx context.Context, id string,
) (config.Constructor, error) {
	r.factoryMtx.Lock()
	defer r.factoryMtx.Unlock()

	for _, f := range r.factories {
		cid := f.GetConfigID()
		if cid == id {
			return NewConfigCtor(cid, f), nil
		}
	}

	return nil, nil
}

// GetFactoryMatchingConfig returns the factory that matches the config.
// If no factory is found, return nil.
// If an unexpected error occurs, return it.
func (r *Resolver) GetFactoryMatchingConfig(
	ctx context.Context, c config.Config,
) (controller.Factory, error) {
	r.factoryMtx.Lock()
	defer r.factoryMtx.Unlock()

	f, ok := r.factories[c.GetConfigID()]
	if ok {
		return f, nil
	}

	return nil, nil
}

// _ is a type assertion
var _ controller.FactoryResolver = ((*Resolver)(nil))
