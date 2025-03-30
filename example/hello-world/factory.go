package main

import (
	"context"
	"errors"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/blang/semver/v4"
)

// toyFactoryVersion is the compiled-in version
var toyFactoryVersion = semver.MustParse("0.1.0")

// toyControllerID is the controller id
const toyControllerID = "hello-world"

// ToyFactory implements the toy controller factory.
type ToyFactory struct{}

// NewToyFactory builds a new generic factory.
func NewToyFactory() *ToyFactory {
	return &ToyFactory{}
}

// GetControllerID returns the unique ID for the controller.
func (f *ToyFactory) GetControllerID() string {
	return toyControllerID
}

// GetConfigID returns the unique ID for the controller config.
func (f *ToyFactory) GetConfigID() string {
	return f.GetControllerID()
}

// ConstructConfig constructs an instance of the controller configuration.
func (f *ToyFactory) ConstructConfig() config.Config {
	return &ToyControllerConfig{}
}

// Construct constructs the associated controller given configuration.
func (f *ToyFactory) Construct(
	ctx context.Context,
	c config.Config,
	opts controller.ConstructOpts,
) (controller.Controller, error) {
	conf, ok := c.(*ToyControllerConfig)
	if !ok {
		return nil, errors.New("wrong type of config")
	}

	return NewToyController(opts.GetLogger(), conf)
}

// GetVersion returns the version of this controller.
func (f *ToyFactory) GetVersion() semver.Version {
	return toyFactoryVersion
}

// _ is a type assertion
var _ controller.Factory = ((*ToyFactory)(nil))
