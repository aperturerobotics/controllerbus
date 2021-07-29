# Plugin Demo

[![asciicast](https://asciinema.org/a/I4LOCViLwzRlztYc1rytgAxWp.svg)](https://asciinema.org/a/I4LOCViLwzRlztYc1rytgAxWp)

## Demo

You can run this with `go run -v -trimpath ./` in this directory:

 - Runs the module-compiler on the boilerplate, demo, and CLI controllers.
 - Executes the CLI pre-configured to load the demo plugins + controllers.
 - Exits a few seconds after successfully loading + executing
 
This is an end to end demo of plugin build + load via modules.

There are additional scripts to run the demo using the plugin CLI interface,
and/or to run the codegen and stop.

## Known Bugs

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
