# Controller Bus

> Declarative configuration for concurrently executing controllers.

## Introduction

[![asciicast](https://asciinema.org/a/418275.svg)](https://asciinema.org/a/418275)

Controller Bus is a framework for declarative configuration, dynamic linking,
and separation of concerns between application components. 

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
added directive                               directive="LoadControllerWithConfig<config-id=controllerbus/hot/loader/filesystem/1>"
added directive                               directive="LoadControllerWithConfig<config-id=controllerbus/configset/1>"
added directive                               directive="ExecController<config-id=controllerbus/configset/1>"
added directive                               directive="ExecController<config-id=controllerbus/hot/loader/filesystem/1>"
added directive                               directive="LoadConfigConstructorByID<config-id=controllerbus/example/boilerplate/1>"
starting controller                           controller=controllerbus/configset/1
added directive                               directive="ApplyConfigSet<controller-keys=boilerplate-example-0@1>"
added directive                               directive="LoadControllerWithConfig<config-id=controllerbus/bus/api/1>"
starting controller                           controller=controllerbus/hot/loader/filesystem/1
removed directive                             directive="LoadConfigConstructorByID<config-id=controllerbus/example/boilerplate/1>"
added directive                               directive="ExecController<config-id=controllerbus/bus/api/1>"
executing controller                          config-key=boilerplate-example-0 controller=controllerbus/configset/1
added directive                               directive="LoadControllerWithConfig<config-id=controllerbus/example/boilerplate/1>"
added directive                               directive="ExecController<config-id=controllerbus/example/boilerplate/1>"
starting controller                           controller=controllerbus/bus/api/1
grpc api listening on: :5110                 
starting controller                           controller=controllerbus/example/boilerplate/1
hello from boilerplate controller 1: hello world  controller=controllerbus/example/boilerplate/1
controller exited normally                    controller=controllerbus/example/boilerplate/1 exec-time="136.268µs"
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
