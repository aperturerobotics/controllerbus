![Controller Bus](./doc/img/controller-bus-logo.png)

## Introduction

[![Go Reference Widget]][Go Reference] [![Go Report Card Widget]][Go Report Card] [![DeepWiki Widget]][DeepWiki]

[Go Reference]: https://pkg.go.dev/github.com/aperturerobotics/controllerbus
[Go Reference Widget]:https://pkg.go.dev/badge/github.com/aperturerobotics/controllerbus.svg
[Go Report Card Widget]: https://goreportcard.com/badge/github.com/aperturerobotics/controllerbus
[Go Report Card]: https://goreportcard.com/report/github.com/aperturerobotics/controllerbus
[DeepWiki Widget]: https://img.shields.io/badge/DeepWiki-aperturerobotics%2Fcontrollerbus-blue.svg?logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACwAAAAyCAYAAAAnWDnqAAAAAXNSR0IArs4c6QAAA05JREFUaEPtmUtyEzEQhtWTQyQLHNak2AB7ZnyXZMEjXMGeK/AIi+QuHrMnbChYY7MIh8g01fJoopFb0uhhEqqcbWTp06/uv1saEDv4O3n3dV60RfP947Mm9/SQc0ICFQgzfc4CYZoTPAswgSJCCUJUnAAoRHOAUOcATwbmVLWdGoH//PB8mnKqScAhsD0kYP3j/Yt5LPQe2KvcXmGvRHcDnpxfL2zOYJ1mFwrryWTz0advv1Ut4CJgf5uhDuDj5eUcAUoahrdY/56ebRWeraTjMt/00Sh3UDtjgHtQNHwcRGOC98BJEAEymycmYcWwOprTgcB6VZ5JK5TAJ+fXGLBm3FDAmn6oPPjR4rKCAoJCal2eAiQp2x0vxTPB3ALO2CRkwmDy5WohzBDwSEFKRwPbknEggCPB/imwrycgxX2NzoMCHhPkDwqYMr9tRcP5qNrMZHkVnOjRMWwLCcr8ohBVb1OMjxLwGCvjTikrsBOiA6fNyCrm8V1rP93iVPpwaE+gO0SsWmPiXB+jikdf6SizrT5qKasx5j8ABbHpFTx+vFXp9EnYQmLx02h1QTTrl6eDqxLnGjporxl3NL3agEvXdT0WmEost648sQOYAeJS9Q7bfUVoMGnjo4AZdUMQku50McDcMWcBPvr0SzbTAFDfvJqwLzgxwATnCgnp4wDl6Aa+Ax283gghmj+vj7feE2KBBRMW3FzOpLOADl0Isb5587h/U4gGvkt5v60Z1VLG8BhYjbzRwyQZemwAd6cCR5/XFWLYZRIMpX39AR0tjaGGiGzLVyhse5C9RKC6ai42ppWPKiBagOvaYk8lO7DajerabOZP46Lby5wKjw1HCRx7p9sVMOWGzb/vA1hwiWc6jm3MvQDTogQkiqIhJV0nBQBTU+3okKCFDy9WwferkHjtxib7t3xIUQtHxnIwtx4mpg26/HfwVNVDb4oI9RHmx5WGelRVlrtiw43zboCLaxv46AZeB3IlTkwouebTr1y2NjSpHz68WNFjHvupy3q8TFn3Hos2IAk4Ju5dCo8B3wP7VPr/FGaKiG+T+v+TQqIrOqMTL1VdWV1DdmcbO8KXBz6esmYWYKPwDL5b5FA1a0hwapHiom0r/cKaoqr+27/XcrS5UwSMbQAAAABJRU5ErkJggg==
[DeepWiki]: https://deepwiki.com/aperturerobotics/controllerbus

**ControllerBus** is a framework for **communicating control loops**:

 - **Configurable**: flexible self-documenting config with Protobuf and YAML.
 - **Cross-platform**: supports web browsers, servers, desktop, mobile, ...
 - **Hot-loadable**: plugins and IPC dynamically add controllers at runtime.
 - **Modular**: easily combine together application components w/o glue code.
 - **Declarative**: de-duplicated declarative requests between controllers.

