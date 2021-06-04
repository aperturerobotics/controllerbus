# Hot Reloading

This is a **experimental** prototype of Hot Reload for controllerbus.

Features:

 - Automatically codegen and compile overlay modules for a target
 - Load + resolve controllerbus controllers from plugins
 - Full support for Go modules and module replacements
 - Manifest of all currently loaded plugins
 - Detect Go module code changes and hot-reload

Cavets:

 - Go plugin system cannot fully unload memory, making a mem leak.
 - Plugins only work on Linux.
 - Fails if any of the packages shared w/ host process are changed.
 - All builds must use the "-trimpath" flag.
 
These are acceptable for development environments. Memory leak issue could be
addressed with ControllerBus IPC using cross-process controllers.

Reference: https://github.com/golang/go/issues/27751

## Demo

[![asciicast](https://asciinema.org/a/418277.svg)](https://asciinema.org/a/418277)

You can run this with `go run github.com/aperturerobotics/controllerbus/hot/demo`:

 - Runs the module-compiler on the boilerplate, demo, and CLI controllers.
 - Executes the CLI pre-configured to load the demo plugins + controllers.
 - Exits a few seconds after successfully loading + executing
 
This is an end to end demo of plugin build + load via modules.

There are additional scripts to run the demo using the hot-builder CLI
interface, and/or to run the codegen and stop.

## Known Bugs

### Plugin Package Version Mismatches

Note: this is currently not working properly, possibly due to Go module
incompatibilities:

WARN[0000] unable to load plugin file 
controller=controllerbus/hot/loader/filesystem/1 
error="plugin.Open(\"plugins/example.cbus-hot-abcdef.cbus\"): 
plugin was built with a different version of package cbus-hot-abcdef/github.com/aperturerobotics/controllerbus/example/boilerplate/controller

However, the "hot compilation" system is nevertheless useful for bundling
together controllers into IPC libraries that are then hosted by the IPC host
controller. The IPC system will be used until the Go "plugin" DLL loading system
is better supported, particularly with unloading modules.

### Unable to Unload Plugins

When a plugin is compiled, the "main" package is actually compiled to a module
named "plugin/unnamed-hashstrhere" with a hash of the main file contents. This
package is reloaded correctly because it has a unique name (unique ID).

However, this doesn't accomplish what we want. I actually want to re-load the
package named "github.com/aperturerobotics/controllerbus/example/boilerplate".
Since this package is already loaded, Go does not re-load it from the plugin.

This code already will automatically insert entropy into the main package
contents to produce a unique plugin package ID. But this isn't enough, we need
to clobber the old code in the module tables as well.

Loading plugins works great. Re-loading them: not quite yet.

## Codegen Output Example

Example generated code:

```go
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
```
