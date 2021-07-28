# Controller Bus

> Modular Go application framework with concurrent control loops.

## Introduction

[![asciicast](https://asciinema.org/a/418275.svg)](https://asciinema.org/a/418275)

ControllerBus is a framework for modular Go applications.

Applications are built with concurrently executing Controllers that communicate
over a shared bus (either in-memory or **networked** with IPC) using Directive
requests. The Directives can be deduplicated and their outputs cached to
optimize multiple controllers requesting the same thing simultaneously.

Config objects are Protobuf messages with attached validation functions and
controller IDs. On startup, a configuration object is passed to a Factory which
constructs the associated Controller, which is then attached to the Bus. When
attaching, all ongoing Directives are passed to the Controller, which can
optionally return a Resolver object to concurrently fetch results.

Controllers can be attached and detached on-demand. There is an associated
example daemon and GRPC API for remotely starting Controllers over a network.

Concurrent execution of communicating components allows for improved
multi-threading and faster startup time. Decoupling the implementations of the
components from the API surfaces makes it trivial to swap-in new
implementations, even without restarting the program.

## Example

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

Using the **LoadControllerWithConfig** directive, we can instruct the system to
resolve the configuration type to a controller factory & exec the controller:

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

You can also run this demo by:

```sh
cd ./cmd/controllerbus
go build -v 
./controllerbus daemon
```

This will load `controllerbus_daemon.yaml` and execute the boilerplate demo:

```
added directive                               directive="LoadControllerWithConfig<config-id=controllerbus/configset/1>"
added directive                               directive="ExecController<config-id=controllerbus/configset/1>"
added directive                               directive="LoadConfigConstructorByID<config-id=controllerbus/example/boilerplate/1>"
starting controller                           controller=controllerbus/configset/1
added directive                               directive="ApplyConfigSet<controller-keys=boilerplate-example-0@1>"
added directive                               directive="LoadControllerWithConfig<config-id=controllerbus/bus/api/1>"
removed directive                             directive="LoadConfigConstructorByID<config-id=controllerbus/example/boilerplate/1>"
added directive                               directive="ExecController<config-id=controllerbus/bus/api/1>"
executing controller                          config-key=boilerplate-example-0 controller=controllerbus/configset/1
starting controller                           controller=controllerbus/bus/api/1
grpc api listening on: :5110                 
added directive                               directive="LoadControllerWithConfig<config-id=controllerbus/example/boilerplate/1>"
added directive                               directive="ExecController<config-id=controllerbus/example/boilerplate/1>"
starting controller                           controller=controllerbus/example/boilerplate/1
hello from boilerplate controller 1: hello world  controller=controllerbus/example/boilerplate/1
controller exited normally                    controller=controllerbus/example/boilerplate/1 exec-time="31.053µs"
```

### ConfigSet

**ConfigSet** is a key/value set of controller configurations to load.

The following is an example ConfigSet in YAML format for a program:

```yaml
example-1:
  # configuration object
  config:
    exampleField: "Hello world 1!"
  # ID of the configuration type
  id: controllerbus/example/boilerplate/1
  # revision # for overriding previous configs
  revision: 1
```

In this case, `example-1` is the ID of the controller. If multiple ConfigSet are
applied with the same ID, the latest revision wins. The ConfigSet controller
will automatically start and stop controllers as ConfigSets are changed.

## Overview

The primary components of controller bus are:

 - **Config**: an object that configures a controller at construct time.
 - **Controller**: state machine / goroutine processing Directives on a bus.
 - **Bus**: a channel to connect together multiple Controllers.
 - **Factory**: contains controller implementation metadata and constructors.
 - **Directive**: an ongoing request for data or desired state.
 - **Resolver**: concurrent process(es) computing values to satisfy a directive.

Controllers are started attached to a common Bus. They can be directly attached
or loaded with directives to the "loading controller." A directive to load and
start a controller might be resolved by fetching code from the network and
loading a dynamic library, for example. Controllers have a single entrypoint
Goroutine but can spawn other routines as needed.

Directive objects can be attached to a Bus, where they are passed to all running
controllers for handling. Directives are de-duplicated, and reference counting
is used to determine when a directive can be canceled and released.

The controllerbus system manages starting and stopping resolvers yielded by the
controller handlers. A resolver executes until the directive has the desired
number of values, or the directive is canceled. Resolvers can be started and
stopped multiple times in the life-span of a directive.

A "Value" is an opaque object attached to a Directive, which will ultimately be
returned to the originator of the Directive. Bounded directives accept a limited
number of values before canceling remaining resolvers. Values can be expired,
and if the desired value count drops below a threshold, the resolvers will be
restarted until new values are found. A bounded directive with a value limit of
1 is sometimes referred to as a "singleton" in this document and the codebase.

The controller model is similar to the microservices model:

 - Declare a contract for a component as an API (Rest, gRPC)
 - Other components link against the client for that API
 - Communication between components occurs in-process over network.
 - Subroutines concurrently process requests (distributed model).

The goal of this project is to find a happy medium between the two approaches,
supporting statically linked, dynamically linked (plugin), or networked
(distributed) controller implementations and execution models. In practice, it
declares a common format for controller configuration, construction, and
execution in Go projects.

## Daemon and API

The [example daemon](./cmd/controllerbus) is an associated client and CLI for
the [Bus GRPC API](./bus/api), for example:

```sh
$ controllerbus client exec -f controllerbus_daemon.yaml 
```

```json
  {
    "controllerInfo": {
      "version": "0.0.1",
      "id": "controllerbus/example/boilerplate/1"
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

The GRPC API is itself implemented as a controller, which can be configured:

```yaml
grpc-api:
  config:
    listenAddr: ":5000"
    busApiConfig:
      enableExecController: true
  id: controllerbus/bus/api/1
  revision: 1
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

## Plugins

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

## Daemon and Client CLIs

Plugins can be bundled together with a set of root configurations into a CLI.
This can be used to bundle modules into a daemon and/or client for an
application - similar to the [controllerbus cli](./cmd/controllerbus).

## Testing

An in-memory Bus can be created for testing, an
[example](./example/boilerplate/controller/controller_test.go) is provided in
the boilerplate package.
