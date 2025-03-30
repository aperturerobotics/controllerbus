# Controller Bus Design

This document contains design notes and details for the Controller Bus system.

## Introduction

ControllerBus is a Go framework designed for building modular, decoupled, and state-driven applications. It facilitates communication and coordination between independent components called "Controllers" through a central "Bus" using declarative "Directives". This design promotes loose coupling, allowing components to request desired states or data without needing direct knowledge of which other components will fulfill those requests.

## Core Components

The primary components of ControllerBus are:

*   **Bus**: The central message broker and coordinator (`bus/bus.go`). It manages controllers, directives, and resolvers, orchestrating their lifecycles and communication.
*   **Controller**: A state machine or long-running process (`controller/controller.go`) that interacts with the bus. Controllers handle directives and can yield resolvers to fulfill them. A base implementation `BusController` (`bus/controller.go`) simplifies common patterns.
*   **Factory**: Contains metadata and constructors for a specific controller implementation (`controller/factory.go`). It defines the controller's ID, version, and how to create its configuration and instances. A base implementation `BusFactory` (`bus/factory.go`) is provided.
*   **Config**: An object, often a Protobuf message (`controller/controller.pb.go`), that holds the configuration parameters for a specific controller instance at construction time.
*   **Directive**: An ongoing request for data or a desired state (`directive/directive.go`, `directive/directive.pb.go`). Directives are identified by their type and parameters and serve as the primary means of inter-controller communication.
*   **Resolver**: A concurrent process, typically a goroutine, yielded by a controller to compute values that satisfy a specific directive instance (`directive/resolver.go`).
*   **Value**: An opaque object yielded by a Resolver and attached to a Directive instance (`directive/directive.go`). Values represent the results fulfilling the directive's request.

## Directives In-Depth

Directives are the heart of the ControllerBus communication model.

### Purpose and Function

Directives represent a request for a specific piece of data or a desired system state. They are added to the bus by controllers or external actors. The bus then dispatches the directive to all running controllers to see which ones can handle it.

### Specialized Directive Interfaces

Beyond the basic `directive.Directive` interface (`directive/directive.go`), several specialized interfaces enhance directive behavior:

*   **`DirectiveWithEquiv`**: Allows the bus to de-duplicate directives. If a new directive `IsEquivalent` to an existing one, the bus may reuse the existing `directive.Instance` instead of creating a new one, incrementing its reference count.
*   **`DirectiveWithSuperceeds`**: Used with `DirectiveWithEquiv`. If a new directive `IsEquivalent` and also `Superceeds` an existing one (e.g., a newer version), the existing directive instance may be canceled and replaced by the new one.

### Directive Options

The `directive.ValueOptions` struct (`directive/directive.go`) allows directives to specify how their values should be handled:

*   `MaxValueCount`: Limits the number of values the bus should collect. Resolvers might be stopped once this limit is reached.
*   `MaxValueHardCap`: Determines if `MaxValueCount` is a strict limit. If false, values arriving after the limit is reached might still be accepted.
*   `UnrefDisposeDur`: A delay before disposing of an unreferenced directive instance.
*   `UnrefDisposeEmptyImmediate`: If true, unreferenced directives with no values are disposed of immediately, ignoring the duration.

### Bounded vs. Unbounded Directives

*   **Bounded**: Directives with `MaxValueCount > 0`. The bus aims to collect this specific number of values.
*   **Unbounded**: Directives with `MaxValueCount == 0`. The bus collects values indefinitely until the directive is canceled.

### Values and Attached Values

*   **`Value`**: The raw data or object returned by a resolver (`directive/directive.go`). It's an empty interface (`interface{}`).
*   **`AttachedValue`**: A `Value` associated with a directive instance, augmented with a unique `ValueID` assigned by the bus (`directive/directive.go`).
*   **`TypedAttachedValue[T]`**: A generic version of `AttachedValue` for type safety (`directive/directive.go`).
*   **`TransformedAttachedValue[T, E]`**: An `AttachedValue` that also holds a transformed version of the original value (`directive/directive.go`).
*   **`ProtoDebugValue`**: A Protobuf representation of a debug key-value pair (`directive/directive.pb.go`).

