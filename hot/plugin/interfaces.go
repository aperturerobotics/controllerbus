package hot_plugin

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
)

// UnloadHandler is called before the plugin is unloaded.
type UnloadHandler interface {
	// PrePluginUnload is called just before the plugin is unloaded.
	//
	// Unloading:
	// 1. context is canceled.
	// 2. PrePluginUnload called on all resolvers.
	// 3. PrePluginUnload called on plugin
	// 4. Unload of plugin
	PrePluginUnload()
}

// HotResolver resolves types included in a Hot binary.
type HotResolver interface {
	// FactoryResolver indicates this implements FactoryResolver.
	controller.FactoryResolver
	// UnloadHandler indicates resolver implements pre-plugin unload
	UnloadHandler
}

// HotPlugin is the top-level type exposed in a Hot binary.
type HotPlugin interface {
	// UnloadHandler indicates plugin implements pre-plugin unload
	UnloadHandler

	// GetBinaryID returns the plugin binary ID.
	// Usually the go.mod package name.
	GetBinaryID() string
	// GetBinaryVersion returns the plugin binary version
	// Does not need to be semver (usually uses Go.mod versioning)
	GetBinaryVersion() string
	// NewHotResolver constructs the resolver and inits the plugin.
	// ctx is canceled when the plugin is about to be unloaded.
	NewHotResolver(ctx context.Context, bus bus.Bus) (HotResolver, error)
}
