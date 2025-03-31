import { Context } from '../../core/core'
import {
  Handler,
  Directive,
  Instance,
  Reference,
  ReferenceHandler,
  Controller as DirectiveController,
  DirectiveWithEquiv,
  DirectiveWithSuperceeds,
  Resolver,
  ResolverHandler,
} from '../directive'
import { Info, newInfo } from '../../controller/controller'
import { Logger } from '../../core/core'

/**
 * Controller ID for the directive controller.
 */
export const ControllerID = 'controllerbus/directive/controller'

/**
 * DirectiveInstance implements the Instance interface.
 */
export class DirectiveInstance implements Instance {
  private readonly directive: Directive
  private readonly context: Context
  private readonly cancelContext: () => void
  private readonly ident: string
  private readonly logger: Logger

  private handlers: Map<ReferenceHandler, Reference> = new Map()
  private weakHandlers: Map<ReferenceHandler, Reference> = new Map()
  private disposeCallbacks: Set<() => void> = new Set()
  private idleCallbacks: Set<(isIdle: boolean, errs: Error[]) => void> =
    new Set()
  private resolverErrors: Error[] = []
  private idle: boolean = true
  private resolvers: Map<Resolver, ResolverHandler> = new Map()
  private disposed: boolean = false

  constructor(dir: Directive, logger: Logger) {
    this.directive = dir

    // Create a new cancelable context
    const [ctx, cancelCtx] = (globalThis as any).context.withCancel()
    this.context = ctx
    this.cancelContext = cancelCtx

    // Generate directive identifier
    this.ident = this.generateIdent(dir)
    this.logger = logger
  }

  /**
   * Generates a human-readable identifier for a directive.
   */
  private generateIdent(dir: Directive): string {
    let ident = dir.getName()

    // If the directive has debug values, append them
    if ('getDebugVals' in dir) {
      const debugDir = dir as any
      const debugVals = debugDir.getDebugVals()

      if (Object.keys(debugVals).length > 0) {
        const parts: string[] = []
        for (const [key, values] of Object.entries(debugVals)) {
          if (Array.isArray(values) && values.length > 0) {
            parts.push(`${key}=${values[0]}`)
          }
        }

        if (parts.length > 0) {
          ident += `<${parts.join(', ')}>`
        }
      }
    }

    return ident
  }

  /**
   * GetContext returns a context that is canceled when Instance is released.
   */
  getContext(): Context {
    return this.context
  }

  /**
   * GetDirective returns the underlying directive object.
   */
  getDirective(): Directive {
    return this.directive
  }

  /**
   * GetDirectiveIdent returns a human-readable string identifying the directive.
   */
  getDirectiveIdent(): string {
    return this.ident
  }

  /**
   * GetResolverErrors returns a snapshot of any errors returned by resolvers.
   */
  getResolverErrors(): Error[] {
    return [...this.resolverErrors]
  }

  /**
   * AddReference adds a reference to the directive.
   */
  addReference(cb: ReferenceHandler | null, weakRef: boolean): Reference {
    if (this.disposed) {
      // Create a no-op reference if already disposed
      return {
        release: () => {},
      }
    }

    const handler = cb || {
      handleValueAdded: () => {},
      handleValueRemoved: () => {},
      handleInstanceDisposed: () => {},
    }

    const ref: Reference = {
      release: () => {
        if (weakRef) {
          this.weakHandlers.delete(handler)
        } else {
          this.handlers.delete(handler)
        }

        // If no more references, start disposal process
        this.maybeDispose()
      },
    }

    if (weakRef) {
      this.weakHandlers.set(handler, ref)
    } else {
      this.handlers.set(handler, ref)
    }

    return ref
  }

  /**
   * AddDisposeCallback adds a callback that will be called when the instance
   * is disposed.
   */
  addDisposeCallback(cb: () => void): () => void {
    if (this.disposed) {
      // Call callback immediately if already disposed
      setTimeout(cb, 0)
      return () => {}
    }

    this.disposeCallbacks.add(cb)
    return () => {
      this.disposeCallbacks.delete(cb)
    }
  }

  /**
   * AddIdleCallback adds a callback that will be called when the idle state changes.
   */
  addIdleCallback(cb: (isIdle: boolean, errs: Error[]) => void): () => void {
    this.idleCallbacks.add(cb)

    // Call immediately with current state
    cb(this.idle, this.resolverErrors)

    return () => {
      this.idleCallbacks.delete(cb)
    }
  }

  /**
   * CloseIfUnreferenced cancels the directive instance if there are no refs.
   */
  closeIfUnreferenced(inclWeakRefs: boolean): boolean {
    if (this.handlers.size > 0) {
      return false
    }

    if (inclWeakRefs && this.weakHandlers.size > 0) {
      return false
    }

    this.close()
    return true
  }

