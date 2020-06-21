package example_hot_plugin

import (
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
type ExamplePlugin = hot_plugin.StaticPlugin

// NewExamplePlugin constructs a new example plugin.
func NewExamplePlugin() *ExamplePlugin {
	return hot_plugin.NewStaticPlugin(
		BinaryID,
		BinaryVersion,
		BinaryFactories,
	)
}

// _ is a type assertion
var _ hot_plugin.HotPlugin = ((*ExamplePlugin)(nil))
