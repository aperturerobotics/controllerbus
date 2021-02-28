# Hot Loader Demo

You can run this with `go run main.go`:

 - Runs the module-compiler on the boilerplate, demo, and CLI controllers.
 - Executes the CLI pre-configured to load the demo plugins + controllers.
 - Exits a few seconds after successfully loading + executing
 
This is an end to end demo of plugin build + load via modules.

There are additional scripts to run the demo using the hot-builder CLI
interface, and/or to run the codegen and stop.

## Known Issues

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