### Directive References and Handlers

When a directive is added via `AddDirective`, a `directive.Reference` is returned (`directive/directive.go`).

*   **`Reference`**: Represents a caller's interest in a directive. Must be held and eventually `Release()`'d.
*   **`ReferenceHandler`**: An interface implemented by callers to receive notifications about the directive's values (`directive/directive.go`). Key methods:
    *   `HandleValueAdded`: Called when a new value is available.
    *   `HandleValueRemoved`: Called when a value is no longer valid.
    *   `HandleInstanceDisposed`: Called when the directive instance is fully removed.

### Directive Lifecycle (Reference Counting)

Directives are reference-counted. `AddDirective` creates or finds an existing `directive.Instance` and increments its reference count. `Reference.Release()` decrements the count. When the count reaches zero (and potentially after the `UnrefDisposeDur`), the bus cancels the `directive.Instance` and its associated resolvers, calling `HandleInstanceDisposed` on any remaining handlers.

### Networked Directives

The `Networked` and `NetworkedCodec` interfaces (`directive/directive.go`) define how directives can be serialized and deserialized, enabling their use across process boundaries or networks.

### Directive Debugging

*   **`Debuggable`**: An optional interface for directives to provide key-value pairs for debugging (`directive/directive.go`).
*   **`DebugValues`**: The map type returned by `GetDebugVals()` (`directive/directive.go`).
*   **`DirectiveInfo` / `DirectiveState`**: Protobuf messages (`directive/directive.pb.go`) used to represent directive information and state, often including debug values.

## Controllers In-Depth

Controllers are the active components within the ControllerBus system.

### Interface (`controller/controller.go`)

All controllers must implement the `controller.Controller` interface.

### Key Methods

*   **`GetControllerInfo()`**: Returns static metadata about the controller (`*controller.Info`).
*   **`HandleDirective(ctx context.Context, di directive.Instance)`**: Called by the bus for each active directive instance. The controller inspects the directive (`di.GetDirective()`) and returns a slice of `directive.Resolver` implementations if it can handle the directive, or `nil` otherwise. The context is canceled when the directive instance expires.
*   **`Execute(ctx context.Context)`**: The main entry point for the controller's long-running logic. It receives a context that is canceled when the controller should shut down. Returning `nil` indicates successful completion. Returning an error triggers a retry mechanism by the bus (without reconstructing the controller).
*   **`Close()`**: Called when the controller is permanently shutting down, allowing it to release resources.

### Base Implementation (`bus/controller.go`)

`bus.BusController[T]` provides a convenient base struct for controllers. It handles common boilerplate like storing the logger, bus instance, configuration, and controller metadata. It provides default implementations for the `Controller` interface methods, which can be overridden as needed. `NewBusControllerFactory` simplifies creating factories for controllers based on `BusController`.

### Controller Info (`controller/info.go`, `controller/controller.pb.go`)

The `controller.Info` struct (and its Protobuf representation in `controller.pb.go`) holds static information about a controller:

*   `Id`: A unique string identifier.
*   `Version`: The controller's version (SemVer string).
*   `Description`: A human-readable description.
The `controller.NewInfo` helper function (`controller/info.go`) simplifies creating these objects.

## Factories In-Depth

Factories are responsible for creating controller instances.

### Interface (`controller/factory.go`)

All factories must implement the `controller.Factory` interface.

### Purpose

Factories decouple the *request* to run a controller from its *implementation*. They provide:

*   **Metadata**: The controller's unique configuration ID (`GetConfigID`) and version (`GetVersion`).
*   **Configuration Handling**: A way to construct a default configuration object (`ConstructConfig`).
*   **Instantiation**: The core `Construct` method, which builds a controller instance given a configuration object (`config.Config`) and `ConstructOpts` (containing the logger).

