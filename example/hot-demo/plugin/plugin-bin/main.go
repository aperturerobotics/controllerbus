package main

import (
	demo "github.com/aperturerobotics/controllerbus/example/hot-demo/plugin"
	hot_plugin "github.com/aperturerobotics/controllerbus/hot/plugin"
)

// ControllerBusHotPlugin is the root hot plugin that is loaded.
var ControllerBusHotPlugin hot_plugin.HotPlugin = &demo.ExamplePlugin{}

// plugin
// func main() {}
