# Plugins

[![asciicast](https://asciinema.org/a/I4LOCViLwzRlztYc1rytgAxWp.svg)](../example/plugin-demo)

Check out the [plugin demos](../example/plugin-demo).

Analyze Go code to bundle controller factories into Plugin modules.

Features:

 - Scan packages for ControllerBus factories to include
 - Automatically codegen and compile modules for all targets
 - Full support for Go modules and module replacements
 - Manifest of all currently loaded plugins
 - Detect Go module code changes and hot-reload
 - Multiple bundling and plugin loading approaches

Using the IPC system, a plugin can also be loaded as a sub-process communicating
over stdin/stdout (or named pipes on Windows).
 
## Codegen Output Example

Example generated code for a plugin:

```go
// +build controllerbus_plugin

package main

import (
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	boilerplate_controller "cbus-plugin-abcdef/github.com/aperturerobotics/controllerbus/example/boilerplate/controller"
	hot_compile_demo_controller "cbus-plugin-abcdef/github.com/aperturerobotics/controllerbus/example/plugin-demo/demo-controller"
	"github.com/aperturerobotics/controllerbus/plugin"
)

// BinaryID is the binary identifier.
const (
	BinaryID	= "hot-demo-module"
	BinaryVersion	= "cbus-plugin-abcdef"
)

// BinaryFactories are the factories included in the binary.
var BinaryFactories = func(b bus.Bus) []controller.Factory {
	return []controller.Factory{boilerplate_controller.NewFactory(b), hot_compile_demo_controller.NewFactory(b)}
}

// Plugin is the top-level static plugin container.
type Plugin = plugin.StaticPlugin

// NewPlugin constructs the static container plugin.
func NewPlugin() *Plugin {
	return plugin.NewStaticPlugin(BinaryID, BinaryVersion, BinaryFactories)
}

// ControllerBusPlugin is the variable read by the plugin loader.
var (
	ControllerBusPlugin	plugin.Plugin	= NewPlugin()
	_			plugin.Plugin	= ((*Plugin)(nil))
	HotPluginBuildPrefix			= "cbus-plugin-abcdef"
)

var HotPluginBuildUUID = `cbus-plugin-abcdef`
```
