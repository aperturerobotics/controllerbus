package example_hot_plugin

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	boilerplate_controller "github.com/aperturerobotics/controllerbus/example/boilerplate/controller"
	hot_plugin "github.com/aperturerobotics/controllerbus/hot/plugin"
)

// BinaryID is the identifier for the plugin binary.
const BinaryID = "github.com/aperturerobotics/controllerbus/example/hot-demo/plugin"

// BinaryVersion is the version string.
const BinaryVersion = "0.0.0"

// BinaryFactories are the factories included in the binary.
var BinaryFactories = func(b bus.Bus) []controller.Factory {
	return []controller.Factory{
		boilerplate_controller.NewFactory(b),
	}
}

// ExamplePlugin is the top-level example plugin.
type ExamplePlugin struct {
}

// GetBinaryID returns the plugin binary ID.
// Usually the go.mod package name.
func (e *ExamplePlugin) GetBinaryID() string {
	return BinaryID
}

// GetBinaryVersion returns the plugin binary version
// Does not need to be semver (usually uses Go.mod versioning)
func (e *ExamplePlugin) GetBinaryVersion() string {
	return BinaryVersion
}

// NewHotResolver constructs the resolver and inits the plugin.
// ctx is canceled when the plugin is about to be unloaded.
func (e *ExamplePlugin) NewHotResolver(ctx context.Context, b bus.Bus) (hot_plugin.HotResolver, error) {
	return hot_plugin.NewResolver(
		e.GetBinaryID(),
		e.GetBinaryVersion(),
		BinaryFactories(b)...,
	), nil
}

// PrePluginUnload is called just before the plugin is unloaded.
func (e *ExamplePlugin) PrePluginUnload() {
	// noop
}

// _ is a type assertion
var _ hot_plugin.HotPlugin = ((*ExamplePlugin)(nil))