The primary concepts are:

 - **Config**: configuration for a controller or process.
 - **Controller**: goroutine which can create & handle Directives.
 - **Directive**: a cross-controller request or declaration of target state.
 - **Bus**: communication channel between running Controllers.
 - **Factory**: constructor for controller and configuration objects.

Controller Bus provides a pattern for structuring modular Go projects.

## Examples

[![Support Server](https://img.shields.io/discord/803825858599059487.svg?label=Discord&logo=Discord&colorB=7289da&style=for-the-badge)](https://discord.gg/ZAZSt8CweP)

[![asciicast](https://asciinema.org/a/418275.svg)](https://asciinema.org/a/418275)

Basic demo of the controllerbus daemon and ConfigSet format:

```sh
cd ./cmd/controllerbus
go build -v
./controllerbus daemon
```

This will load `controllerbus_daemon.yaml` and execute the boilerplate demo:

```
added directive                               directive="LoadControllerWithConfig<config-id=controllerbus/configset>"
added directive                               directive="ExecController<config-id=controllerbus/configset"
added directive                               directive="LoadConfigConstructorByID<config-id=controllerbus/example/boilerplate>"
starting controller                           controller=controllerbus/configset
added directive                               directive="ApplyConfigSet<controller-keys=boilerplate-example-0@1>"
added directive                               directive="LoadControllerWithConfig<config-id=controllerbus/bus/api>"
removed directive                             directive="LoadConfigConstructorByID<config-id=controllerbus/example/boilerplate>"
added directive                               directive="ExecController<config-id=controllerbus/bus/api>"
executing controller                          config-key=boilerplate-example-0 controller=controllerbus/configset
starting controller                           controller=controllerbus/bus/api
grpc api listening on: :5110                 
added directive                               directive="LoadControllerWithConfig<config-id=controllerbus/example/boilerplate>"
added directive                               directive="ExecController<config-id=controllerbus/example/boilerplate>"
starting controller                           controller=controllerbus/example/boilerplate
hello from boilerplate controller 1: hello world  controller=controllerbus/example/boilerplate
controller exited normally                    controller=controllerbus/example/boilerplate exec-dur="31.053µs"
```

### ConfigSet

**ConfigSet** is a key/value set of controller configurations to load.

The following is an example ConfigSet in YAML format for a program:

```yaml
example-1:
  # configuration object
  config:
    exampleField: "Hello world!"
  # ID of the configuration type
  id: controllerbus/example/boilerplate
  # rev # for overriding previous configs
  rev: 1
```

In this case, `example-1` is the ID of the controller. If multiple ConfigSet are
applied with the same ID, the latest rev wins. The ConfigSet controller
will automatically start and stop controllers as ConfigSets are changed.

### How does it work?

Controllers are executed by attaching them to a Bus. When attaching to a Bus,
all ongoing Directives are passed to the new Controller. The Controllers can
return Resolver objects to resolve result objects for Directives.

There are multiple ways to start a Controller:

 - `AddController`: with the Go API: construct & add the controller.
 - `AddDirective` -> `ExecControllerWithConfig`: with a directive.
 - yaml/json: resolving human-readable configuration to Config objects.

Config objects are Protobuf messages with attached validation functions. They
can be hand written in YAML and parsed to Protobuf or be created as Go objects.

### Protobuf Configuration

The [boilerplate](./example/boilerplate/controller/config.proto) example has the
following configuration proto:

```protobuf
// Config is the boilerplate configuration.
message Config {
  // ExampleField is an example configuration field.
  string example_field = 1;
}
```

This is an example YAML configuration for this controller:

```yaml
exampleField: "Hello world!"
```

With the Go API, we can use the **LoadControllerWithConfig** directive to
execute the controller with a configuration object:

```go
	bus.ExecOneOff(
		ctx,
		cb,
		resolver.NewLoadControllerWithConfig(&boilerplate_controller.Config{
			ExampleField: "Hello World!",
		}),
		nil,
	)
```

## Daemon and API

The [example daemon](./cmd/controllerbus) has an associated client and CLI for
the [Bus API](./bus/api), for example:

```sh
$ controllerbus client exec -f controllerbus_daemon.yaml
```

```json
  {
    "controllerInfo": {
      "version": "0.0.1",
      "id": "controllerbus/example/boilerplate"
    },
    "status": "ControllerStatus_RUNNING",
    "id": "boilerplate-example-0"
  }
```

The bus service has the following API:

```protobuf
// ControllerBusService is a generic controller bus lookup api.
service ControllerBusService {
  // GetBusInfo requests information about the controller bus.
  rpc GetBusInfo(GetBusInfoRequest) returns (GetBusInfoResponse) {}
  // ExecController executes a controller configuration on the bus.
  rpc ExecController(controller.exec.ExecControllerRequest) returns (stream controller.exec.ExecControllerResponse) {}
}
```

The RPC API is itself implemented as a controller, which can be configured:

```yaml
grpc-api:
  config:
    listenAddr: ":5000"
    busApiConfig:
      enableExecController: true
  id: controllerbus/bus/api
  rev: 1
```

For security, the default value of `enableExecController` is `false` to disallow
executing controllers via the API.

The structure under `cmd/controllerbus` and `example/boilerplate` are examples
which are intended to be copied to other projects, which reference the core
`controllerbus` controllers. A minimal program is as follows:

```go
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	b, sr, err := core.NewCoreBus(ctx, le)
	if err != nil {
		t.Fatal(err.Error())
	}
	sr.AddFactory(NewFactory(b))

	execDir := resolver.NewLoadControllerWithConfig(&Config{
		ExampleField: "testing",
	})
	_, ctrlRef, err := bus.ExecOneOff(ctx, b, execDir, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer ctrlRef.Release()
```

This provides logging, context cancelation. A single Factory is attached which
provides support for the Config type, (see the boilerplate example).

## Daemon and Client CLIs

Plugins can be bundled together with a set of root configurations into a CLI.
This can be used to bundle modules into a daemon and/or client for an
application - similar to the [controllerbus cli](./cmd/controllerbus).

## Testing

An in-memory Bus can be created for testing, an
[example](./example/boilerplate/controller/controller_test.go) is provided in
the boilerplate package.

## Plugins

[![asciicast](https://asciinema.org/a/I4LOCViLwzRlztYc1rytgAxWp.svg)](./example/plugin-demo)

**⚠ Plugins are experimental and not yet feature-complete.**

The [plugin](./plugin) system and compiler scans a set of Go packages for
ControllerBus factories and bundles them together into a hashed Plugin bundle.
The compiler CLI can watch code files for changes and re-build automatically.
Multiple plugin loaders and binary formats are supported.

```
USAGE:
   controllerbus hot compile - compile packages specified as arguments once

OPTIONS:
   --build-prefix value           prefix to prepend to import paths, generated on default [$CONTROLLER_BUS_PLUGIN_BUILD_PREFIX]
   --codegen-dir value            path to directory to create/use for codegen, if empty uses tmpdir [$CONTROLLER_BUS_CODEGEN_DIR]
   --output PATH, -o PATH         write the output plugin to PATH - accepts {buildHash} [$CONTROLLER_BUS_OUTPUT]
   --plugin-binary-id value       binary id for the output plugin [$CONTROLLER_BUS_PLUGIN_BINARY_ID]
   --plugin-binary-version value  binary version for the output plugin, accepts {buildHash} [$CONTROLLER_BUS_PLUGIN_BINARY_VERSION]
   --no-cleanup                   disable cleaning up the codegen dirs [$CONTROLLER_BUS_NO_CLEANUP]
   --help, -h                     show help
```

The CLI will analyze a list of Go package paths, discover all Factories
available in the packages, generate a Go module for importing all of the
factories into a single Plugin, and compile that package to a .so library.

## Projects

List of projects known to use Controller Bus:

 - [Bifrost]: networking and p2p library + daemon

[Bifrost]: https://github.com/aperturerobotics/bifrost

Open a PR to add your project to this list!

## Support

Please open a [GitHub issue] with any questions / issues.

[GitHub issue]: https://github.com/aperturerobotics/controllerbus/issues/new

... or feel free to reach out on [Matrix Chat] or [Discord].

[Discord]: https://discord.gg/ZAZSt8CweP
[Matrix Chat]: https://matrix.to/#/#aperturerobotics:matrix.org

## License

MIT
