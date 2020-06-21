package hot_plugin

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
)

// FactoryCtorSet is a list of factory constructors.
type FactoryCtorSet = func(b bus.Bus) []controller.Factory

// StaticPlugin contains a compiled set of controller factories.
type StaticPlugin struct {
	binaryID      string
	binaryVersion string
	factories     FactoryCtorSet
}

// NewStaticPlugin constructs a new static plugin.
func NewStaticPlugin(
	binaryID string,
	binaryVersion string,
	factories FactoryCtorSet,
) *StaticPlugin {
	return &StaticPlugin{
		binaryID:      binaryID,
		binaryVersion: binaryVersion,
		factories:     factories,
	}
}

// GetBinaryID returns the plugin binary ID.
// Usually the go.mod package name.
func (e *StaticPlugin) GetBinaryID() string {
	return e.binaryID
}

// GetBinaryVersion returns the plugin binary version
// Does not need to be semver (usually uses Go.mod versioning)
func (e *StaticPlugin) GetBinaryVersion() string {
	return e.binaryVersion
}

// NewHotResolver constructs the resolver and inits the plugin.
// ctx is canceled when the plugin is about to be unloaded.
func (e *StaticPlugin) NewHotResolver(ctx context.Context, b bus.Bus) (HotResolver, error) {
	factories := e.factories(b)
	return NewResolver(
		e.GetBinaryID(),
		e.GetBinaryVersion(),
		factories...,
	), nil
}

// PrePluginUnload is called just before the plugin is unloaded.
func (e *StaticPlugin) PrePluginUnload() {
	// noop
}

// _ is a type assertion
var _ HotPlugin = ((*StaticPlugin)(nil))
