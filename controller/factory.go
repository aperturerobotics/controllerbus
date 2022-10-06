package controller

import (
	"context"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
)

// ConstructOpts contains optional parameters when constructing a controller.
type ConstructOpts struct {
	// Logger is the root logger to use.
	Logger *logrus.Entry
}

// GetLogger returns the specified logger or a default one if nil.
func (c *ConstructOpts) GetLogger() *logrus.Entry {
	if l := c.Logger; l != nil {
		return l
	}

	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	return logrus.NewEntry(log)
}

// Factory represents an available controller implementation.
// The factory can construct instances of the controller given configuration.
type Factory interface {
	// GetConfigID returns the unique config ID for the controller.
	GetConfigID() string
	// ConstructConfig constructs an instance of the controller configuration.
	ConstructConfig() config.Config
	// Construct constructs the associated controller given configuration.
	Construct(config.Config, ConstructOpts) (Controller, error)
	// GetVersion returns the version of this controller.
	GetVersion() semver.Version
}

// FactoryWithContext is a factory with a cancelation context.
type FactoryWithContext interface {
	Factory
	// GetFactoryContext returns a context that is canceled if the factory is unloaded.
	GetFactoryContext() context.Context
}

// FactoryResolver looks up factories that match configurations.
type FactoryResolver interface {
	// GetResolverID returns the resolver identifier.
	// Ex: static, go-plugin
	GetResolverID() string
	// GetResolverVersion returns the resolver version.
	GetResolverVersion() semver.Version
	// GetConfigCtorByID returns a config constructor by ID.
	// If none found, return nil, nil
	GetConfigCtorByID(ctx context.Context, id string) (config.Constructor, error)
	// GetFactoryMatchingConfig returns the factory that matches the config.
	// If no factory is found, return nil.
	// If an unexpected error occurs, return it.
	GetFactoryMatchingConfig(ctx context.Context, conf config.Config) (Factory, error)
}
