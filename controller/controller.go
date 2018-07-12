package controller

import (
	"context"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
)

// Controller tracks a particular process.
type Controller interface {
	// Handler handles directives.
	directive.Handler

	// Execute executes the given controller.
	// Returning nil ends execution.
	// Returning an error triggers a retry with backoff.
	Execute(ctx context.Context) error
	// Close releases any resources used by the controller.
	// Error indicates any issue encountered releasing.
	Close() error
}

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
	// GetControllerID returns the unique ID for the controller.
	GetControllerID() string
	// ConstructConfig constructs an instance of the controller configuration.
	ConstructConfig() config.Config
	// Construct constructs the associated controller given configuration.
	Construct(config.Config, ConstructOpts) (Controller, error)
	// GetVersion returns the version of this controller.
	GetVersion() semver.Version
}

// FactoryResolver looks up factories that match configurations.
type FactoryResolver interface {
	// GetConfigByID returns a config object and factory matching the ID.
	// If none found, return nil, nil
	GetConfigByID(id string) (config.Config, Factory, error)
	// GetFactoryMatchingConfig returns the factory that matches the config.
	// If no factory is found, return nil.
	// If an unexpected error occurs, return it.
	GetFactoryMatchingConfig(config.Config) (Factory, error)
}
