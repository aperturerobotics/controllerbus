//+build controllerbus_hot_plugin

package main

import (
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	boilerplate_controller "github.com/aperturerobotics/controllerbus/example/boilerplate/controller"
	"github.com/aperturerobotics/controllerbus/hot/plugin"
)
// BinaryID is the binary identifier.
const BinaryID = "example-binary"
// BinaryVersion is the binary version string.
const BinaryVersion = "0.0.0"
// BinaryFactories are the factories included in the binary.
var BinaryFactories = func(b bus.Bus) []controller.Factory {
	return []controller.Factory{boilerplate_controller.NewFactory(b)}
}
// Plugin is the top-level static plugin container.
type Plugin = hot_plugin.StaticPlugin
// NewPlugin constructs the static container plugin.
func NewPlugin() *Plugin {
	return hot_plugin.NewStaticPlugin(BinaryID, BinaryVersion, BinaryFactories)
}
// ControllerBusHotPlugin is the variable read by the plugin loader.
var ControllerBusHotPlugin hot_plugin.HotPlugin = NewPlugin()
// _ is a type assertion
var _ hot_plugin.HotPlugin = ((*Plugin)(nil))
