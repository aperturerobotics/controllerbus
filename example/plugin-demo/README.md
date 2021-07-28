# Plugin Demo

[![asciicast](https://asciinema.org/a/418277.svg)](https://asciinema.org/a/418277)

## Demo

Note: Go plugins do not support unloading, and have other limitations at the
moment, see: https://github.com/golang/go/issues/27751

[![asciicast](https://asciinema.org/a/418277.svg)](https://asciinema.org/a/418277)

You can run this with `go run github.com/aperturerobotics/example/plugin-demo`:

 - Runs the module-compiler on the boilerplate, demo, and CLI controllers.
 - Executes the CLI pre-configured to load the demo plugins + controllers.
 - Exits a few seconds after successfully loading + executing
 
This is an end to end demo of plugin build + load via modules.

There are additional scripts to run the demo using the plugin CLI interface,
and/or to run the codegen and stop.

## Known Bugs

### Plugin Package Version Mismatches

Note: this is currently not working properly, possibly due to Go module
incompatibilities:

```
error="plugin.Open(\"plugins/example.cbus-plugin-abcdef.cbus\"): plugin was built
with a different version of package
cbus-plugin-abcdef/github.com/aperturerobotics/controllerbus/example/boilerplate/controller
```

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