### Base Implementation (`bus/factory.go`)

`bus.BusFactory[T, C]` provides a generic factory implementation, suitable for controllers built using `bus.BusController`. It takes the bus, config/controller IDs, version, and constructor functions as parameters.

### Factory Resolver (`controller/factory.go`)

The `FactoryResolver` interface defines a mechanism for locating factories, often used by loader controllers. Implementations might find factories statically, via plugins, etc.

## Resolvers In-Depth

Resolvers perform the actual work of fulfilling directives by producing values.

### Interface (`directive/resolver.go`)

All resolvers must implement the `directive.Resolver` interface.

### Key Methods

*   **`Resolve(ctx context.Context, handler directive.ResolverHandler)`**: Called by the bus to start or restart the resolver's work for a specific directive instance activation.
    *   `ctx`: Canceled when this specific activation should stop (e.g., directive canceled, value limit reached).
    *   `handler`: The callback interface used by the resolver to yield values (`AddValue`), manage them (`RemoveValue`, `ClearValues`), signal idleness (`MarkIdle`), and manage lifecycle callbacks (`AddValueRemovedCallback`, etc.).
    *   Return Value: `nil` indicates the resolver finished its current task (might be called again). An error indicates a fatal failure for this resolver.

### Resolver Lifecycle

Resolvers are managed by the bus. They can be started, stopped (context canceled), and restarted multiple times during a directive instance's lifetime, depending on factors like value expiration or changes in the number of required values.

### Resolver Handlers (`directive/directive.go`)

The `handler` passed to `Resolve` implements `directive.ResolverHandler`, which embeds `directive.ValueHandler`:

*   **`ValueHandler`**: Methods for managing values (`AddValue`, `RemoveValue`, `CountValues`, `ClearValues`).
*   **`ResolverHandler`**: Adds methods for managing resolver state and lifecycle:
    *   `MarkIdle(bool)`: Signals whether the resolver is currently active or idle.
    *   `AddValueRemovedCallback(uint32, func())`: Registers a callback for when a specific value is removed.
    *   `AddResolverRemovedCallback(func())`: Registers a callback for when this resolver instance is removed.
    *   `AddResolver(Resolver, func())`: Allows nesting resolvers.

### Common Resolver Implementations

ControllerBus provides several pre-built resolver types for common patterns:

*   **`FuncResolver` (`directive/func-resolver.go`)**: Wraps a simple function `func(ctx, handler)` as a resolver.
*   **`ValueResolver` (`directive/value-resolver.go`)**: Resolves with a static list of pre-defined values.
*   **`GetterResolver` (`directive/getter-resolver.go`)**: Resolves by calling a getter function `func(ctx) (T, error)` once.
*   **`KeyedGetterResolver` (`directive/keyed-getter-resolver.go`)**: Resolves using a keyed getter function `func(ctx, key, releaseCb) (V, releaseVal, error)`, potentially re-fetching if the value is released externally.
*   **`AccessResolver` (`directive/access-resolver.go`)**: Resolves using an access function `func(ctx, releasedCb) (T, releaseVal, error)`, similar to `KeyedGetterResolver` but without the key, designed for resources needing explicit acquire/release.
*   **`WatchableResolver` (`directive/watchable-resolver.go`, `_test.go`)**: Wraps a `ccontainer.Watchable` container. It watches the container and adds/removes values from the handler as the container's value changes.
*   **`WatchableTransformResolver` (`directive/watchable-transform-resolver.go`, `_test.go`)**: Similar to `WatchableResolver`, but applies a transformation function `TransformValueFunc` to the value before adding it.
*   **`RefCountResolver` (`directive/refcount-resolver.go`)**: Resolves by accessing a `refcount.RefCount` container. It acquires a reference, gets the value, adds it to the handler, marks idle, and waits for the context to be canceled before releasing the reference. Can optionally transform the value.
*   **`KeyedRefCountResolver` (`directive/keyed-refcount-resolver.go`)**: Similar to `RefCountResolver` but operates on a `keyed.KeyedRefCount` using a specific key.
*   **`RetryResolver` (`directive/resolver-retry.go`)**: Wraps another resolver, adding retry logic with backoff if the inner resolver returns an error.
*   **`TransformResolver` (`directive/transform-resolver.go`)**: Resolves by adding a *different* directive to the bus and applying a transformation function (`TransformAttachedValueFunc`) to the resulting `AttachedValue`s before yielding them.
*   **`TypedTransformResolver` (`directive/typed-transform-resolver.go`)**: A type-safe version of `TransformResolver` using `TypedAttachedValue` and `TransformTypedAttachedValueFunc`.
*   **`UniqueListResolver` / `UniqueListXfrmResolver` (`directive/unique-list-resolver.go`)**: Manages a list of unique values identified by a key function, adding/removing/updating values in the handler. `Xfrm` version applies a transform. Uses `unique.KeyedList`.
*   **`UniqueMapResolver` / `UniqueMapXfrmResolver` (`directive/unique-map-resolver.go`)**: Manages a map of unique values identified by their key, adding/removing/updating values in the handler. `Xfrm` version applies a transform. Uses `unique.KeyedMap`.