  /**
   * Close cancels the directive instance and removes the directive.
   */
  close(): void {
    if (this.disposed) {
      return
    }

    this.disposed = true
    this.cancelContext()

    // Call dispose callbacks
    for (const cb of this.disposeCallbacks) {
      try {
        cb()
      } catch (e) {
        this.logger.error(`Error in dispose callback: ${e}`)
      }
    }

    // Notify handlers
    for (const handler of this.handlers.keys()) {
      try {
        handler.handleInstanceDisposed(this)
      } catch (e) {
        this.logger.error(`Error in handleInstanceDisposed: ${e}`)
      }
    }

    for (const handler of this.weakHandlers.keys()) {
      try {
        handler.handleInstanceDisposed(this)
      } catch (e) {
        this.logger.error(`Error in handleInstanceDisposed: ${e}`)
      }
    }

    // Clear all collections
    this.handlers.clear()
    this.weakHandlers.clear()
    this.disposeCallbacks.clear()
    this.idleCallbacks.clear()
    this.resolvers.clear()
  }

  /**
   * Check if the instance should be disposed due to no references.
   */
  private maybeDispose(): void {
    if (this.handlers.size === 0) {
      const opts = this.directive.getValueOptions()

      // If unrefDisposeEmptyImmediate is true, dispose immediately
      if (opts.unrefDisposeEmptyImmediate) {
        this.close()
        return
      }

      // Otherwise, schedule disposal after unrefDisposeDur
      if (opts.unrefDisposeDur > 0) {
        setTimeout(() => {
          // Recheck reference count when timer fires
          if (this.handlers.size === 0 && !this.disposed) {
            this.close()
          }
        }, opts.unrefDisposeDur)
      }
    }
  }

  /**
   * Update the idle state and notify callbacks if changed.
   */
  updateIdleState(isIdle: boolean): void {
    if (this.idle === isIdle) {
      return
    }

    this.idle = isIdle

    // Notify idle callbacks
    for (const cb of this.idleCallbacks) {
      try {
        cb(isIdle, this.resolverErrors)
      } catch (e) {
        this.logger.error(`Error in idle callback: ${e}`)
      }
    }
  }

  /**
   * Add a resolver error and update idle state.
   */
  addResolverError(err: Error): void {
    this.resolverErrors.push(err)
    this.updateIdleState(true)
  }
}

/**
 * DirectiveControllerImpl implements the DirectiveController interface.
 */
export class DirectiveControllerImpl implements DirectiveController {
  private readonly ctx: Context
  private readonly logger: Logger
  private readonly directives: Map<string, DirectiveInstance> = new Map()
  private readonly handlers: Set<Handler> = new Set()

  constructor(ctx: Context, logger: Logger) {
    this.ctx = ctx
    this.logger = logger
  }

  /**
   * GetDirectives returns a list of all currently executing directives.
   */
  getDirectives(): Instance[] {
    return Array.from(this.directives.values())
  }

  /**
   * AddDirective adds a directive to the controller.
   */
  async addDirective(
    dir: Directive,
    handler: ReferenceHandler | null,
  ): Promise<[Instance, Reference, Error | null]> {
    // Validate the directive
    const validationErr = dir.validate()
    if (validationErr) {
      return [null as any, null as any, validationErr]
    }

    // Check for equivalent directives
    if ('isEquivalent' in dir) {
      const equivDir = dir as DirectiveWithEquiv

      for (const [key, instance] of this.directives.entries()) {
        const existingDir = instance.getDirective()

        if (equivDir.isEquivalent(existingDir)) {
          // Check if the new directive superceeds the existing one
          if ('superceeds' in dir) {
            const supDir = dir as DirectiveWithSuperceeds
            if (supDir.superceeds(existingDir)) {
              // Replace the existing directive
              instance.close()
              this.directives.delete(key)
              break
            }
          }

          // Use the existing directive
          const ref = instance.addReference(handler, false)
          return [instance, ref, null]
        }
      }
    }

    // Create a new directive instance
    const instance = new DirectiveInstance(dir, this.logger)
    this.directives.set(instance.getDirectiveIdent(), instance)

    // Add handler reference
    const ref = instance.addReference(handler, false)

    // Set up to clean up when context is done
    instance
      .getContext()
      .done()
      .then(() => {
        this.directives.delete(instance.getDirectiveIdent())
      })

    // Inform handlers of the new directive
    for (const handler of this.handlers) {
      try {
        await handler.handleDirective(instance.getContext(), instance)
      } catch (e) {
        this.logger.error(`Error in handleDirective: ${e}`)
      }
    }

    return [instance, ref, null]
  }

  /**
   * AddHandler adds a directive handler.
   */
  async addHandler(handler: Handler): Promise<[() => void, Error | null]> {
    if (this.handlers.has(handler)) {
      return [() => {}, new Error('handler already added')]
    }

    this.handlers.add(handler)

    // Handle all existing directives
    for (const instance of this.directives.values()) {
      try {
        await handler.handleDirective(instance.getContext(), instance)
      } catch (e) {
        this.logger.error(
          `Error in handleDirective during initial handling: ${e}`,
        )
      }
    }

    return [
      () => {
        this.handlers.delete(handler)
      },
      null,
    ]
  }
}

/**
 * Creates a new directive controller.
 *
 * @param ctx - The parent context
 * @param logger - The logger to use
 * @returns A new directive controller
 */
export function newController(
  ctx: Context,
  logger: Logger,
): DirectiveController {
  return new DirectiveControllerImpl(ctx, logger)
}
