package static

import (
	"sync"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
)

// Resolver implements the controller resolver using a list of built-in
// controller implementations.
type Resolver struct {
	factoryMtx sync.Mutex
	factories  map[string]controller.Factory
}

// NewResolver constructs a new resolver.
func NewResolver() *Resolver {
	return &Resolver{factories: make(map[string]controller.Factory)}
}

// AddFactory adds a factory to the resolver.
func (r *Resolver) AddFactory(factory controller.Factory) {
	controllerID := factory.GetControllerID()
	version := factory.GetVersion()

	r.factoryMtx.Lock()
	defer r.factoryMtx.Unlock()

	existing, ok := r.factories[controllerID]
	if ok {
		existingVer := existing.GetVersion()
		if existingVer.GTE(version) {
			return
		}
	}

	r.factories[controllerID] = factory
}

// GetConfigByID returns a config object and factory matching the ID.
// If none found, return nil, nil
func (r *Resolver) GetConfigByID(id string) (config.Config, controller.Factory, error) {
	r.factoryMtx.Lock()
	defer r.factoryMtx.Unlock()

	f, ok := r.factories[id]
	if !ok {
		return nil, nil, nil
	}

	return f.ConstructConfig(), f, nil
}

// GetFactoryMatchingConfig returns the factory that matches the config.
// If no factory is found, return nil.
// If an unexpected error occurs, return it.
func (r *Resolver) GetFactoryMatchingConfig(c config.Config) (controller.Factory, error) {
	r.factoryMtx.Lock()
	defer r.factoryMtx.Unlock()

	for _, f := range r.factories {
		cid := f.ConstructConfig().GetConfigID()
		if cid == c.GetConfigID() {
			return f, nil
		}
	}

	return nil, nil
}

// _ is a type assertion
var _ controller.FactoryResolver = ((*Resolver)(nil))
