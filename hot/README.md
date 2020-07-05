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

## Known Bugs

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
