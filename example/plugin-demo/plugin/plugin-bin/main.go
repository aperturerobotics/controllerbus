package main

import (
	demo "github.com/aperturerobotics/controllerbus/example/plugin-demo/plugin"
	plugin "github.com/aperturerobotics/controllerbus/plugin"
)

// ControllerBusPlugin is the root plugin that is loaded.
var ControllerBusPlugin plugin.Plugin = demo.NewExamplePlugin()