## Bus In-Depth

The Bus is the central nervous system of the application.

### Interface (`bus/bus.go`)

The `bus.Bus` interface combines `directive.Controller` (which includes `DirectiveAdder`, `DirectiveLister`, `HandlerAdder`) with methods for managing controller lifecycles (`AddController`, `ExecuteController`, `RemoveController`, `GetControllers`).

### Responsibilities

*   **Controller Registry**: Tracking active controllers.
*   **Directive Dispatch**: Forwarding directives to controllers (`HandleDirective`).
*   **Resolver Management**: Managing the lifecycle of resolvers (`Resolve`).
*   **Value Aggregation**: Collecting values from resolvers and managing `directive.Instance` state.
*   **Lifecycle Orchestration**: Starting (`Execute`) and stopping (`Close`) controllers.
*   **Reference Counting**: Managing directive instance lifecycles based on `Reference` counts.

### Bus API Helpers

The `bus` package provides high-level functions simplifying common interactions:

*   **`ExecOneOff*` (`bus/oneoff.go`)**: Execute a directive, wait for one value (optionally filtered/typed/transformed), handle idle states, and manage the reference. Ideal for request-response patterns. `ExecIdleCallback` types (`WaitWhenIdle`, `ReturnWhenIdle`) customize behavior when the directive has no active resolvers.
*   **`ExecWaitValue` (`bus/wait.go`)**: Similar to `ExecOneOff`, but specifically waits for a value matching a type `T` and a check callback `checkCb`.
*   **`ExecCollectValues` (`bus/multi.go`)**: Executes a directive and collects multiple values of type `T` until the context is done or the directive becomes idle.
*   **`ExecWatch*` (`bus/watch.go`)**: Execute a directive and continuously monitor its values, updating callbacks, channels, containers (`ccontainer`), or routines (`routine.StateRoutineContainer`) as the resolved value(s) change. Includes variants for selecting specific values (`ExecOneOffWatchSelectCb`), applying effects (`ExecWatchEffect`), and transforming values before applying effects (`ExecWatchTransformEffect`).
*   **`New*RefCount*` (`bus/refcount.go`)**: Create `refcount.RefCount` containers that automatically manage a one-off directive on the bus. The directive is active only when the container has references.
*   **`WaitExecControllerRunning*` (`controller/loader/ex-load-controller.go`)**: Executes an `ExecController` directive and waits specifically until the controller reaches a running state (or errors out), returning the `controller.Controller` instance.

### Reference Handlers

Implementations of `directive.ReferenceHandler` process value updates for `AddDirective`:

