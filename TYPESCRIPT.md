# ControllerBus in TypeScript

This document describes the TypeScript implementation of ControllerBus.

## Core Concepts

The core concepts of ControllerBus are the same in TypeScript as they are in Go:

- **Config**: configuration for a controller or process.
- **Controller**: a process which can create & handle Directives.
- **Directive**: a cross-controller request or declaration of target state.
- **Bus**: communication channel between running Controllers.
- **Factory**: constructor for controller and configuration objects.

## Implementation

The TypeScript implementation follows the same structure as the Go implementation, with `.ts` files placed next to their Go counterparts. For example:

- `controller/controller.go` → `controller/controller.ts`
- `directive/directive.go` → `directive/directive.ts`

## Key Differences from Go

1. **Contexts**: Go's context package is implemented as a TypeScript interface with similar semantics.
2. **Concurrency**: Instead of goroutines, the TypeScript implementation uses async/await and Promises.
3. **Error Handling**: Rather than returning multiple values with errors, TypeScript functions typically return tuples (arrays) or Promises that can reject.
4. **Type System**: TypeScript uses interfaces and type annotations instead of Go's struct types.
5. **Reference Counting**: TypeScript's garbage collection is automatic, but we explicitly implement reference counting for directives.

## Example Usage

Here's a simple example of creating and using a controller:

```typescript
import { background } from './core/core';
import { newBus } from './bus/inmem/inmem';
import { newController as newDirectiveController } from './directive/controller/controller';
import { newController as newLoaderController, waitExecControllerRunning, newExecController } from './controller/loader/loader';
import { MyControllerConfig } from './mycontroller/controller.pb';
import { newMyFactory } from './mycontroller/factory';

async function main() {
  const ctx = background();
  const logger = createLogger();

  // Create a directive controller
  const dc = newDirectiveController(ctx, logger);
  // Create a bus with the directive controller
  const b = newBus(dc);

  // Create a loader controller
  const [cl, err] = await newLoaderController(logger, b);
  if (err) {
    throw err;
  }

  // Execute the loader controller
  const [releaseCl, addErr] = await b.addController(ctx, cl);
  if (addErr) {
    throw addErr;
  }

  // Create a controller config
  const myConfig = new MyControllerConfig();
  myConfig.someField = 'some value';

  // Issue directive to run the controller
  const loadMyController = newExecController(newMyFactory(), myConfig);
  
  try {
    const [ctrl, instance, valRef, execErr] = await waitExecControllerRunning(ctx, b, loadMyController, null);
    if (execErr) {
      throw execErr;
    }

    // Now the controller is running
    
    // Clean up
    valRef.release();
  } catch (e) {
    console.error(e);
  }
}
```