*   **`CallbackHandler` (`directive/callback-handler.go`, `bus/callback.go`)**: Uses simple function callbacks (`valCb`, `removedCb`, `disposeCb`).
*   **`TypedCallbackHandler` (`directive/typed-callback-handler.go`)**: A generic, type-safe version of `CallbackHandler`. Can delegate handling of unknown types.
*   **`PassThruHandler` (`bus/passthru.go`)**: Forwards value additions/removals directly to a `directive.ResolverHandler`. Useful for composing resolvers.
*   **`TransformHandler` (`bus/transform.go`)**: Similar to `PassThruHandler`, but applies a transformation function to values before forwarding them.
*   **`LogHandler` (`directive/log-handler.go`)**: Wraps another handler (optional) and logs value addition/removal/disposal events.

### Idle Callbacks

Functions conforming to `directive.IdleCallback` are used with `directive.Instance.AddIdleCallback` and helpers like `ExecOneOff` to react to changes in a directive's idle state (i.e., whether it has active resolvers).

*   **`NewErrChIdleCallback` (`directive/err-idle-callback.go`)**: An `IdleCallback` implementation that pushes resolver errors (other than context cancellation) to a channel.

## Controller Loading

ControllerBus supports dynamic loading of controllers, typically managed by a dedicated loader controller.

*   **`ExecController` Directive (`controller/loader/directive.go`)**: The standard directive to request loading and execution of a controller. It specifies the `controller.Factory` and `config.Config` to use, along with optional retry backoff settings.
*   **`ExecControllerValue` (`controller/loader/dir-exec-controller-value.go`)**: The value yielded by resolvers handling `ExecController`. It contains the `controller.Controller` instance (if running), any execution error, and timestamps for updates and potential retries.
*   **Loader Controller (`controller/loader/loader.go`)**: An implementation (`loader.Controller`) that handles `ExecController` directives. It uses the bus to construct and execute the requested controller.
*   **Loader Resolver (`controller/loader/loader-resolve.go`)**: The internal resolver used by `loader.Controller`. It manages the construction, execution (`bus.ExecuteController`), and retry logic (using `cbackoff`) for the target controller, yielding `ExecControllerValue` updates.

## Lifecycle and Error Handling

Controllers and directives have well-defined lifecycles.

### Controller Lifecycle Steps

1.  **Construction**: Via `Factory.Construct`.
2.  **Addition to Bus**: Via `Bus.AddController` or `Bus.ExecuteController`. `HandleDirective` called for existing directives.
3.  **Execution**: `Controller.Execute` called by the bus in a goroutine.
4.  **Running**: Controller processes directives (`HandleDirective`) and executes its logic. Resolvers are managed by the bus.
5.  **Shutdown Signal**: Context passed to `Execute` is canceled.
6.  **Execution Exit**: `Execute` returns. `nil` for clean exit, `error` for retry.
7.  **Closing**: `Controller.Close` called by the bus.
8.  **Removal from Bus**: Controller removed from the bus registry.

### Error Handling

*   Errors from `Controller.Execute` trigger retries with backoff.
*   Errors from `Resolver.Resolve` mark that resolver as failed; it won't be retried automatically by the core bus (though `RetryResolver` can be used). Resolver errors are reported via `IdleCallback`.
*   `directive.ErrDirectiveDisposed` (`directive/errors.go`) is a standard error indicating a directive was unexpectedly terminated.

## Concurrency Model

ControllerBus is inherently concurrent:

*   Each **Controller**'s `Execute` method runs in its own goroutine.
*   **`HandleDirective`** calls are synchronous from the bus's perspective, but controllers should avoid blocking.
*   **Resolvers** run concurrently, managed by the bus, often spawning their own goroutines or sub-resolvers.
*   The **Bus** uses internal locking (e.g., mutexes, broadcast variables) to manage shared state safely across multiple goroutines involved in handling directives, controllers, and resolvers.

